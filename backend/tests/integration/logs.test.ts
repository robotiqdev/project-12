/**
 * Integration tests for TASK-4589 — Log Export as ZIP and clipboard copy service.
 *
 * Tests the three log-export endpoints:
 *   GET /api/validation-runs/:id/logs/zip   — download all logs as a ZIP archive
 *   GET /api/validation-runs/:id/logs/text  — concatenated plain-text log dump
 *   GET /api/validation-runs/:id/stages/:stageIndex/logs — single-stage log
 *
 * Prerequisites: DATABASE_URL env var pointing to a running PostgreSQL instance
 * with the schema already applied, and the Express app listening.
 */

import supertest from "supertest";
import { Client } from "pg";
import { randomUUID } from "crypto";
import Unzipper from "unzipper";
import app from "../../src/app";

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const STAGE_NAMES = ["SETUP", "LINT", "TEST", "BUILD", "DEPLOY"] as const;

const SAMPLE_LOGS: Record<string, string> = {
  SETUP:  "Installing dependencies...\nnpm install complete\n",
  LINT:   "Running ESLint...\nNo lint errors found\n",
  TEST:   "Running unit tests...\n42 passed, 0 failed\n",
  BUILD:  "Building project...\nBuild successful\n",
  DEPLOY: "Deploying to staging...\nDeploy complete\n",
};

const UNKNOWN_RUN_ID = "00000000-0000-0000-0000-000000000000";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Read all entries from a ZIP buffer and return a map of filename → content. */
async function parseZipBuffer(
  buf: Buffer
): Promise<Map<string, string>> {
  const entries = new Map<string, string>();
  const directory = await Unzipper.Open.buffer(buf);
  for (const file of directory.files) {
    const content = await file.buffer();
    entries.set(file.path, content.toString("utf8"));
  }
  return entries;
}

// ---------------------------------------------------------------------------
// Test suite
// ---------------------------------------------------------------------------

