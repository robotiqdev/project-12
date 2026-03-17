// Package integration_test validates the TASK-4623 query performance test
// infrastructure. These file-system tests verify that:
//   - backend/tests/helpers/db.helper.ts exists and exports a seedRuns factory
//   - backend/tests/setup.ts exists with Jest global setup logic
//   - backend/jest.config.ts exists with required configuration
//   - backend/tests/integration/query-performance.test.ts contains the
//     TASK-4623 performance tests (500 seeds, 10 branches, 5 EXPLAIN assertions)
package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---- Helper file readers ----

func readDbHelper(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "tests", "helpers", "db.helper.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read db.helper.ts: %v", err)
	}
	return string(data)
}

func readSetupFile(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "tests", "setup.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read setup.ts: %v", err)
	}
	return string(data)
}

func readJestConfig(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "jest.config.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read jest.config.ts: %v", err)
	}
	return string(data)
}

func readQueryPerformanceTest(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "tests", "integration", "query-performance.test.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read query-performance.test.ts: %v", err)
	}
	return string(data)
}

// ---- db.helper.ts: file existence ----

func TestDbHelperFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "tests", "helpers", "db.helper.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("backend/tests/helpers/db.helper.ts not found — implementation_developer must create it")
	}
}

// ---- db.helper.ts: seedRuns factory function ----

func TestDbHelperExportsSeedRunsFunction(t *testing.T) {
	content := readDbHelper(t)
	if !strings.Contains(content, "seedRuns") {
		t.Error("db.helper.ts must export a seedRuns factory function")
	}
}

func TestDbHelperSeedRunsUsesCreateMany(t *testing.T) {
	content := readDbHelper(t)
	if !strings.Contains(content, "createMany") {
		t.Error("db.helper.ts seedRuns must use prisma.validationRun.createMany for efficient bulk insert")
	}
}

func TestDbHelperSeedRunsAcceptsCountParameter(t *testing.T) {
	content := readDbHelper(t)
	// Factory should accept a count argument
	if !strings.Contains(content, "count") {
		t.Error("db.helper.ts seedRuns must accept a count parameter")
	}
}

func TestDbHelperImportsPrismaClient(t *testing.T) {
	content := readDbHelper(t)
	if !strings.Contains(content, "prisma") && !strings.Contains(content, "PrismaClient") {
		t.Error("db.helper.ts must import or use the Prisma client")
	}
}

func TestDbHelperExportsFunction(t *testing.T) {
	content := readDbHelper(t)
	if !strings.Contains(content, "export") {
		t.Error("db.helper.ts must export seedRuns so it can be imported by test files")
	}
}

// ---- db.helper.ts: seeding data variety ----

func TestDbHelperSeedRunsSupportsOptions(t *testing.T) {
	content := readDbHelper(t)
	// Should accept options for customising seed (branches, statuses, timestamps)
	if !strings.Contains(content, "options") && !strings.Contains(content, "Options") && !strings.Contains(content, "opts") {
		t.Error("db.helper.ts seedRuns must accept an options parameter for customising seed data (branches, statuses, timestamps)")
	}
}

// ---- setup.ts: file existence ----

func TestSetupFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "tests", "setup.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("backend/tests/setup.ts not found — implementation_developer must create it")
	}
}

// ---- setup.ts: Jest global setup ----

func TestSetupFileConnectsToTestDatabase(t *testing.T) {
	content := readSetupFile(t)
	if !strings.Contains(content, "DATABASE_URL") {
		t.Error("setup.ts must reference DATABASE_URL environment variable for test database connection")
	}
}

func TestSetupFileRunsMigrations(t *testing.T) {
	content := readSetupFile(t)
	lower := strings.ToLower(content)
	if !strings.Contains(lower, "migrat") {
		t.Error("setup.ts must run database migrations as part of global setup")
	}
}

func TestSetupFileExportsDefaultFunction(t *testing.T) {
	content := readSetupFile(t)
	if !strings.Contains(content, "export default") && !strings.Contains(content, "module.exports") {
		t.Error("setup.ts must export a default async function for Jest globalSetup")
	}
}

// ---- jest.config.ts: file existence ----

func TestJestConfigFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "jest.config.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("backend/jest.config.ts not found — implementation_developer must create it")
	}
}

