-- Migration: 002_partitioning
-- Convert validation_runs to a range-partitioned table by createdAt (TASK-4621)
--
-- TODO: Monthly partitions beyond 2026 must be pre-created by a scheduled admin
-- task (e.g. a cron job) before the start of each new month to prevent new rows
-- from falling into the default partition.

-- Step 1: Preserve existing data
CREATE TABLE validation_runs_old AS SELECT * FROM validation_runs;

-- Step 2: Drop existing foreign keys and dependent objects
ALTER TABLE run_stages DROP CONSTRAINT IF EXISTS "run_stages_runId_fkey";

-- Step 3: Drop the original table
DROP TABLE validation_runs;

-- Step 4: Re-create validation_runs as a partitioned table (PARTITION BY RANGE on createdAt)
CREATE TABLE "validation_runs" (
    "id"          TEXT NOT NULL,
    "branchName"  TEXT NOT NULL,
    "commitSha"   TEXT NOT NULL,
    "status"      "ValidationRunStatus" NOT NULL DEFAULT 'PENDING',
    "config"      JSONB NOT NULL,
    "userId"      TEXT NOT NULL,
    "createdAt"   TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt"   TIMESTAMP(3) NOT NULL,
    "completedAt" TIMESTAMP(3),
    "cancelledAt" TIMESTAMP(3),

    CONSTRAINT "validation_runs_pkey" PRIMARY KEY ("id", "createdAt")
) PARTITION BY RANGE ("createdAt");

-- Step 5: Create monthly partitions for 2026
CREATE TABLE "validation_runs_y2026m01" PARTITION OF "validation_runs"
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

CREATE TABLE "validation_runs_y2026m02" PARTITION OF "validation_runs"
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');

CREATE TABLE "validation_runs_y2026m03" PARTITION OF "validation_runs"
    FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');

CREATE TABLE "validation_runs_y2026m04" PARTITION OF "validation_runs"
    FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');

CREATE TABLE "validation_runs_y2026m05" PARTITION OF "validation_runs"
    FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');

CREATE TABLE "validation_runs_y2026m06" PARTITION OF "validation_runs"
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');

CREATE TABLE "validation_runs_y2026m07" PARTITION OF "validation_runs"
    FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');

CREATE TABLE "validation_runs_y2026m08" PARTITION OF "validation_runs"
    FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');

CREATE TABLE "validation_runs_y2026m09" PARTITION OF "validation_runs"
    FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');

CREATE TABLE "validation_runs_y2026m10" PARTITION OF "validation_runs"
    FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');

CREATE TABLE "validation_runs_y2026m11" PARTITION OF "validation_runs"
    FOR VALUES FROM ('2026-11-01') TO ('2026-12-01');

CREATE TABLE "validation_runs_y2026m12" PARTITION OF "validation_runs"
    FOR VALUES FROM ('2026-12-01') TO ('2027-01-01');

-- Step 6: Create default catch-all partition for rows outside defined ranges
CREATE TABLE "validation_runs_default" PARTITION OF "validation_runs" DEFAULT;

-- Step 7: Restore data from old table
INSERT INTO validation_runs SELECT * FROM validation_runs_old;
DROP TABLE validation_runs_old;

-- Step 8: Re-add foreign key constraint
ALTER TABLE run_stages ADD CONSTRAINT "run_stages_runId_fkey"
    FOREIGN KEY ("runId") REFERENCES "validation_runs"("id") ON DELETE CASCADE ON UPDATE CASCADE;
