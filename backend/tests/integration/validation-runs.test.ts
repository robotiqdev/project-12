/**
 * Integration tests for TASK-4588 — ValidationRun API endpoints with filtering
 * and ValidationRunDetail API.
 *
 * Tests the following API behaviors:
 *   1. GET /api/validation-runs?status=CANCELLED returns only cancelled runs.
 *   2. GET /api/validation-runs?commitSha=3a2b returns runs whose SHA starts with '3a2b'.
 *   3. GET /api/validation-runs?branch=main&status=SUCCESS combined filter works.
 *   4. GET /api/validation-runs/:id includes a `stages` array with all 5 stages,
 *      each having stageName, stageIndex, status, logs, startedAt, completedAt.
 *   5. Stages in the detail response are ordered by stageIndex ascending.
 *   6. GET /api/validation-runs/:non-existent-id returns 404.
 *   7. GET /api/validation-runs?limit=500 returns at most 100 runs (max cap).
 *
 * Prerequisites:
 *   - DATABASE_URL  env var pointing to a running PostgreSQL instance with the
 *                  schema applied.
 *   - API_BASE_URL  env var pointing to the running API server
 *                  (default: http://localhost:3000).
 */

import { Client } from "pg";
import { randomUUID } from "crypto";

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

const API_BASE_URL = process.env.API_BASE_URL ?? "http://localhost:3000";

const STAGE_TYPES = ["SETUP", "LINT", "TEST", "BUILD", "DEPLOY"] as const;
const ALL_STATUSES = [
  "PENDING",
  "RUNNING",
  "SUCCESS",
  "FAILED",
  "CANCELLED",
] as const;

type ValidationRunStatus = (typeof ALL_STATUSES)[number];
type StageType = (typeof STAGE_TYPES)[number];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/**
 * Make a GET request to the API and return the parsed JSON body alongside the
 * HTTP status code.
 */
async function apiGet(
  path: string
): Promise<{ status: number; body: unknown }> {
  const url = `${API_BASE_URL}${path}`;
  const res = await fetch(url, {
    headers: { "Content-Type": "application/json" },
  });
  let body: unknown;
  try {
    body = await res.json();
  } catch {
    body = null;
  }
  return { status: res.status, body };
}

/** Seed a single ValidationRun and return its id. */
async function seedRun(
  client: Client,
  opts: {
    branch: string;
    commitSha: string;
    status: ValidationRunStatus;
    userId?: string;
    withAllStages?: boolean;
  }
): Promise<string> {
  const runId = randomUUID();
  const userId = opts.userId ?? "api_test_user";

  await client.query(
    `INSERT INTO validation_runs
       (id, "branchName", "commitSha", status, config, "userId", "createdAt", "updatedAt")
     VALUES ($1, $2, $3, $4::"ValidationRunStatus", $5, $6, NOW(), NOW())`,
    [
      runId,
      opts.branch,
      opts.commitSha,
      opts.status,
      JSON.stringify({ timeout: 300 }),
      userId,
    ]
  );

  if (opts.withAllStages) {
    // Insert all 5 stages in shuffled order to test ordering
    const shuffled = [...STAGE_TYPES].reverse(); // DEPLOY, BUILD, TEST, LINT, SETUP
    for (const [i, stageName] of shuffled.entries()) {
      const originalIndex = STAGE_TYPES.indexOf(stageName);
      await client.query(
        `INSERT INTO run_stages
           (id, "runId", "stageName", "stageIndex", status, logs, "startedAt", "completedAt")
         VALUES ($1, $2, $3::"StageType", $4, 'SUCCESS'::"StageStatus",
                 $5, NOW() - INTERVAL '10 minutes', NOW() - INTERVAL '5 minutes')`,
        [
          randomUUID(),
          runId,
          stageName,
          originalIndex, // stageIndex matches the canonical STAGE_TYPES order
          `Log output for stage ${stageName}`,
        ]
      );
    }
  }

  return runId;
}

// ---------------------------------------------------------------------------
// Test suite
// ---------------------------------------------------------------------------

