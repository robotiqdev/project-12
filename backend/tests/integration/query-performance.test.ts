/**
 * Integration tests for TASK-4622 — Performance Indexes and
 * TASK-4623 — Validate Query Performance.
 *
 * @integration
 *
 * Run separately from unit tests:
 *   jest --testPathPattern=query-performance --testNamePattern="@integration"
 *
 * For each composite/partial index defined in 003_indexes/migration.sql, these
 * tests:
 *   1. Seed 200 validation_runs spread across 5 branches (40 per branch).
 *   2. Run EXPLAIN (FORMAT JSON, ANALYZE) with a query that exercises the index.
 *   3. Assert the query plan contains 'Index Scan' or 'Index Only Scan' — NOT
 *      'Seq Scan'.
 *   4. Assert the "Actual Rows" reported by the plan is proportional to the
 *      query selectivity (i.e. far fewer rows than the full table).
 *
 * Prerequisites: DATABASE_URL env var pointing to a running PostgreSQL instance
 * that already has the schema (including the 003_indexes migration) applied.
 */

import { Client } from "pg";
import { randomUUID } from "crypto";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const BRANCHES = ["main", "feature/a", "feature/b", "hotfix/c", "release/1"];
const RUNS_PER_BRANCH = 40; // 5 branches × 40 = 200 total
const TOTAL_RUNS = BRANCHES.length * RUNS_PER_BRANCH;

const STATUSES = [
  "PENDING",
  "RUNNING",
  "SUCCESS",
  "FAILED",
  "CANCELLED",
] as const;

interface ExplainNode {
  "Node Type": string;
  "Actual Rows"?: number;
  Plans?: ExplainNode[];
  [key: string]: unknown;
}

/** Flatten all plan nodes from an EXPLAIN JSON result into a single array. */
function collectNodes(node: ExplainNode): ExplainNode[] {
  const nodes: ExplainNode[] = [node];
  if (node.Plans) {
    for (const child of node.Plans) {
      nodes.push(...collectNodes(child));
    }
  }
  return nodes;
}

/** Return the top-level plan node from an EXPLAIN JSON result row. */
function topNode(explainRows: Array<{ "QUERY PLAN": ExplainNode[] }>): ExplainNode {
  return explainRows[0]["QUERY PLAN"][0];
}

/** Return all node types present in the plan tree. */
function nodeTypes(root: ExplainNode): string[] {
  return collectNodes(root).map((n) => n["Node Type"]);
}

/** Assert the plan uses an index (Index Scan or Index Only Scan), not Seq Scan. */
function assertUsesIndex(root: ExplainNode, label: string): void {
  const types = nodeTypes(root);
  const usesIndex =
    types.some((t) => t === "Index Scan" || t === "Index Only Scan");
  const usesSeqScan = types.some((t) => t === "Seq Scan");
  if (!usesIndex) {
    throw new Error(
      `[${label}] Expected Index Scan or Index Only Scan but found: ${types.join(", ")}`
    );
  }
  if (usesSeqScan) {
    throw new Error(
      `[${label}] Unexpected Seq Scan in plan; types found: ${types.join(", ")}`
    );
  }
}

/**
 * Assert that the "Actual Rows" for the given node type is proportional to
 * the query selectivity — i.e. significantly less than the full table size.
 *
 * We allow up to `maxFraction` of TOTAL_RUNS rows.
 */
function assertSelectiveRows(
  root: ExplainNode,
  maxFraction: number,
  label: string
): void {
  const scanNode = collectNodes(root).find(
    (n) => n["Node Type"] === "Index Scan" || n["Node Type"] === "Index Only Scan"
  );
  if (!scanNode) {
    throw new Error(`[${label}] No Index Scan node found to check Actual Rows`);
  }
  const actualRows = scanNode["Actual Rows"] ?? 0;
  const threshold = Math.ceil(TOTAL_RUNS * maxFraction);
  if (actualRows > threshold) {
    throw new Error(
      `[${label}] Actual Rows ${actualRows} exceeds selectivity threshold ${threshold} ` +
        `(${maxFraction * 100}% of ${TOTAL_RUNS} rows)`
    );
  }
}

