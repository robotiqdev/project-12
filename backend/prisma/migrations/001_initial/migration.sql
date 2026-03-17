-- CreateEnum
CREATE TYPE "ValidationRunStatus" AS ENUM ('PENDING', 'RUNNING', 'SUCCESS', 'FAILED', 'CANCELLED');

-- CreateEnum
CREATE TYPE "StageStatus" AS ENUM ('PENDING', 'RUNNING', 'SUCCESS', 'FAILED', 'CANCELLED');

-- CreateEnum
CREATE TYPE "StageType" AS ENUM ('SETUP', 'LINT', 'TEST', 'BUILD', 'DEPLOY');

-- CreateEnum
CREATE TYPE "CleanupJobType" AS ENUM ('AGE_BASED', 'COUNT_BASED');

-- CreateEnum
CREATE TYPE "CleanupJobStatus" AS ENUM ('PENDING', 'RUNNING', 'DONE', 'FAILED');

-- CreateTable
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

    CONSTRAINT "validation_runs_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "run_stages" (
    "id"          TEXT NOT NULL,
    "runId"       TEXT NOT NULL,
    "stageName"   "StageType" NOT NULL,
    "stageIndex"  INTEGER NOT NULL,
    "status"      "StageStatus" NOT NULL DEFAULT 'PENDING',
    "logs"        TEXT NOT NULL DEFAULT '',
    "startedAt"   TIMESTAMP(3),
    "completedAt" TIMESTAMP(3),

    CONSTRAINT "run_stages_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "validation_presets" (
    "id"           TEXT NOT NULL,
    "userId"       TEXT NOT NULL,
    "name"         TEXT NOT NULL,
    "branchName"   TEXT,
    "commitSha"    TEXT,
    "envVars"      JSONB NOT NULL DEFAULT '{}',
    "featureFlags" JSONB NOT NULL DEFAULT '{}',
    "createdAt"    TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt"    TIMESTAMP(3) NOT NULL,

    CONSTRAINT "validation_presets_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "cleanup_jobs" (
    "id"           TEXT NOT NULL,
    "type"         "CleanupJobType" NOT NULL,
    "status"       "CleanupJobStatus" NOT NULL DEFAULT 'PENDING',
    "deletedCount" INTEGER NOT NULL DEFAULT 0,
    "startedAt"    TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "completedAt"  TIMESTAMP(3),
    "errorMessage" TEXT,

    CONSTRAINT "cleanup_jobs_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE INDEX "validation_runs_branchName_status_idx" ON "validation_runs"("branchName", "status");

-- CreateIndex
CREATE INDEX "validation_runs_branchName_commitSha_idx" ON "validation_runs"("branchName", "commitSha");

-- CreateIndex
CREATE INDEX "validation_runs_branchName_createdAt_idx" ON "validation_runs"("branchName", "createdAt" DESC);

-- CreateIndex
CREATE INDEX "validation_runs_userId_createdAt_idx" ON "validation_runs"("userId", "createdAt" DESC);

-- CreateIndex
CREATE UNIQUE INDEX "run_stages_runId_stageIndex_key" ON "run_stages"("runId", "stageIndex");

-- CreateIndex
CREATE INDEX "run_stages_runId_stageIndex_idx" ON "run_stages"("runId", "stageIndex");

-- CreateIndex
CREATE INDEX "validation_presets_userId_createdAt_idx" ON "validation_presets"("userId", "createdAt" DESC);

-- AddForeignKey
ALTER TABLE "run_stages" ADD CONSTRAINT "run_stages_runId_fkey" FOREIGN KEY ("runId") REFERENCES "validation_runs"("id") ON DELETE CASCADE ON UPDATE CASCADE;