describe("ValidationRun API — filtering and detail (TASK-4588)", () => {
  let client: Client;

  // IDs of runs seeded by this suite — used for cleanup
  const seededRunIds: string[] = [];

  // -------------------------------------------------------------------------
  // Setup / teardown
  // -------------------------------------------------------------------------

  beforeAll(async () => {
    client = new Client({ connectionString: process.env.DATABASE_URL });
    await client.connect();
  });

  afterAll(async () => {
    if (seededRunIds.length > 0) {
      // CASCADE deletes also remove associated run_stages
      await client.query(
        `DELETE FROM validation_runs WHERE id = ANY($1::uuid[])`,
        [seededRunIds]
      );
    }
    await client.end();
  });

  // -------------------------------------------------------------------------
  // Test 1: filter by status=CANCELLED
  // -------------------------------------------------------------------------

  describe("GET /api/validation-runs?status=CANCELLED", () => {
    let cancelledId: string;
    let pendingId: string;

    beforeAll(async () => {
      cancelledId = await seedRun(client, {
        branch: "filter-test-branch",
        commitSha: `abc000${randomUUID().replace(/-/g, "").slice(0, 10)}`,
        status: "CANCELLED",
        userId: "api_test_user",
      });
      pendingId = await seedRun(client, {
        branch: "filter-test-branch",
        commitSha: `def000${randomUUID().replace(/-/g, "").slice(0, 10)}`,
        status: "PENDING",
        userId: "api_test_user",
      });
      seededRunIds.push(cancelledId, pendingId);
    });

    it("returns HTTP 200", async () => {
      const { status } = await apiGet("/api/validation-runs?status=CANCELLED");
      expect(status).toBe(200);
    });

    it("returns only CANCELLED runs in the response body", async () => {
      const { body } = await apiGet("/api/validation-runs?status=CANCELLED");
      const runs = (body as { data: Array<{ id: string; status: string }> }).data;
      expect(Array.isArray(runs)).toBe(true);
      expect(runs.every((r) => r.status === "CANCELLED")).toBe(true);
    });

    it("does not include non-CANCELLED runs", async () => {
      const { body } = await apiGet("/api/validation-runs?status=CANCELLED");
      const runs = (body as { data: Array<{ id: string }> }).data;
      const ids = runs.map((r) => r.id);
      expect(ids).not.toContain(pendingId);
    });

    it("includes the seeded CANCELLED run", async () => {
      const { body } = await apiGet("/api/validation-runs?status=CANCELLED");
      const runs = (body as { data: Array<{ id: string }> }).data;
      const ids = runs.map((r) => r.id);
      expect(ids).toContain(cancelledId);
    });
  });

  // -------------------------------------------------------------------------
  // Test 2: filter by commitSha prefix '3a2b'
  // -------------------------------------------------------------------------

  describe("GET /api/validation-runs?commitSha=3a2b", () => {
    let matchingId: string;
    let nonMatchingId: string;

    beforeAll(async () => {
      matchingId = await seedRun(client, {
        branch: "sha-filter-branch",
        commitSha: "3a2bfeed1234567890abcdef",
        status: "SUCCESS",
        userId: "api_test_user",
      });
      nonMatchingId = await seedRun(client, {
        branch: "sha-filter-branch",
        commitSha: "deadbeef1234567890abcdef",
        status: "SUCCESS",
        userId: "api_test_user",
      });
      seededRunIds.push(matchingId, nonMatchingId);
    });

    it("returns HTTP 200", async () => {
      const { status } = await apiGet("/api/validation-runs?commitSha=3a2b");
      expect(status).toBe(200);
    });

    it("returns runs whose commitSha starts with the provided prefix", async () => {
      const { body } = await apiGet("/api/validation-runs?commitSha=3a2b");
      const runs = (body as { data: Array<{ id: string; commitSha: string }> }).data;
      expect(Array.isArray(runs)).toBe(true);
      expect(runs.every((r) => r.commitSha.startsWith("3a2b"))).toBe(true);
    });

    it("includes the matching seeded run", async () => {
      const { body } = await apiGet("/api/validation-runs?commitSha=3a2b");
      const runs = (body as { data: Array<{ id: string }> }).data;
      const ids = runs.map((r) => r.id);
      expect(ids).toContain(matchingId);
    });

    it("excludes runs whose commitSha does not start with the prefix", async () => {
      const { body } = await apiGet("/api/validation-runs?commitSha=3a2b");
      const runs = (body as { data: Array<{ id: string }> }).data;
      const ids = runs.map((r) => r.id);
      expect(ids).not.toContain(nonMatchingId);
    });
  });

  // -------------------------------------------------------------------------
  // Test 3: combined branch + status filter
  // -------------------------------------------------------------------------

  describe("GET /api/validation-runs?branch=combined-test&status=SUCCESS", () => {
    let matchBothId: string;
    let wrongStatusId: string;
    let wrongBranchId: string;

    beforeAll(async () => {
      matchBothId = await seedRun(client, {
        branch: "combined-test",
        commitSha: `comb01${randomUUID().replace(/-/g, "").slice(0, 10)}`,
        status: "SUCCESS",
        userId: "api_test_user",
      });
      wrongStatusId = await seedRun(client, {
        branch: "combined-test",
        commitSha: `comb02${randomUUID().replace(/-/g, "").slice(0, 10)}`,
        status: "FAILED",
        userId: "api_test_user",
      });
      wrongBranchId = await seedRun(client, {
        branch: "other-branch",
        commitSha: `comb03${randomUUID().replace(/-/g, "").slice(0, 10)}`,
        status: "SUCCESS",
        userId: "api_test_user",
      });
      seededRunIds.push(matchBothId, wrongStatusId, wrongBranchId);
    });

    it("returns HTTP 200", async () => {
      const { status } = await apiGet(
        "/api/validation-runs?branch=combined-test&status=SUCCESS"
      );
      expect(status).toBe(200);
    });

    it("returns only runs matching both branch and status", async () => {
      const { body } = await apiGet(
        "/api/validation-runs?branch=combined-test&status=SUCCESS"
      );
      const runs = (
        body as { data: Array<{ id: string; branchName: string; status: string }> }
      ).data;
      expect(Array.isArray(runs)).toBe(true);
      expect(
        runs.every(
          (r) => r.branchName === "combined-test" && r.status === "SUCCESS"
        )
      ).toBe(true);
    });

    it("includes the run matching both filters", async () => {
      const { body } = await apiGet(
        "/api/validation-runs?branch=combined-test&status=SUCCESS"
      );
      const runs = (body as { data: Array<{ id: string }> }).data;
      expect(runs.map((r) => r.id)).toContain(matchBothId);
    });

    it("excludes run with wrong status (FAILED instead of SUCCESS)", async () => {
      const { body } = await apiGet(
        "/api/validation-runs?branch=combined-test&status=SUCCESS"
      );
      const runs = (body as { data: Array<{ id: string }> }).data;
      expect(runs.map((r) => r.id)).not.toContain(wrongStatusId);
    });

    it("excludes run with wrong branch (other-branch instead of combined-test)", async () => {
      const { body } = await apiGet(
        "/api/validation-runs?branch=combined-test&status=SUCCESS"
      );
      const runs = (body as { data: Array<{ id: string }> }).data;
      expect(runs.map((r) => r.id)).not.toContain(wrongBranchId);
    });
  });

  // -------------------------------------------------------------------------
  // Test 4 & 5: detail endpoint — GET /api/validation-runs/:id
  // -------------------------------------------------------------------------

  describe("GET /api/validation-runs/:id", () => {
    let detailRunId: string;

    beforeAll(async () => {
      detailRunId = await seedRun(client, {
        branch: "detail-test-branch",
        commitSha: `detail${randomUUID().replace(/-/g, "").slice(0, 10)}`,
        status: "RUNNING",
        userId: "api_test_user",
        withAllStages: true,
      });
      seededRunIds.push(detailRunId);
    });

    it("returns HTTP 200 for an existing run", async () => {
      const { status } = await apiGet(`/api/validation-runs/${detailRunId}`);
      expect(status).toBe(200);
    });

    it("response body includes an id field matching the requested run", async () => {
      const { body } = await apiGet(`/api/validation-runs/${detailRunId}`);
      const run = body as { id: string };
      expect(run.id).toBe(detailRunId);
    });

    it("response includes a stages array", async () => {
      const { body } = await apiGet(`/api/validation-runs/${detailRunId}`);
      const run = body as { stages?: unknown };
      expect(Array.isArray(run.stages)).toBe(true);
    });

    it("stages array contains all 5 stages", async () => {
      const { body } = await apiGet(`/api/validation-runs/${detailRunId}`);
      const run = body as { stages: unknown[] };
      expect(run.stages).toHaveLength(5);
    });

    it("each stage has stageName field", async () => {
      const { body } = await apiGet(`/api/validation-runs/${detailRunId}`);
      const run = body as { stages: Array<{ stageName?: unknown }> };
      expect(run.stages.every((s) => typeof s.stageName === "string")).toBe(true);
    });

    it("each stage has stageIndex field", async () => {
      const { body } = await apiGet(`/api/validation-runs/${detailRunId}`);
      const run = body as { stages: Array<{ stageIndex?: unknown }> };
      expect(run.stages.every((s) => typeof s.stageIndex === "number")).toBe(true);
    });

    it("each stage has status field", async () => {
      const { body } = await apiGet(`/api/validation-runs/${detailRunId}`);
      const run = body as { stages: Array<{ status?: unknown }> };
      expect(run.stages.every((s) => typeof s.status === "string")).toBe(true);
    });

    it("each stage has logs field", async () => {
      const { body } = await apiGet(`/api/validation-runs/${detailRunId}`);
      const run = body as { stages: Array<{ logs?: unknown }> };
      expect(run.stages.every((s) => typeof s.logs === "string")).toBe(true);
    });

    it("each stage has startedAt field (string or null)", async () => {
      const { body } = await apiGet(`/api/validation-runs/${detailRunId}`);
      const run = body as { stages: Array<{ startedAt?: unknown }> };
      expect(
        run.stages.every(
          (s) => s.startedAt === null || typeof s.startedAt === "string"
        )
      ).toBe(true);
    });

    it("each stage has completedAt field (string or null)", async () => {
      const { body } = await apiGet(`/api/validation-runs/${detailRunId}`);
      const run = body as { stages: Array<{ completedAt?: unknown }> };
      expect(
        run.stages.every(
          (s) => s.completedAt === null || typeof s.completedAt === "string"
        )
      ).toBe(true);
    });

    it("all 5 canonical stage names are present", async () => {
      const { body } = await apiGet(`/api/validation-runs/${detailRunId}`);
      const run = body as { stages: Array<{ stageName: string }> };
      const names = run.stages.map((s) => s.stageName).sort();
      expect(names).toEqual([...STAGE_TYPES].sort());
    });

    // Test 5: stages ordered by stageIndex ascending
    it("stages are ordered by stageIndex ascending", async () => {
      const { body } = await apiGet(`/api/validation-runs/${detailRunId}`);
      const run = body as { stages: Array<{ stageIndex: number }> };
      const indices = run.stages.map((s) => s.stageIndex);
      const sorted = [...indices].sort((a, b) => a - b);
      expect(indices).toEqual(sorted);
    });

    it("first stage has stageIndex 0 (SETUP)", async () => {
      const { body } = await apiGet(`/api/validation-runs/${detailRunId}`);
      const run = body as { stages: Array<{ stageIndex: number; stageName: string }> };
      expect(run.stages[0].stageIndex).toBe(0);
      expect(run.stages[0].stageName).toBe("SETUP");
    });

    it("last stage has stageIndex 4 (DEPLOY)", async () => {
      const { body } = await apiGet(`/api/validation-runs/${detailRunId}`);
      const run = body as { stages: Array<{ stageIndex: number; stageName: string }> };
      const last = run.stages[run.stages.length - 1];
      expect(last.stageIndex).toBe(4);
      expect(last.stageName).toBe("DEPLOY");
    });
  });

  // -------------------------------------------------------------------------
  // Test 6: 404 for non-existent ID
  // -------------------------------------------------------------------------

  describe("GET /api/validation-runs/:id — non-existent", () => {
    it("returns HTTP 404 for a UUID that does not exist", async () => {
      const nonExistentId = randomUUID(); // fresh UUID — highly unlikely to exist
      const { status } = await apiGet(`/api/validation-runs/${nonExistentId}`);
      expect(status).toBe(404);
    });

    it("response body for 404 includes an error or message field", async () => {
      const nonExistentId = randomUUID();
      const { body } = await apiGet(`/api/validation-runs/${nonExistentId}`);
      const payload = body as Record<string, unknown>;
      const hasErrorField =
        typeof payload.error === "string" ||
        typeof payload.message === "string";
      expect(hasErrorField).toBe(true);
    });
  });

  // -------------------------------------------------------------------------
  // Test 7: limit parameter caps at 100
  // -------------------------------------------------------------------------

  describe("GET /api/validation-runs?limit=500", () => {
    beforeAll(async () => {
      // Seed 110 runs to ensure the DB has more than 100 matching records
      await client.query("BEGIN");
      try {
        for (let i = 0; i < 110; i++) {
          const runId = randomUUID();
          seededRunIds.push(runId);
          await client.query(
            `INSERT INTO validation_runs
               (id, "branchName", "commitSha", status, config, "userId", "createdAt", "updatedAt")
             VALUES ($1, 'limit-test-branch', $2, 'SUCCESS'::"ValidationRunStatus",
                     $3, 'api_test_user', NOW() - ($4 * INTERVAL '1 second'), NOW())`,
            [
              runId,
              `limitsha${i.toString().padStart(4, "0")}abcdef`,
              JSON.stringify({ timeout: 300 }),
              i,
            ]
          );
        }
        await client.query("COMMIT");
      } catch (err) {
        await client.query("ROLLBACK");
        throw err;
      }
    });

    it("returns HTTP 200", async () => {
      const { status } = await apiGet(
        "/api/validation-runs?limit=500&branch=limit-test-branch"
      );
      expect(status).toBe(200);
    });

    it("returns at most 100 runs even when limit=500 is requested", async () => {
      const { body } = await apiGet(
        "/api/validation-runs?limit=500&branch=limit-test-branch"
      );
      const runs = (body as { data: unknown[] }).data;
      expect(Array.isArray(runs)).toBe(true);
      expect(runs.length).toBeLessThanOrEqual(100);
    });

    it("returns exactly 100 runs when more than 100 matching rows exist", async () => {
      const { body } = await apiGet(
        "/api/validation-runs?limit=500&branch=limit-test-branch"
      );
      const runs = (body as { data: unknown[] }).data;
      expect(runs.length).toBe(100);
    });
  });
});