describe("Log Export Service (TASK-4589)", () => {
  let client: Client;
  let seededRunId: string;

  // -------------------------------------------------------------------------
  // Setup / teardown
  // -------------------------------------------------------------------------

  beforeAll(async () => {
    client = new Client({ connectionString: process.env.DATABASE_URL });
    await client.connect();

    seededRunId = randomUUID();
    const testUserId = `log_export_test_user_${randomUUID()}`;

    // Insert a ValidationRun with SUCCESS status
    await client.query(
      `INSERT INTO validation_runs
         (id, "branchName", "commitSha", status, config, "userId", "createdAt", "updatedAt", "completedAt")
       VALUES ($1, 'main', 'abc123', 'SUCCESS'::"ValidationRunStatus",
               '{"timeout":300}', $2, NOW() - INTERVAL '5 minutes', NOW(), NOW())`,
      [seededRunId, testUserId]
    );

    // Insert all 5 stages with sample log content
    for (let i = 0; i < STAGE_NAMES.length; i++) {
      const stageName = STAGE_NAMES[i];
      const stageId = randomUUID();
      await client.query(
        `INSERT INTO run_stages
           (id, "runId", "stageName", "stageIndex", status, logs,
            "startedAt", "completedAt")
         VALUES ($1, $2, $3::"StageType", $4, 'SUCCESS'::"StageStatus",
                 $5, NOW() - INTERVAL '${5 - i} minutes', NOW() - INTERVAL '${4 - i} minutes')`,
        [stageId, seededRunId, stageName, i, SAMPLE_LOGS[stageName]]
      );
    }
  });

  afterAll(async () => {
    if (client) {
      await client.query(
        `DELETE FROM validation_runs WHERE id = $1`,
        [seededRunId]
      );
      await client.end();
    }
  });

  // -------------------------------------------------------------------------
  // GET /api/validation-runs/:id/logs/zip — response headers
  // -------------------------------------------------------------------------

  describe("GET /api/validation-runs/:id/logs/zip", () => {
    it("returns HTTP 200 for a known run", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/zip`)
        .buffer(true)
        .parse((res, callback) => {
          const chunks: Buffer[] = [];
          res.on("data", (chunk: Buffer) => chunks.push(chunk));
          res.on("end", () => callback(null, Buffer.concat(chunks)));
        });
      expect(res.status).toBe(200);
    });

    it("returns Content-Type: application/zip", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/zip`)
        .buffer(true)
        .parse((res, callback) => {
          const chunks: Buffer[] = [];
          res.on("data", (chunk: Buffer) => chunks.push(chunk));
          res.on("end", () => callback(null, Buffer.concat(chunks)));
        });
      expect(res.headers["content-type"]).toMatch(/application\/zip/);
    });

    it("returns Content-Disposition: attachment; filename=\"run-{id}.zip\"", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/zip`)
        .buffer(true)
        .parse((res, callback) => {
          const chunks: Buffer[] = [];
          res.on("data", (chunk: Buffer) => chunks.push(chunk));
          res.on("end", () => callback(null, Buffer.concat(chunks)));
        });
      const disposition = res.headers["content-disposition"];
      expect(disposition).toMatch(/attachment/);
      expect(disposition).toContain(`filename="run-${seededRunId}.zip"`);
    });

    it("returns a valid ZIP archive (magic bytes PK\\x03\\x04)", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/zip`)
        .buffer(true)
        .parse((res, callback) => {
          const chunks: Buffer[] = [];
          res.on("data", (chunk: Buffer) => chunks.push(chunk));
          res.on("end", () => callback(null, Buffer.concat(chunks)));
        });
      const body = res.body as Buffer;
      // ZIP files start with PK signature: 0x50 0x4B 0x03 0x04
      expect(body[0]).toBe(0x50); // P
      expect(body[1]).toBe(0x4b); // K
      expect(body[2]).toBe(0x03);
      expect(body[3]).toBe(0x04);
    });

    it("ZIP contains stage-0-SETUP.log", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/zip`)
        .buffer(true)
        .parse((res, callback) => {
          const chunks: Buffer[] = [];
          res.on("data", (chunk: Buffer) => chunks.push(chunk));
          res.on("end", () => callback(null, Buffer.concat(chunks)));
        });
      const entries = await parseZipBuffer(res.body as Buffer);
      expect(entries.has("stage-0-SETUP.log")).toBe(true);
    });

    it("ZIP contains stage-1-LINT.log", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/zip`)
        .buffer(true)
        .parse((res, callback) => {
          const chunks: Buffer[] = [];
          res.on("data", (chunk: Buffer) => chunks.push(chunk));
          res.on("end", () => callback(null, Buffer.concat(chunks)));
        });
      const entries = await parseZipBuffer(res.body as Buffer);
      expect(entries.has("stage-1-LINT.log")).toBe(true);
    });

    it("ZIP contains stage-2-TEST.log", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/zip`)
        .buffer(true)
        .parse((res, callback) => {
          const chunks: Buffer[] = [];
          res.on("data", (chunk: Buffer) => chunks.push(chunk));
          res.on("end", () => callback(null, Buffer.concat(chunks)));
        });
      const entries = await parseZipBuffer(res.body as Buffer);
      expect(entries.has("stage-2-TEST.log")).toBe(true);
    });

    it("ZIP contains stage-3-BUILD.log", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/zip`)
        .buffer(true)
        .parse((res, callback) => {
          const chunks: Buffer[] = [];
          res.on("data", (chunk: Buffer) => chunks.push(chunk));
          res.on("end", () => callback(null, Buffer.concat(chunks)));
        });
      const entries = await parseZipBuffer(res.body as Buffer);
      expect(entries.has("stage-3-BUILD.log")).toBe(true);
    });

    it("ZIP contains stage-4-DEPLOY.log", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/zip`)
        .buffer(true)
        .parse((res, callback) => {
          const chunks: Buffer[] = [];
          res.on("data", (chunk: Buffer) => chunks.push(chunk));
          res.on("end", () => callback(null, Buffer.concat(chunks)));
        });
      const entries = await parseZipBuffer(res.body as Buffer);
      expect(entries.has("stage-4-DEPLOY.log")).toBe(true);
    });

    it("ZIP contains metadata.json", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/zip`)
        .buffer(true)
        .parse((res, callback) => {
          const chunks: Buffer[] = [];
          res.on("data", (chunk: Buffer) => chunks.push(chunk));
          res.on("end", () => callback(null, Buffer.concat(chunks)));
        });
      const entries = await parseZipBuffer(res.body as Buffer);
      expect(entries.has("metadata.json")).toBe(true);
    });

    it("ZIP contains exactly 6 entries (5 stage logs + metadata.json)", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/zip`)
        .buffer(true)
        .parse((res, callback) => {
          const chunks: Buffer[] = [];
          res.on("data", (chunk: Buffer) => chunks.push(chunk));
          res.on("end", () => callback(null, Buffer.concat(chunks)));
        });
      const entries = await parseZipBuffer(res.body as Buffer);
      expect(entries.size).toBe(6);
    });

    it("ZIP stage log content matches seeded logs", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/zip`)
        .buffer(true)
        .parse((res, callback) => {
          const chunks: Buffer[] = [];
          res.on("data", (chunk: Buffer) => chunks.push(chunk));
          res.on("end", () => callback(null, Buffer.concat(chunks)));
        });
      const entries = await parseZipBuffer(res.body as Buffer);
      expect(entries.get("stage-0-SETUP.log")).toContain("Installing dependencies");
      expect(entries.get("stage-1-LINT.log")).toContain("Running ESLint");
      expect(entries.get("stage-2-TEST.log")).toContain("Running unit tests");
      expect(entries.get("stage-3-BUILD.log")).toContain("Building project");
      expect(entries.get("stage-4-DEPLOY.log")).toContain("Deploying to staging");
    });

    it("ZIP metadata.json contains run id", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/zip`)
        .buffer(true)
        .parse((res, callback) => {
          const chunks: Buffer[] = [];
          res.on("data", (chunk: Buffer) => chunks.push(chunk));
          res.on("end", () => callback(null, Buffer.concat(chunks)));
        });
      const entries = await parseZipBuffer(res.body as Buffer);
      const metadata = JSON.parse(entries.get("metadata.json") ?? "{}");
      expect(metadata.id).toBe(seededRunId);
    });

    it("ZIP metadata.json contains run status", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/zip`)
        .buffer(true)
        .parse((res, callback) => {
          const chunks: Buffer[] = [];
          res.on("data", (chunk: Buffer) => chunks.push(chunk));
          res.on("end", () => callback(null, Buffer.concat(chunks)));
        });
      const entries = await parseZipBuffer(res.body as Buffer);
      const metadata = JSON.parse(entries.get("metadata.json") ?? "{}");
      expect(metadata.status).toBe("SUCCESS");
    });

    it("ZIP metadata.json contains timestamps", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/zip`)
        .buffer(true)
        .parse((res, callback) => {
          const chunks: Buffer[] = [];
          res.on("data", (chunk: Buffer) => chunks.push(chunk));
          res.on("end", () => callback(null, Buffer.concat(chunks)));
        });
      const entries = await parseZipBuffer(res.body as Buffer);
      const metadata = JSON.parse(entries.get("metadata.json") ?? "{}");
      expect(metadata.createdAt).toBeDefined();
    });

    it("returns 404 for unknown runId", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${UNKNOWN_RUN_ID}/logs/zip`);
      expect(res.status).toBe(404);
    });
  });

  // -------------------------------------------------------------------------
  // GET /api/validation-runs/:id/logs/text — concatenated plain-text
  // -------------------------------------------------------------------------

  describe("GET /api/validation-runs/:id/logs/text", () => {
    it("returns HTTP 200 for a known run", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/text`);
      expect(res.status).toBe(200);
    });

    it("returns Content-Type: text/plain", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/text`);
      expect(res.headers["content-type"]).toMatch(/text\/plain/);
    });

    it("response body contains SETUP stage log content", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/text`);
      expect(res.text).toContain("Installing dependencies");
    });

    it("response body contains LINT stage log content", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/text`);
      expect(res.text).toContain("Running ESLint");
    });

    it("response body contains TEST stage log content", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/text`);
      expect(res.text).toContain("Running unit tests");
    });

    it("response body contains BUILD stage log content", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/text`);
      expect(res.text).toContain("Building project");
    });

    it("response body contains DEPLOY stage log content", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/text`);
      expect(res.text).toContain("Deploying to staging");
    });

    it("response body contains SETUP stage header", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/text`);
      expect(res.text).toContain("=== STAGE: SETUP ===");
    });

    it("response body contains LINT stage header", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/text`);
      expect(res.text).toContain("=== STAGE: LINT ===");
    });

    it("response body contains TEST stage header", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/text`);
      expect(res.text).toContain("=== STAGE: TEST ===");
    });

    it("response body contains BUILD stage header", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/text`);
      expect(res.text).toContain("=== STAGE: BUILD ===");
    });

    it("response body contains DEPLOY stage header", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/text`);
      expect(res.text).toContain("=== STAGE: DEPLOY ===");
    });

    it("stages appear in order (SETUP before DEPLOY)", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/text`);
      const setupPos = res.text.indexOf("=== STAGE: SETUP ===");
      const deployPos = res.text.indexOf("=== STAGE: DEPLOY ===");
      expect(setupPos).toBeGreaterThanOrEqual(0);
      expect(deployPos).toBeGreaterThan(setupPos);
    });

    it("stages are separated by --- delimiter", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/logs/text`);
      expect(res.text).toContain("---");
    });

    it("returns 404 for unknown runId", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${UNKNOWN_RUN_ID}/logs/text`);
      expect(res.status).toBe(404);
    });
  });

  // -------------------------------------------------------------------------
  // GET /api/validation-runs/:id/stages/:stageIndex/logs — single stage log
  // -------------------------------------------------------------------------

  describe("GET /api/validation-runs/:id/stages/:stageIndex/logs", () => {
    it("returns HTTP 200 for stageIndex 0 (SETUP)", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/stages/0/logs`);
      expect(res.status).toBe(200);
    });

    it("returns only SETUP log content for stageIndex 0", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/stages/0/logs`);
      expect(res.text).toContain("Installing dependencies");
    });

    it("does NOT include other stage logs for stageIndex 0", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/stages/0/logs`);
      expect(res.text).not.toContain("Running ESLint");
      expect(res.text).not.toContain("Running unit tests");
      expect(res.text).not.toContain("Building project");
      expect(res.text).not.toContain("Deploying to staging");
    });

    it("returns HTTP 200 for stageIndex 1 (LINT)", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/stages/1/logs`);
      expect(res.status).toBe(200);
    });

    it("returns only LINT log content for stageIndex 1", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/stages/1/logs`);
      expect(res.text).toContain("Running ESLint");
      expect(res.text).not.toContain("Installing dependencies");
    });

    it("returns HTTP 200 for stageIndex 2 (TEST)", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/stages/2/logs`);
      expect(res.status).toBe(200);
    });

    it("returns only TEST log content for stageIndex 2", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/stages/2/logs`);
      expect(res.text).toContain("Running unit tests");
      expect(res.text).not.toContain("Installing dependencies");
    });

    it("returns HTTP 200 for stageIndex 3 (BUILD)", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/stages/3/logs`);
      expect(res.status).toBe(200);
    });

    it("returns only BUILD log content for stageIndex 3", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/stages/3/logs`);
      expect(res.text).toContain("Building project");
      expect(res.text).not.toContain("Deploying to staging");
    });

    it("returns HTTP 200 for stageIndex 4 (DEPLOY)", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/stages/4/logs`);
      expect(res.status).toBe(200);
    });

    it("returns only DEPLOY log content for stageIndex 4", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/stages/4/logs`);
      expect(res.text).toContain("Deploying to staging");
      expect(res.text).not.toContain("Running ESLint");
    });

    it("returns 404 for unknown runId", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${UNKNOWN_RUN_ID}/stages/0/logs`);
      expect(res.status).toBe(404);
    });

    it("returns 404 for valid runId but non-existent stageIndex", async () => {
      const res = await supertest(app)
        .get(`/api/validation-runs/${seededRunId}/stages/99/logs`);
      expect(res.status).toBe(404);
    });
  });
});
