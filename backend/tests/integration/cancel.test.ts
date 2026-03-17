/**
 * Integration tests for TASK-4606 — POST /api/validation-runs/{id}/cancel endpoint.
 *
 * For each status variant, these tests:
 *   1. Seed a validation_run in the target status (and any run_stages).
 *   2. POST to /api/validation-runs/:id/cancel.
 *   3. Assert the HTTP response status and body.
 *   4. Where applicable, assert the updated DB state.
 *
 * Prerequisites:
 *   - DATABASE_URL env var pointing to a running PostgreSQL instance.
 *   - API_URL env var pointing to the running API server (e.g. http://localhost:3000).
 *   - Schema (including migration 001_initial and later) must be applied.
 */

import { Client } from "pg";
import { randomUUID } from "crypto";

// ---------------------------------------------------------------------------
// Constants & types
// ---------------------------------------------------------------------------

type RunStatus = "PENDING" | "RUNNING" | "SUCCESS" | "FAILED" | "CANCELLED";
type StageStatus = "PENDING" | "RUNNING" | "SUCCESS" | "FAILED" | "CANCELLED";

const API_URL = process.env.API_URL ?? "http://localhost:3000";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

async function seedRun(
  client: Client,
  status: RunStatus,
  opts: { cancelledAt?: Date | null } = {}
): Promise<string> {
  const id = randomUUID();
  const now = new Date();
  await client.query(
    `INSERT INTO validation_runs
       (id, "branchName", "commitSha", status, config, "userId", "createdAt", "updatedAt", "cancelledAt")
     VALUES ($1, $2, $3, $4::"ValidationRunStatus", $5, $6, $7, $7, $8)`,
    [
      id,
      "feature/test-cancel",
      `sha-cancel-${id.slice(0, 8)}`,
      status,
      JSON.stringify({ timeout: 300 }),
      "user_cancel_test",
      now,
      opts.cancelledAt ?? null,
    ]
  );
  return id;
}

async function seedStage(
  client: Client,
  runId: string,
  stageIndex: number,
  status: StageStatus
): Promise<string> {
  const stageId = randomUUID();
  await client.query(
    `INSERT INTO run_stages
       (id, "runId", "stageName", "stageIndex", status)
     VALUES ($1, $2, $3::"StageType", $4, $5::"StageStatus")`,
    [stageId, runId, "SETUP", stageIndex, status]
  );
  return stageId;
}

async function cancelRun(
  id: string
): Promise<{ status: number; body: unknown }> {
  const response = await fetch(`${API_URL}/api/validation-runs/${id}/cancel`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
  });
  const body = await response.json().catch(() => null);
  return { status: response.status, body };
}

async function getRunStatus(client: Client, id: string): Promise<RunStatus | null> {
  const result = await client.query<{ status: RunStatus }>(
    `SELECT status FROM validation_runs WHERE id = $1`,
    [id]
  );
  return result.rows[0]?.status ?? null;
}

async function getStageStatuses(
  client: Client,
  runId: string
): Promise<Array<{ stageIndex: number; status: StageStatus }>> {
  const result = await client.query<{ stageIndex: number; status: StageStatus }>(
    `SELECT "stageIndex", status FROM run_stages WHERE "runId" = $1 ORDER BY "stageIndex"`,
    [runId]
  );
  return result.rows;
}

// ---------------------------------------------------------------------------
// Test suite
// ---------------------------------------------------------------------------