// ---- jest.config.ts: required configuration ----

func TestJestConfigHasGlobalSetup(t *testing.T) {
	content := readJestConfig(t)
	if !strings.Contains(content, "globalSetup") {
		t.Error("jest.config.ts must include globalSetup pointing to the setup file")
	}
}

func TestJestConfigGlobalSetupPointsToSetupFile(t *testing.T) {
	content := readJestConfig(t)
	if !strings.Contains(content, "setup") {
		t.Error("jest.config.ts globalSetup must reference the setup.ts file")
	}
}

func TestJestConfigHasNodeTestEnvironment(t *testing.T) {
	content := readJestConfig(t)
	if !strings.Contains(content, "testEnvironment") {
		t.Error("jest.config.ts must set testEnvironment")
	}
	if !strings.Contains(content, "node") {
		t.Error("jest.config.ts testEnvironment must be 'node' for PostgreSQL integration tests")
	}
}

func TestJestConfigHasTypeScriptTransform(t *testing.T) {
	content := readJestConfig(t)
	lower := strings.ToLower(content)
	// Should reference ts-jest or similar TypeScript transform
	if !strings.Contains(lower, "ts-jest") && !strings.Contains(lower, "transform") {
		t.Error("jest.config.ts must configure a TypeScript transform (e.g. ts-jest)")
	}
}

// ---- query-performance.test.ts: TASK-4623 describe block ----

func TestQueryPerformanceTestHasTask4623Block(t *testing.T) {
	content := readQueryPerformanceTest(t)
	if !strings.Contains(content, "TASK-4623") {
		t.Error("query-performance.test.ts must include a describe block labelled 'TASK-4623'")
	}
}

func TestQueryPerformanceTestSeeds500Records(t *testing.T) {
	content := readQueryPerformanceTest(t)
	if !strings.Contains(content, "500") {
		t.Error("query-performance.test.ts TASK-4623 must seed 500 ValidationRun records")
	}
}

func TestQueryPerformanceTestUses10Branches(t *testing.T) {
	content := readQueryPerformanceTest(t)
	// 10 branches must be defined for TASK-4623
	if !strings.Contains(content, "10") {
		t.Error("query-performance.test.ts TASK-4623 must spread runs across 10 branches")
	}
}

func TestQueryPerformanceTestUsesExplainFormatJson(t *testing.T) {
	content := readQueryPerformanceTest(t)
	if !strings.Contains(content, "EXPLAIN") || !strings.Contains(content, "FORMAT JSON") {
		t.Error("query-performance.test.ts must use EXPLAIN (FORMAT JSON) to parse query plan output")
	}
}

// ---- Test 1: idx_runs_branch_status ----

func TestQueryPerformanceTest1BranchStatusOrderByLimit(t *testing.T) {
	content := readQueryPerformanceTest(t)
	// Test 1 must include ORDER BY createdAt DESC LIMIT 20 with branchName+status filter
	if !strings.Contains(content, "LIMIT 20") {
		t.Error("query-performance.test.ts Test 1 must query with LIMIT 20")
	}
}

func TestQueryPerformanceTest1AssertsBranchStatusIndex(t *testing.T) {
	content := readQueryPerformanceTest(t)
	if !strings.Contains(content, "idx_runs_branch_status") {
		t.Error("query-performance.test.ts must assert Index Scan on idx_runs_branch_status for Test 1")
	}
}

// ---- Test 2: idx_runs_branch_commit ----

func TestQueryPerformanceTest2BranchCommitLikeQuery(t *testing.T) {
	content := readQueryPerformanceTest(t)
	upper := strings.ToUpper(content)
	if !strings.Contains(upper, "LIKE") {
		t.Error("query-performance.test.ts Test 2 must include a LIKE predicate on commitSha")
	}
}

func TestQueryPerformanceTest2AssertsBranchCommitIndex(t *testing.T) {
	content := readQueryPerformanceTest(t)
	if !strings.Contains(content, "idx_runs_branch_commit") {
		t.Error("query-performance.test.ts must assert Index Scan on idx_runs_branch_commit for Test 2")
	}
}

// ---- Test 3: partial index concurrency check ----