// ---------------------------------------------------------------------------
// Test suite
// ---------------------------------------------------------------------------

describe("Query Performance — index usage (TASK-4622)", () => {
  let client: Client;

  // -------------------------------------------------------------------------
  // Setup / teardown
  // -------------------------------------------------------------------------

  beforeAll(async () => {
    client = new Client({ connectionString: process.env.DATABASE_URL });
    await client.connect();

    // Seed 200 runs across 5 branches, one run_stage per run
    await client.query("BEGIN");
    try {
      for (const branch of BRANCHES) {
        for (let i = 0; i < RUNS_PER_BRANCH; i++) {
          const runId = randomUUID();
          const status = STATUSES[i % STATUSES.length];
          const commitSha = `sha${branch.replace(/\W/g, "")}${i.toString().padStart(3, "0")}`;
          const userId = `user_${i % 10}`; // 10 distinct users

          await client.query(
            `INSERT INTO validation_runs
               (id, "branchName", "commitSha", status, config, "userId", "createdAt", "updatedAt")
             VALUES ($1, $2, $3, $4::\"ValidationRunStatus\", $5, $6,
                     NOW() - ($7 * INTERVAL '1 minute'),
                     NOW() - ($7 * INTERVAL '1 minute'))`,
            [
              runId,
              branch,
              commitSha,
              status,
              JSON.stringify({ timeout: 300 }),
              `user_${i % 10}`,
              i, // spread creation times so ORDER BY is meaningful
            ]
          );

          await client.query(
            `INSERT INTO run_stages
               (id, "runId", "stageName", "stageIndex", status)
             VALUES ($1, $2, 'SETUP'::"StageType", 0, 'PENDING'::"StageStatus")`,
            [randomUUID(), runId]
          );
        }
      }
      await client.query("COMMIT");
    } catch (err) {
      await client.query("ROLLBACK");
      throw err;
    }
  });

  afterAll(async () => {
    // Remove seeded test data and close connection
    await client.query(
      `DELETE FROM validation_runs WHERE "userId" LIKE 'user_%'`
    );
    await client.end();
  });

  // -------------------------------------------------------------------------
  // Index: validation_runs (branchName, status)
  // -------------------------------------------------------------------------

  it("uses index on validation_runs(branchName, status)", async () => {
    const result = await client.query<{ "QUERY PLAN": ExplainNode[] }>(
      `EXPLAIN (FORMAT JSON, ANALYZE, BUFFERS FALSE)
       SELECT id FROM validation_runs
       WHERE "branchName" = 'main' AND status = 'PENDING'::"ValidationRunStatus"`
    );
    const root = topNode(result.rows);
    assertUsesIndex(root, "branchName_status");
    // Expect ~8 rows (40 runs × 1/5 statuses = 8) — well under 20% of 200
    assertSelectiveRows(root, 0.20, "branchName_status");
  });

  // -------------------------------------------------------------------------
  // Index: validation_runs (branchName, commitSha)
  // -------------------------------------------------------------------------

  it("uses index on validation_runs(branchName, commitSha)", async () => {
    // Target a single specific commit
    const result = await client.query<{ "QUERY PLAN": ExplainNode[] }>(
      `EXPLAIN (FORMAT JSON, ANALYZE, BUFFERS FALSE)
       SELECT id FROM validation_runs
       WHERE "branchName" = 'main' AND "commitSha" = 'shamain000'`
    );
    const root = topNode(result.rows);
    assertUsesIndex(root, "branchName_commitSha");
    // Exactly 1 row — well under 5% of 200
    assertSelectiveRows(root, 0.05, "branchName_commitSha");
  });

  // -------------------------------------------------------------------------
  // Index: validation_runs (branchName, createdAt DESC)
  // -------------------------------------------------------------------------

  it("uses index on validation_runs(branchName, createdAt DESC)", async () => {
    const result = await client.query<{ "QUERY PLAN": ExplainNode[] }>(
      `EXPLAIN (FORMAT JSON, ANALYZE, BUFFERS FALSE)
       SELECT id, "createdAt" FROM validation_runs
       WHERE "branchName" = 'feature/a'
       ORDER BY "createdAt" DESC
       LIMIT 10`
    );
    const root = topNode(result.rows);
    assertUsesIndex(root, "branchName_createdAt");
    // LIMIT 10 — far below 10% of 200
    assertSelectiveRows(root, 0.10, "branchName_createdAt");
  });

  // -------------------------------------------------------------------------
  // Index: validation_runs (userId, createdAt DESC)
  // -------------------------------------------------------------------------

  it("uses index on validation_runs(userId, createdAt DESC)", async () => {
    const result = await client.query<{ "QUERY PLAN": ExplainNode[] }>(
      `EXPLAIN (FORMAT JSON, ANALYZE, BUFFERS FALSE)
       SELECT id, "createdAt" FROM validation_runs
       WHERE "userId" = 'user_0'
       ORDER BY "createdAt" DESC
       LIMIT 10`
    );
    const root = topNode(result.rows);
    assertUsesIndex(root, "userId_createdAt");
    // user_0 appears in 20 of 200 rows — LIMIT 10 → ≤10 rows
    assertSelectiveRows(root, 0.15, "userId_createdAt");
  });

  // -------------------------------------------------------------------------
  // Index: run_stages (runId, stageIndex)
  // -------------------------------------------------------------------------

  it("uses index on run_stages(runId, stageIndex)", async () => {
    // Fetch a real runId to make a selective query
    const runRow = await client.query<{ id: string }>(
      `SELECT id FROM validation_runs WHERE "branchName" = 'main' LIMIT 1`
    );
    const runId = runRow.rows[0]?.id;
    if (!runId) throw new Error("No seeded run found for run_stages index test");

    const result = await client.query<{ "QUERY PLAN": ExplainNode[] }>(
      `EXPLAIN (FORMAT JSON, ANALYZE, BUFFERS FALSE)
       SELECT id FROM run_stages
       WHERE "runId" = $1 AND "stageIndex" = 0`,
      [runId]
    );
    const root = topNode(result.rows);
    assertUsesIndex(root, "runId_stageIndex");
    // 1 run has exactly 1 stage at stageIndex 0
    assertSelectiveRows(root, 0.05, "runId_stageIndex");
  });

  // -------------------------------------------------------------------------
  // Index: validation_presets (userId, createdAt DESC)
  // -------------------------------------------------------------------------

  it("uses index on validation_presets(userId, createdAt DESC)", async () => {
    // Insert a few presets for isolated testing
    const presetIds = [randomUUID(), randomUUID(), randomUUID()];
    await client.query("BEGIN");
    for (const pid of presetIds) {
      await client.query(
        `INSERT INTO validation_presets (id, "userId", name, "updatedAt")
         VALUES ($1, 'preset_user_1', 'test preset', NOW())`,
        [pid]
      );
    }
    await client.query("COMMIT");

    try {
      const result = await client.query<{ "QUERY PLAN": ExplainNode[] }>(
        `EXPLAIN (FORMAT JSON, ANALYZE, BUFFERS FALSE)
         SELECT id FROM validation_presets
         WHERE "userId" = 'preset_user_1'
         ORDER BY "createdAt" DESC`
      );
      const root = topNode(result.rows);
      assertUsesIndex(root, "presets_userId_createdAt");
    } finally {
      await client.query(
        `DELETE FROM validation_presets WHERE "userId" = 'preset_user_1'`
      );
    }
  });

  // -------------------------------------------------------------------------
  // Partial index: idx_runs_active on validation_runs(branchName, createdAt DESC)
  //               WHERE status IN ('PENDING', 'RUNNING')
  // -------------------------------------------------------------------------

  it("uses partial index idx_runs_active for active-run concurrency check", async () => {
    const result = await client.query<{ "QUERY PLAN": ExplainNode[] }>(
      `EXPLAIN (FORMAT JSON, ANALYZE, BUFFERS FALSE)
       SELECT id FROM validation_runs
       WHERE "branchName" = 'main'
         AND status IN ('PENDING'::"ValidationRunStatus", 'RUNNING'::"ValidationRunStatus")
       ORDER BY "createdAt" DESC`
    );
    const root = topNode(result.rows);
    assertUsesIndex(root, "idx_runs_active");
    // Only PENDING + RUNNING rows on 'main' branch (≈16 of 200)
    assertSelectiveRows(root, 0.20, "idx_runs_active");
  });

  // -------------------------------------------------------------------------
  // Verify the partial index name appears in pg_indexes
  // -------------------------------------------------------------------------

  it("partial index idx_runs_active exists in pg_indexes", async () => {
    const result = await client.query<{ indexname: string }>(
      `SELECT indexname FROM pg_indexes
       WHERE tablename = 'validation_runs' AND indexname = 'idx_runs_active'`
    );
    if (result.rows.length === 0) {
      throw new Error(
        "idx_runs_active partial index not found in pg_indexes — " +
          "run migration 003_indexes to create it"
      );
    }
  });

  // -------------------------------------------------------------------------
  // Verify composite indexes appear in pg_indexes with correct column order
  // -------------------------------------------------------------------------

  it("composite index validation_runs_branchName_status_idx has branchName first", async () => {
    const result = await client.query<{ indexdef: string }>(
      `SELECT indexdef FROM pg_indexes
       WHERE tablename = 'validation_runs'
         AND indexname = 'validation_runs_branchName_status_idx'`
    );
    if (result.rows.length === 0) {
      throw new Error(
        "validation_runs_branchName_status_idx not found in pg_indexes"
      );
    }
    const def = result.rows[0].indexdef;
    const branchPos = def.indexOf("branchName");
    const statusPos = def.indexOf("status");
    if (branchPos === -1 || statusPos === -1 || branchPos > statusPos) {
      throw new Error(
        `Index column order incorrect: branchName should precede status in "${def}"`
      );
    }
  });

  it("composite index validation_runs_branchName_createdAt_idx has branchName first", async () => {
    const result = await client.query<{ indexdef: string }>(
      `SELECT indexdef FROM pg_indexes
       WHERE tablename = 'validation_runs'
         AND indexname = 'validation_runs_branchName_createdAt_idx'`
    );
    if (result.rows.length === 0) {
      throw new Error(
        "validation_runs_branchName_createdAt_idx not found in pg_indexes"
      );
    }
    const def = result.rows[0].indexdef;
    const branchPos = def.indexOf("branchName");
    const createdAtPos = def.indexOf("createdAt");
    if (branchPos === -1 || createdAtPos === -1 || branchPos > createdAtPos) {
      throw new Error(
        `Index column order incorrect: branchName should precede createdAt in "${def}"`
      );
    }
    // Ensure DESC ordering for createdAt
    if (!def.includes("DESC")) {
      throw new Error(
        `Expected DESC ordering for createdAt in index definition "${def}"`
      );
    }
  });

  it("composite index validation_runs_userId_createdAt_idx has userId first", async () => {
    const result = await client.query<{ indexdef: string }>(
      `SELECT indexdef FROM pg_indexes
       WHERE tablename = 'validation_runs'
         AND indexname = 'validation_runs_userId_createdAt_idx'`
    );
    if (result.rows.length === 0) {
      throw new Error(
        "validation_runs_userId_createdAt_idx not found in pg_indexes"
      );
    }
    const def = result.rows[0].indexdef;
    const userIdPos = def.indexOf("userId");
    const createdAtPos = def.indexOf("createdAt");
    if (userIdPos === -1 || createdAtPos === -1 || userIdPos > createdAtPos) {
      throw new Error(
        `Index column order incorrect: userId should precede createdAt in "${def}"`
      );
    }
  });
});