describe("POST /api/validation-runs/:id/cancel (TASK-4606)", () => {
  let client: Client;

  beforeAll(async () => {
    client = new Client({ connectionString: process.env.DATABASE_URL });
    await client.connect();
  });

  afterAll(async () => {
    // Clean up all test data inserted by this suite
    await client.query(
      `DELETE FROM validation_runs WHERE "userId" = 'user_cancel_test'`
    );
    await client.end();
  });

  // -------------------------------------------------------------------------
  // 200 — RUNNING run is cancellable
  // -------------------------------------------------------------------------

  it("returns 200 and updated run with status=CANCELLED when run is RUNNING", async () => {
    const id = await seedRun(client, "RUNNING");

    const { status, body } = await cancelRun(id);

    expect(status).toBe(200);
    expect(body).toMatchObject({
      id,
      status: "CANCELLED",
    });

    // Verify DB state
    const dbStatus = await getRunStatus(client, id);
    expect(dbStatus).toBe("CANCELLED");
  });

  it("response body for RUNNING cancel includes cancelledAt timestamp", async () => {
    const id = await seedRun(client, "RUNNING");

    const { status, body } = await cancelRun(id);

    expect(status).toBe(200);
    expect((body as Record<string, unknown>).cancelledAt).toBeTruthy();
  });

  // -------------------------------------------------------------------------
  // 200 — PENDING run is cancellable (no pipeline token, just DB update)
  // -------------------------------------------------------------------------

  it("returns 200 and updated run with status=CANCELLED when run is PENDING", async () => {
    const id = await seedRun(client, "PENDING");

    const { status, body } = await cancelRun(id);

    expect(status).toBe(200);
    expect(body).toMatchObject({
      id,
      status: "CANCELLED",
    });

    // Verify DB state
    const dbStatus = await getRunStatus(client, id);
    expect(dbStatus).toBe("CANCELLED");
  });

  it("response body for PENDING cancel includes cancelledAt timestamp", async () => {
    const id = await seedRun(client, "PENDING");

    const { status, body } = await cancelRun(id);

    expect(status).toBe(200);
    expect((body as Record<string, unknown>).cancelledAt).toBeTruthy();
  });

  // -------------------------------------------------------------------------
  // 409 — FAILED run is NOT cancellable
  // -------------------------------------------------------------------------

  it("returns 409 with error RUN_NOT_CANCELLABLE when run is FAILED", async () => {
    const id = await seedRun(client, "FAILED");

    const { status, body } = await cancelRun(id);

    expect(status).toBe(409);
    expect((body as Record<string, unknown>).error).toBe("RUN_NOT_CANCELLABLE");
  });

  it("does not change DB status when cancelling a FAILED run", async () => {
    const id = await seedRun(client, "FAILED");

    await cancelRun(id);

    const dbStatus = await getRunStatus(client, id);
    expect(dbStatus).toBe("FAILED");
  });

  // -------------------------------------------------------------------------
  // 409 — SUCCESS run is NOT cancellable
  // -------------------------------------------------------------------------

  it("returns 409 with error RUN_NOT_CANCELLABLE when run is SUCCESS", async () => {
    const id = await seedRun(client, "SUCCESS");

    const { status, body } = await cancelRun(id);

    expect(status).toBe(409);
    expect((body as Record<string, unknown>).error).toBe("RUN_NOT_CANCELLABLE");
  });

  it("does not change DB status when cancelling a SUCCESS run", async () => {
    const id = await seedRun(client, "SUCCESS");

    await cancelRun(id);

    const dbStatus = await getRunStatus(client, id);
    expect(dbStatus).toBe("SUCCESS");
  });

  // -------------------------------------------------------------------------
  // 409 — CANCELLED run is NOT cancellable (idempotency not required)
  // -------------------------------------------------------------------------

  it("returns 409 with error RUN_NOT_CANCELLABLE when run is already CANCELLED", async () => {
    const id = await seedRun(client, "CANCELLED", { cancelledAt: new Date() });

    const { status, body } = await cancelRun(id);

    expect(status).toBe(409);
    expect((body as Record<string, unknown>).error).toBe("RUN_NOT_CANCELLABLE");
  });

  // -------------------------------------------------------------------------
  // 404 — unknown id
  // -------------------------------------------------------------------------

  it("returns 404 when run id does not exist", async () => {
    const nonExistentId = randomUUID();

    const { status } = await cancelRun(nonExistentId);

    expect(status).toBe(404);
  });

  it("returns 404 body with meaningful error for unknown id", async () => {
    const nonExistentId = randomUUID();

    const { status, body } = await cancelRun(nonExistentId);

    expect(status).toBe(404);
    expect(body).toBeTruthy();
  });

  // -------------------------------------------------------------------------
  // Stage results preservation — completed stages remain untouched
  // -------------------------------------------------------------------------

  it("preserves SUCCESS stage status when run is cancelled (RUNNING)", async () => {
    const id = await seedRun(client, "RUNNING");
    // Stage 0 completed successfully before cancellation
    await seedStage(client, id, 0, "SUCCESS");
    // Stage 1 was in progress when cancel was triggered
    await seedStage(client, id, 1, "RUNNING");

    const { status } = await cancelRun(id);
    expect(status).toBe(200);

    const stages = await getStageStatuses(client, id);
    const stage0 = stages.find((s) => s.stageIndex === 0);
    expect(stage0?.status).toBe("SUCCESS");
  });

  it("preserves FAILED stage status when run is cancelled (RUNNING)", async () => {
    const id = await seedRun(client, "RUNNING");
    // Stage 0 already failed before cancel
    await seedStage(client, id, 0, "FAILED");

    const { status } = await cancelRun(id);
    expect(status).toBe(200);

    const stages = await getStageStatuses(client, id);
    const stage0 = stages.find((s) => s.stageIndex === 0);
    expect(stage0?.status).toBe("FAILED");
  });

  it("preserves all SUCCESS stages for a PENDING run on cancel", async () => {
    const id = await seedRun(client, "PENDING");
    // All stages pending (no pipeline started yet), but stage 0 was pre-queued as SUCCESS
    await seedStage(client, id, 0, "SUCCESS");
    await seedStage(client, id, 1, "PENDING");

    const { status } = await cancelRun(id);
    expect(status).toBe(200);

    const stages = await getStageStatuses(client, id);
    const successStages = stages.filter((s) => s.status === "SUCCESS");
    expect(successStages).toHaveLength(1);
    expect(successStages[0].stageIndex).toBe(0);
  });

  // -------------------------------------------------------------------------
  // Response shape — cancelled run record structure
  // -------------------------------------------------------------------------

  it("returned run record includes all expected fields", async () => {
    const id = await seedRun(client, "RUNNING");

    const { status, body } = await cancelRun(id);
    expect(status).toBe(200);

    const run = body as Record<string, unknown>;
    expect(run.id).toBe(id);
    expect(run.status).toBe("CANCELLED");
    expect(run.cancelledAt).toBeTruthy();
    expect(run.updatedAt).toBeTruthy();
  });
});