func TestQueryPerformanceTest3ConcurrencyCheckQuery(t *testing.T) {
	content := readQueryPerformanceTest(t)
	upper := strings.ToUpper(content)
	// Concurrency check queries for PENDING and RUNNING statuses
	if !strings.Contains(upper, "PENDING") || !strings.Contains(upper, "RUNNING") {
		t.Error("query-performance.test.ts Test 3 (concurrency check) must filter on PENDING and RUNNING statuses")
	}
}

func TestQueryPerformanceTest3AssertsPartialIndex(t *testing.T) {
	content := readQueryPerformanceTest(t)
	if !strings.Contains(content, "idx_runs_active") {
		t.Error("query-performance.test.ts Test 3 must assert the partial index idx_runs_active is used for the concurrency check")
	}
}

// ---- Test 4: partition pruning ----

func TestQueryPerformanceTest4PartitionPruning(t *testing.T) {
	content := readQueryPerformanceTest(t)
	lower := strings.ToLower(content)
	if !strings.Contains(lower, "partition") {
		t.Error("query-performance.test.ts Test 4 must test partition pruning for date-range queries")
	}
}

func TestQueryPerformanceTest4DateRangeQuery(t *testing.T) {
	content := readQueryPerformanceTest(t)
	// Must include a date range (createdAt >= ... AND createdAt < ...)
	if !strings.Contains(content, "createdAt") && !strings.Contains(content, "created_at") {
		t.Error("query-performance.test.ts Test 4 must include a createdAt date-range predicate for partition pruning")
	}
}

func TestQueryPerformanceTest4AssertsSubplansRemoved(t *testing.T) {
	content := readQueryPerformanceTest(t)
	// The partition pruning assertion must check Subplans Removed or similar
	if !strings.Contains(content, "Subplans Removed") && !strings.Contains(content, "partition pruning") &&
		!strings.Contains(content, "Partition Pruning") {
		t.Error("query-performance.test.ts Test 4 must assert partition pruning via 'Subplans Removed' or equivalent")
	}
}

// ---- Test 5: idx_stages_run_index ----

func TestQueryPerformanceTest5RunStagesOrderByStageIndex(t *testing.T) {
	content := readQueryPerformanceTest(t)
	upper := strings.ToUpper(content)
	// Must query run_stages with ORDER BY stageIndex
	if !strings.Contains(strings.ToLower(content), "run_stages") {
		t.Error("query-performance.test.ts Test 5 must query the run_stages table")
	}
	if !strings.Contains(upper, "ORDER BY") || !strings.Contains(content, "stageIndex") {
		t.Error("query-performance.test.ts Test 5 must include ORDER BY stageIndex")
	}
}

func TestQueryPerformanceTest5AssertsStagesRunIndex(t *testing.T) {
	content := readQueryPerformanceTest(t)
	if !strings.Contains(content, "idx_stages_run_index") {
		t.Error("query-performance.test.ts must assert Index Scan on idx_stages_run_index for Test 5")
	}
}

// ---- Integration tag ----

func TestQueryPerformanceTestHasIntegrationTag(t *testing.T) {
	content := readQueryPerformanceTest(t)
	if !strings.Contains(content, "@integration") {
		t.Error("query-performance.test.ts must be tagged with @integration so it can be run separately from unit tests")
	}
}

// ---- Seeding uses varied timestamps ----

func TestQueryPerformanceTestSeedsWithVariedTimestamps(t *testing.T) {
	content := readQueryPerformanceTest(t)
	upper := strings.ToUpper(content)
	// Must use interval-based timestamp spread for meaningful ORDER BY tests
	if !strings.Contains(upper, "INTERVAL") && !strings.Contains(upper, "TIMESTAMP") {
		t.Error("query-performance.test.ts must seed records with varied timestamps for ORDER BY createdAt to be meaningful")
	}
}

// ---- Cleanup after tests ----

func TestQueryPerformanceTestCleansUpAfterTask4623(t *testing.T) {
	content := readQueryPerformanceTest(t)
	// Must have an afterAll to clean up seeded data
	if !strings.Contains(content, "afterAll") {
		t.Error("query-performance.test.ts must include afterAll to remove seeded test data from TASK-4623")
	}
}