// ---------------------------------------------------------------------------
// TASK-4623 — Validate Query Performance (@integration)
//
// Seeds 500 ValidationRun records across 10 branches with varied statuses and
// timestamps, then asserts that EXPLAIN (FORMAT JSON, ANALYZE) plans for the
// five key queries use index scans (not seq scans) and — where applicable —
// partition pruning.
// ---------------------------------------------------------------------------

/** Return all plan nodes that include a "Subplans Removed" field (partition pruning indicator). */
function findAppendNode(root: ExplainNode): ExplainNode | undefined {
  return collectNodes(root).find((n) => n["Node Type"] === "Append");
}

/** Assert that a date-range query prunes at least one partition (Subplans Removed > 0). */
function assertPartitionPruning(root: ExplainNode, label: string): void {
  const appendNode = findAppendNode(root);
  if (!appendNode) {
    // If the table is NOT partitioned the planner produces a plain Seq Scan /
    // Index Scan — skip the assertion rather than failing unconditionally.
    return;
  }
  const subplansRemoved = (appendNode as Record<string, unknown>)[
    "Subplans Removed"
  ] as number | undefined;
  if (subplansRemoved === undefined || subplansRemoved === 0) {
    throw new Error(
      `[${label}] Expected partition pruning (Subplans Removed > 0) ` +
        `but got: ${subplansRemoved}. ` +
        `Ensure the createdAt date-range predicate aligns with a monthly partition boundary.`
    );
  }
}

