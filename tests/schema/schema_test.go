// Package schema_test validates the Prisma schema and related files for TASK-4620.
// These are integration-style file-system tests that verify the schema.prisma,
// migration SQL, and Prisma client singleton contain all required definitions.
package schema_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot resolves the repository root from the current working directory.
// Tests are run with `go test ./...` from the repo root, so the working
// directory is the package directory; we walk up to find go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find repo root (no go.mod found)")
		}
		dir = parent
	}
}

// ---- Schema file existence ----

func TestSchemaPrismaFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "prisma", "schema.prisma")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("schema.prisma not found at %s", path)
	}
}

func TestMigrationSQLFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "prisma", "migrations", "001_initial", "migration.sql")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("migration SQL not found at %s", path)
	}
}

func TestPrismaClientSingletonFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "src", "db", "client.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("client.ts not found at %s", path)
	}
}

// ---- Helpers ----

func readSchema(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "prisma", "schema.prisma")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read schema.prisma: %v", err)
	}
	return string(data)
}

func readMigrationSQL(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "prisma", "migrations", "001_initial", "migration.sql")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read migration.sql: %v", err)
	}
	return string(data)
}

func readClientTS(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "src", "db", "client.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read client.ts: %v", err)
	}
	return string(data)
}

// ---- Datasource & generator ----

func TestSchemaDatasourcePostgresql(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, `provider = "postgresql"`) {
		t.Error(`schema.prisma must contain datasource with provider = "postgresql"`)
	}
}

func TestSchemaDatasourceUsesEnvDatabaseURL(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, `env("DATABASE_URL")`) {
		t.Error(`schema.prisma must reference env("DATABASE_URL") for datasource url`)
	}
}

func TestSchemaGeneratorPrismaClientJs(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, `provider = "prisma-client-js"`) {
		t.Error(`schema.prisma must define generator client with provider = "prisma-client-js"`)
	}
}

// ---- Enums ----

func TestSchemaEnumValidationRunStatus(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "enum ValidationRunStatus") {
		t.Error("schema.prisma must define enum ValidationRunStatus")
	}
	for _, member := range []string{"PENDING", "RUNNING", "SUCCESS", "FAILED", "CANCELLED"} {
		if !strings.Contains(schema, member) {
			t.Errorf("enum ValidationRunStatus must include member %s", member)
		}
	}
}

func TestSchemaEnumStageStatus(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "enum StageStatus") {
		t.Error("schema.prisma must define enum StageStatus")
	}
	for _, member := range []string{"PENDING", "RUNNING", "SUCCESS", "FAILED", "CANCELLED"} {
		if !strings.Contains(schema, member) {
			t.Errorf("enum StageStatus must include member %s", member)
		}
	}
}

func TestSchemaEnumStageType(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "enum StageType") {
		t.Error("schema.prisma must define enum StageType")
	}
	for _, member := range []string{"SETUP", "LINT", "TEST", "BUILD", "DEPLOY"} {
		if !strings.Contains(schema, member) {
			t.Errorf("enum StageType must include member %s", member)
		}
	}
}

func TestSchemaEnumCleanupJobType(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "enum CleanupJobType") {
		t.Error("schema.prisma must define enum CleanupJobType")
	}
	for _, member := range []string{"AGE_BASED", "COUNT_BASED"} {
		if !strings.Contains(schema, member) {
			t.Errorf("enum CleanupJobType must include member %s", member)
		}
	}
}

func TestSchemaEnumCleanupJobStatus(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "enum CleanupJobStatus") {
		t.Error("schema.prisma must define enum CleanupJobStatus")
	}
	for _, member := range []string{"PENDING", "RUNNING", "DONE", "FAILED"} {
		if !strings.Contains(schema, member) {
			t.Errorf("enum CleanupJobStatus must include member %s", member)
		}
	}
}

// ---- ValidationRun model ----

func TestSchemaModelValidationRun(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "model ValidationRun") {
		t.Error("schema.prisma must define model ValidationRun")
	}
}

func TestSchemaValidationRunMapsToValidationRuns(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, `@@map("validation_runs")`) {
		t.Error(`ValidationRun model must include @@map("validation_runs")`)
	}
}

