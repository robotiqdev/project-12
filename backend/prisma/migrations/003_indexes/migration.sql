-- Migration: 003_indexes
-- Performance indexes for TASK-4622

-- CreateIndex: validation_runs(branchName, status)
CREATE INDEX "validation_runs_branchName_status_idx" ON "validation_runs"("branchName", "status");

-- CreateIndex: validation_runs(branchName, commitSha)
CREATE INDEX "validation_runs_branchName_commitSha_idx" ON "validation_runs"("branchName", "commitSha");

-- CreateIndex: validation_runs(branchName, createdAt DESC)
CREATE INDEX "validation_runs_branchName_createdAt_idx" ON "validation_runs"("branchName", "createdAt" DESC);

-- CreateIndex: validation_runs(userId, createdAt DESC)
CREATE INDEX "validation_runs_userId_createdAt_idx" ON "validation_runs"("userId", "createdAt" DESC);

-- CreateIndex: run_stages(runId, stageIndex)
CREATE INDEX "run_stages_runId_stageIndex_idx" ON "run_stages"("runId", "stageIndex");

-- CreateUniqueIndex: run_stages(runId, stageIndex)
CREATE UNIQUE INDEX "run_stages_runId_stageIndex_key" ON "run_stages"("runId", "stageIndex");

-- CreateIndex: validation_presets(userId, createdAt DESC)
CREATE INDEX "validation_presets_userId_createdAt_idx" ON "validation_presets"("userId", "createdAt" DESC);

-- Partial index for active runs (speeds up concurrency checks)
CREATE INDEX "idx_runs_active" ON "validation_runs" ("branchName", "createdAt" DESC) WHERE status IN ('PENDING', 'RUNNING');