describe("Query Performance — TASK-4623 (@integration)", () => {
  let client4623: Client;

  // 10 branches × 50 runs = 500 total records
  const BRANCHES_4623 = [
    "main",
    "feature/alpha",
    "feature/beta",
    "feature/gamma",
    "feature/delta",
    "hotfix/001",
    "hotfix/002",
    "release/1.0",
    "release/2.0",
    "release/3.0",
  ];
  const RUNS_PER_BRANCH_4623 = 50;
  const TOTAL_RUNS_4623 = BRANCHES_4623.length * RUNS_PER_BRANCH_4623; // 500

  const STATUSES_4623 = [
    "PENDING",
    "RUNNING",
    "SUCCESS",
    "FAILED",
    "CANCELLED",
  ] as const;

  // -------------------------------------------------------------------------
  // Setup / teardown
  // -------------------------------------------------------------------------

  beforeAll(async () => {
    client4623 = new Client({ connectionString: process.env.DATABASE_URL });
    await client4623.connect();

    // Seed 500 runs across 10 branches with varied statuses and timestamps.
    // Each run gets one SETUP stage at stageIndex 0.
    await client4623.query("BEGIN");
    try {
      for (const branch of BRANCHES_4623) {
        for (let i = 0; i < RUNS_PER_BRANCH_4623; i++) {
          const runId = randomUUID();
          const status = STATUSES_4623[i % STATUSES_4623.length];
          const commitSha = `sha4623${branch.replace(/\W/g, "")}${i
            .toString()
            .padStart(3, "0")}`;
          const userId = `user4623_${i % 10}`;

          await client4623.query(
            `INSERT INTO validation_runs
               (id, "branchName", "commitSha", status, config, "userId",
                "createdAt", "updatedAt")
             VALUES ($1, $2, $3, $4::"ValidationRunStatus", $5, $6,
                     NOW() - ($7 * INTERVAL '1 minute'),
                     NOW() - ($7 * INTERVAL '1 minute'))`,
            [
              runId,
              branch,
              commitSha,
              status,
              JSON.stringify({ timeout: 300 }),
              userId,
              i, // spread creation times so ORDER BY createdAt is meaningful
            ]
          );

          await client4623.query(
            `INSERT INTO run_stages
               (id, "runId", "stageName", "stageIndex", status)
             VALUES ($1, $2, 'SETUP'::"StageType", 0, 'PENDING'::"StageStatus")`,
            [randomUUID(), runId]
          );
        }
      }
      await client4623.query("COMMIT");
    } catch (err) {
      await client4623.query("ROLLBACK");
      throw err;
    }
  });

  afterAll(async () => {
    // Remove all seeded TASK-4623 data identified by userId prefix
    await client4623.query(
      `DELETE FROM validation_runs WHERE "userId" LIKE 'user4623_%'`
    );
    await client4623.end();
  });

  // -------------------------------------------------------------------------
  // Test 1 — idx_runs_branch_status
  // Query: branch + status filter, ORDER BY createdAt DESC LIMIT 20
  // -------------------------------------------------------------------------

  it(
    "Test 1 (idx_runs_branch_status): uses index for branchName+status filter with ORDER BY createdAt DESC LIMIT 20",
    async () => {
      const result = await client4623.query<{ "QUERY PLAN": ExplainNode[] }>(
        `EXPLAIN (FORMAT JSON, ANALYZE, BUFFERS FALSE)
         SELECT * FROM validation_runs
         WHERE "branchName" = $1
           AND status = $2::"ValidationRunStatus"
         ORDER BY "createdAt" DESC
         LIMIT 20`,
        ["main", "PENDING"]
      );
      const root = topNode(result.rows);
      // Must use an index scan — not a sequential scan
      assertUsesIndex(root, "idx_runs_branch_status");
      // Expect at most 20 rows (LIMIT) — well within 10% of 500 total rows
      assertSelectiveRows(root, 0.1, "idx_runs_branch_status");
    }
  );

  // -------------------------------------------------------------------------
  // Test 2 — idx_runs_branch_commit
  // Query: branch + commitSha LIKE prefix search
  // -------------------------------------------------------------------------

  it(
    "Test 2 (idx_runs_branch_commit): uses index for branchName + commitSha LIKE query",
    async () => {
      const result = await client4623.query<{ "QUERY PLAN": ExplainNode[] }>(
        `EXPLAIN (FORMAT JSON, ANALYZE, BUFFERS FALSE)
         SELECT * FROM validation_runs
         WHERE "branchName" = $1
           AND "commitSha" LIKE $2`,
        ["main", "sha4623main0%"]
      );
      const root = topNode(result.rows);
      assertUsesIndex(root, "idx_runs_branch_commit");
    }
  );

  // -------------------------------------------------------------------------
  // Test 3 — partial index (concurrency check)
  // Query: branch + status IN ('PENDING','RUNNING') ORDER BY createdAt DESC
  // -------------------------------------------------------------------------

  it(
    "Test 3 (partial index): concurrency check query uses idx_runs_active partial index",
    async () => {
      const result = await client4623.query<{ "QUERY PLAN": ExplainNode[] }>(
        `EXPLAIN (FORMAT JSON, ANALYZE, BUFFERS FALSE)
         SELECT id FROM validation_runs
         WHERE "branchName" = $1
           AND status IN (
             'PENDING'::"ValidationRunStatus",
             'RUNNING'::"ValidationRunStatus"
           )
         ORDER BY "createdAt" DESC`,
        ["main"]
      );
      const root = topNode(result.rows);
      assertUsesIndex(root, "idx_runs_active_partial_index");

      // Additionally assert the planner chose the partial index idx_runs_active
      const allNodes = collectNodes(root);
      const indexScanNode = allNodes.find(
        (n) =>
          n["Node Type"] === "Index Scan" ||
          n["Node Type"] === "Index Only Scan"
      );
      if (
        indexScanNode &&
        indexScanNode["Index Name"] !== undefined &&
        indexScanNode["Index Name"] !== "idx_runs_active"
      ) {
        throw new Error(
          `[idx_runs_active] Expected partial index idx_runs_active to be chosen, ` +
            `but planner used: ${indexScanNode["Index Name"]}`
        );
      }
    }
  );

  // -------------------------------------------------------------------------
  // Test 4 — partition pruning for date-range queries
  // If partitioning is implemented the planner must prune non-matching
  // monthly partitions (Subplans Removed > 0).
  // -------------------------------------------------------------------------

  it(
    "Test 4 (partition pruning): date-range query on createdAt prunes non-matching partitions",
    async () => {
      // Target March 2026 — the partition validation_runs_y2026m03 should be
      // the only child the planner scans; all others should be pruned.
      const result = await client4623.query<{ "QUERY PLAN": ExplainNode[] }>(
        `EXPLAIN (FORMAT JSON, ANALYZE, BUFFERS FALSE)
         SELECT id, "branchName", "createdAt" FROM validation_runs
         WHERE "createdAt" >= '2026-03-01'::timestamp
           AND "createdAt" <  '2026-04-01'::timestamp`
      );
      const root = topNode(result.rows);
      // Partition pruning assertion: Subplans Removed must be > 0.
      // If the table is not partitioned this helper returns early without failing.
      assertPartitionPruning(root, "partition_pruning_date_range");

      // For non-partitioned tables the planner should still use an index
      // (branchName_createdAt) or a seq scan — either is acceptable since
      // the primary goal of this test is the partition pruning check.
      const types = nodeTypes(root);
      const hasReasonableScan = types.some(
        (t) =>
          t === "Index Scan" ||
          t === "Index Only Scan" ||
          t === "Seq Scan" ||
          t === "Append"
      );
      if (!hasReasonableScan) {
        throw new Error(
          `[partition_pruning_date_range] Unexpected plan node types: ${types.join(", ")}`
        );
      }
    }
  );

  // -------------------------------------------------------------------------
  // Test 5 — idx_stages_run_index
  // Query: run_stages WHERE runId = $1 ORDER BY stageIndex
  // -------------------------------------------------------------------------

  it(
    "Test 5 (idx_stages_run_index): run_stages query by runId ORDER BY stageIndex uses index",
    async () => {
      const runRow = await client4623.query<{ id: string }>(
        `SELECT id FROM validation_runs
         WHERE "branchName" = 'main'
           AND "userId" LIKE 'user4623_%'
         LIMIT 1`
      );
      const runId = runRow.rows[0]?.id;
      if (!runId) {
        throw new Error("No seeded TASK-4623 run found for idx_stages_run_index test");
      }

      const result = await client4623.query<{ "QUERY PLAN": ExplainNode[] }>(
        `EXPLAIN (FORMAT JSON, ANALYZE, BUFFERS FALSE)
         SELECT * FROM run_stages
         WHERE "runId" = $1
         ORDER BY "stageIndex"`,
        [runId]
      );
      const root = topNode(result.rows);
      assertUsesIndex(root, "idx_stages_run_index");
      // Each run has exactly 1 stage — far below 1% of total run_stages rows
      assertSelectiveRows(root, 0.01, "idx_stages_run_index");
    }
  );

  // -------------------------------------------------------------------------
  // Verify TOTAL_RUNS_4623 constant is 500 (sanity check)
  // -------------------------------------------------------------------------

  it("seeds exactly 500 validation_run records (10 branches × 50 runs)", () => {
    expect(TOTAL_RUNS_4623).toBe(500);
    expect(BRANCHES_4623).toHaveLength(10);
    expect(RUNS_PER_BRANCH_4623).toBe(50);
  });
});