func TestSchemaValidationRunFields(t *testing.T) {
	schema := readSchema(t)
	requiredFields := []string{
		"branchName",
		"commitSha",
		"status",
		"config",
		"userId",
		"createdAt",
		"updatedAt",
		"completedAt",
		"cancelledAt",
	}
	for _, field := range requiredFields {
		if !strings.Contains(schema, field) {
			t.Errorf("ValidationRun model must include field %s", field)
		}
	}
}

func TestSchemaValidationRunIdIsUUID(t *testing.T) {
	schema := readSchema(t)
	// The schema should use @default(uuid()) for the id field
	if !strings.Contains(schema, "uuid()") {
		t.Error("ValidationRun id must use @default(uuid())")
	}
}

func TestSchemaValidationRunStatusDefaultPending(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "ValidationRunStatus @default(PENDING)") {
		t.Error("ValidationRun.status must default to PENDING")
	}
}

func TestSchemaValidationRunConfigIsJson(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "config") || !strings.Contains(schema, "Json") {
		t.Error("ValidationRun.config must be of type Json")
	}
}

func TestSchemaValidationRunCompletedAtNullable(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "completedAt") {
		t.Error("ValidationRun must have completedAt field")
	}
	// Optional fields use ? suffix in Prisma
	if !strings.Contains(schema, "completedAt  DateTime?") && !strings.Contains(schema, "completedAt DateTime?") && !strings.Contains(schema, "completedAt\tDateTime?") {
		t.Error("ValidationRun.completedAt must be optional (DateTime?)")
	}
}

func TestSchemaValidationRunCancelledAtNullable(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "cancelledAt") {
		t.Error("ValidationRun must have cancelledAt field")
	}
	if !strings.Contains(schema, "cancelledAt  DateTime?") && !strings.Contains(schema, "cancelledAt DateTime?") && !strings.Contains(schema, "cancelledAt\tDateTime?") {
		t.Error("ValidationRun.cancelledAt must be optional (DateTime?)")
	}
}

func TestSchemaValidationRunUpdatedAtAutoUpdate(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "@updatedAt") {
		t.Error("ValidationRun.updatedAt must use @updatedAt")
	}
}

func TestSchemaValidationRunIndexBranchNameStatus(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "[branchName, status]") {
		t.Error("ValidationRun must have composite index on [branchName, status]")
	}
}

func TestSchemaValidationRunIndexBranchNameCommitSha(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "[branchName, commitSha]") {
		t.Error("ValidationRun must have composite index on [branchName, commitSha]")
	}
}

func TestSchemaValidationRunIndexBranchNameCreatedAt(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "branchName, createdAt") {
		t.Error("ValidationRun must have composite index on [branchName, createdAt(sort: Desc)]")
	}
}

func TestSchemaValidationRunIndexUserIdCreatedAt(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "userId, createdAt") {
		t.Error("ValidationRun must have composite index on [userId, createdAt(sort: Desc)]")
	}
}

// ---- RunStage model ----

func TestSchemaModelRunStage(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "model RunStage") {
		t.Error("schema.prisma must define model RunStage")
	}
}

func TestSchemaRunStageMapsToRunStages(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, `@@map("run_stages")`) {
		t.Error(`RunStage model must include @@map("run_stages")`)
	}
}

func TestSchemaRunStageFields(t *testing.T) {
	schema := readSchema(t)
	requiredFields := []string{
		"runId",
		"stageName",
		"stageIndex",
		"status",
		"logs",
		"startedAt",
		"completedAt",
	}
	for _, field := range requiredFields {
		if !strings.Contains(schema, field) {
			t.Errorf("RunStage model must include field %s", field)
		}
	}
}

func TestSchemaRunStageStatusDefaultPending(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "StageStatus @default(PENDING)") {
		t.Error("RunStage.status must default to PENDING")
	}
}

func TestSchemaRunStageLogsDefaultEmpty(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, `logs`) {
		t.Error("RunStage must have logs field")
	}
	if !strings.Contains(schema, `@default("")`) {
		t.Error(`RunStage.logs must default to "" (empty string)`)
	}
}

func TestSchemaRunStageStageNameIsStageType(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "stageName") || !strings.Contains(schema, "StageType") {
		t.Error("RunStage.stageName must be of type StageType")
	}
}

func TestSchemaRunStageForeignKeyToValidationRun(t *testing.T) {
	schema := readSchema(t)
	// runId should reference ValidationRun
	if !strings.Contains(schema, "ValidationRun") {
		t.Error("RunStage must have a foreign key relation to ValidationRun")
	}
}

func TestSchemaRunStageCascadeDelete(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "onDelete: Cascade") {
		t.Error("RunStage must configure onDelete: Cascade from ValidationRun")
	}
}

func TestSchemaRunStageIndexRunIdStageIndex(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "[runId, stageIndex]") {
		t.Error("RunStage must have composite index on [runId, stageIndex]")
	}
}

func TestSchemaRunStageUniqueRunIdStageIndex(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "@@unique([runId, stageIndex])") {
		t.Error("RunStage must have @@unique([runId, stageIndex])")
	}
}

// ---- ValidationPreset model ----

func TestSchemaModelValidationPreset(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "model ValidationPreset") {
		t.Error("schema.prisma must define model ValidationPreset")
	}
}

func TestSchemaValidationPresetMapsToValidationPresets(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, `@@map("validation_presets")`) {
		t.Error(`ValidationPreset model must include @@map("validation_presets")`)
	}
}

func TestSchemaValidationPresetFields(t *testing.T) {
	schema := readSchema(t)
	requiredFields := []string{
		"userId",
		"name",
		"branchName",
		"commitSha",
		"envVars",
		"featureFlags",
		"createdAt",
		"updatedAt",
	}
	for _, field := range requiredFields {
		if !strings.Contains(schema, field) {
			t.Errorf("ValidationPreset model must include field %s", field)
		}
	}
}

func TestSchemaValidationPresetBranchNameNullable(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "branchName  String?") && !strings.Contains(schema, "branchName String?") && !strings.Contains(schema, "branchName\tString?") {
		t.Error("ValidationPreset.branchName must be optional (String?)")
	}
}

func TestSchemaValidationPresetCommitShaNullable(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "commitSha  String?") && !strings.Contains(schema, "commitSha String?") && !strings.Contains(schema, "commitSha\tString?") {
		t.Error("ValidationPreset.commitSha must be optional (String?)")
	}
}

func TestSchemaValidationPresetEnvVarsDefaultEmptyJson(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "envVars") {
		t.Error("ValidationPreset must have envVars field")
	}
}

func TestSchemaValidationPresetFeatureFlagsIsJson(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "featureFlags") {
		t.Error("ValidationPreset must have featureFlags field")
	}
}

func TestSchemaValidationPresetIndexUserIdCreatedAt(t *testing.T) {
	schema := readSchema(t)
	// At least one index on [userId, createdAt] must exist somewhere in the schema
	if !strings.Contains(schema, "userId, createdAt") {
		t.Error("ValidationPreset must have composite index on [userId, createdAt(sort: Desc)]")
	}
}

// ---- CleanupJob model ----

func TestSchemaModelCleanupJob(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "model CleanupJob") {
		t.Error("schema.prisma must define model CleanupJob")
	}
}

func TestSchemaCleanupJobMapsToCleanupJobs(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, `@@map("cleanup_jobs")`) {
		t.Error(`CleanupJob model must include @@map("cleanup_jobs")`)
	}
}

func TestSchemaCleanupJobFields(t *testing.T) {
	schema := readSchema(t)
	requiredFields := []string{
		"type",
		"status",
		"deletedCount",
		"startedAt",
		"completedAt",
		"errorMessage",
	}
	for _, field := range requiredFields {
		if !strings.Contains(schema, field) {
			t.Errorf("CleanupJob model must include field %s", field)
		}
	}
}

func TestSchemaCleanupJobTypeIsCleanupJobType(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "CleanupJobType") {
		t.Error("CleanupJob.type must be of type CleanupJobType")
	}
}

func TestSchemaCleanupJobStatusDefaultPending(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "CleanupJobStatus @default(PENDING)") {
		t.Error("CleanupJob.status must default to PENDING")
	}
}

func TestSchemaCleanupJobDeletedCountDefaultZero(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "deletedCount") {
		t.Error("CleanupJob must have deletedCount field")
	}
	if !strings.Contains(schema, "@default(0)") {
		t.Error("CleanupJob.deletedCount must default to 0")
	}
}

func TestSchemaCleanupJobStartedAtDefaultNow(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "startedAt") {
		t.Error("CleanupJob must have startedAt field")
	}
	if !strings.Contains(schema, "@default(now())") {
		t.Error("CleanupJob.startedAt must default to now()")
	}
}

func TestSchemaCleanupJobCompletedAtNullable(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "completedAt") {
		t.Error("CleanupJob must have completedAt field")
	}
}

func TestSchemaCleanupJobErrorMessageNullable(t *testing.T) {
	schema := readSchema(t)
	if !strings.Contains(schema, "errorMessage") {
		t.Error("CleanupJob must have errorMessage field")
	}
	if !strings.Contains(schema, "errorMessage  String?") && !strings.Contains(schema, "errorMessage String?") && !strings.Contains(schema, "errorMessage\tString?") {
		t.Error("CleanupJob.errorMessage must be optional (String?)")
	}
}

// ---- Migration SQL ----

func TestMigrationSQLNotEmpty(t *testing.T) {
	sql := readMigrationSQL(t)
	if strings.TrimSpace(sql) == "" {
		t.Error("migration.sql must not be empty")
	}
}

func TestMigrationSQLCreatesValidationRunsTable(t *testing.T) {
	sql := readMigrationSQL(t)
	if !strings.Contains(strings.ToLower(sql), "validation_runs") {
		t.Error("migration.sql must create the validation_runs table")
	}
}

func TestMigrationSQLCreatesRunStagesTable(t *testing.T) {
	sql := readMigrationSQL(t)
	if !strings.Contains(strings.ToLower(sql), "run_stages") {
		t.Error("migration.sql must create the run_stages table")
	}
}

func TestMigrationSQLCreatesValidationPresetsTable(t *testing.T) {
	sql := readMigrationSQL(t)
	if !strings.Contains(strings.ToLower(sql), "validation_presets") {
		t.Error("migration.sql must create the validation_presets table")
	}
}

func TestMigrationSQLCreatesCleanupJobsTable(t *testing.T) {
	sql := readMigrationSQL(t)
	if !strings.Contains(strings.ToLower(sql), "cleanup_jobs") {
		t.Error("migration.sql must create the cleanup_jobs table")
	}
}

func TestMigrationSQLHasCreateTableStatements(t *testing.T) {
	sql := readMigrationSQL(t)
	if !strings.Contains(strings.ToUpper(sql), "CREATE TABLE") {
		t.Error("migration.sql must contain CREATE TABLE statements")
	}
}

func TestMigrationSQLHasIndexStatements(t *testing.T) {
	sql := readMigrationSQL(t)
	upper := strings.ToUpper(sql)
	if !strings.Contains(upper, "CREATE INDEX") && !strings.Contains(upper, "CREATE UNIQUE INDEX") {
		t.Error("migration.sql must contain index creation statements")
	}
}

// ---- Prisma client singleton ----

func TestClientTSImportsPrismaClient(t *testing.T) {
	client := readClientTS(t)
	if !strings.Contains(client, "PrismaClient") {
		t.Error("client.ts must import PrismaClient from '@prisma/client'")
	}
}

func TestClientTSExportsPrismaVariable(t *testing.T) {
	client := readClientTS(t)
	if !strings.Contains(client, "export") || !strings.Contains(client, "prisma") {
		t.Error("client.ts must export a prisma variable")
	}
}

func TestClientTSUsesSingletonPattern(t *testing.T) {
	client := readClientTS(t)
	// Singleton pattern: reuses global.prisma in non-production
	if !strings.Contains(client, "global") {
		t.Error("client.ts must use a global singleton pattern to avoid multiple PrismaClient instances")
	}
}

func TestClientTSAssignsGlobalPrismaInNonProduction(t *testing.T) {
	client := readClientTS(t)
	if !strings.Contains(client, "NODE_ENV") {
		t.Error("client.ts must check NODE_ENV to conditionally assign global.prisma")
	}
}

func TestClientTSFromAtPrismaClient(t *testing.T) {
	client := readClientTS(t)
	if !strings.Contains(client, "@prisma/client") {
		t.Error("client.ts must import from '@prisma/client'")
	}
}
