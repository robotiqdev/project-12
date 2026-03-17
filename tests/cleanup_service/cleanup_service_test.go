// Package cleanup_service_test validates the cleanup service file for TASK-4616.
// These are integration-style file-system tests that verify cleanup.service.ts
// and its test file contain all required patterns for the age-based cleanup service.
package cleanup_service_test

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

// ---- Helpers ----

func readCleanupService(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "src", "services", "cleanup.service.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read cleanup.service.ts: %v", err)
	}
	return string(data)
}

func readCleanupServiceTestFile(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "tests", "unit", "services", "cleanup.service.test.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read cleanup.service.test.ts: %v", err)
	}
	return string(data)
}

// ---- File existence ----

func TestCleanupServiceFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "src", "services", "cleanup.service.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("cleanup.service.ts not found at %s", path)
	}
}

func TestCleanupServiceTestFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "tests", "unit", "services", "cleanup.service.test.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("cleanup.service.test.ts not found at %s", path)
	}
}

// ---- Service exports ----

func TestCleanupServiceExportsRunAgeBased(t *testing.T) {
	svc := readCleanupService(t)
	if !strings.Contains(svc, "runAgeBased") {
		t.Error("cleanup.service.ts must export a runAgeBased function")
	}
}

func TestCleanupServiceExportsFunction(t *testing.T) {
	svc := readCleanupService(t)
	if !strings.Contains(svc, "export") {
		t.Error("cleanup.service.ts must use export keyword")
	}
}

// ---- CleanupJob lifecycle ----

func TestCleanupServiceCreatesCleanupJobWithAgeBasedType(t *testing.T) {
	svc := readCleanupService(t)
	if !strings.Contains(svc, "AGE_BASED") {
		t.Error("cleanup.service.ts must create a CleanupJob of type AGE_BASED")
	}
}

func TestCleanupServiceCallsCleanupJobModelCreate(t *testing.T) {
	svc := readCleanupService(t)
	if !strings.Contains(svc, "create") {
		t.Error("cleanup.service.ts must call cleanupJobModel.create (or create)")
	}
}

func TestCleanupServiceCallsStart(t *testing.T) {
	svc := readCleanupService(t)
	if !strings.Contains(svc, "start") {
		t.Error("cleanup.service.ts must call start to mark the job as started")
	}
}

func TestCleanupServiceCallsComplete(t *testing.T) {
	svc := readCleanupService(t)
	if !strings.Contains(svc, "complete") {
		t.Error("cleanup.service.ts must call complete with the deletedCount after deletion")
	}
}

func TestCleanupServiceCallsFailOnError(t *testing.T) {
	svc := readCleanupService(t)
	if !strings.Contains(svc, "fail") {
		t.Error("cleanup.service.ts must call fail on the CleanupJob when an error occurs")
	}
}

// ---- 90-day cutoff calculation ----

func TestCleanupServiceUses90DayCutoff(t *testing.T) {
	svc := readCleanupService(t)
	if !strings.Contains(svc, "90") {
		t.Error("cleanup.service.ts must compute a 90-day cutoff date")
	}
}

func TestCleanupServiceComputesCutoffFromDateNow(t *testing.T) {
	svc := readCleanupService(t)
	if !strings.Contains(svc, "Date.now()") && !strings.Contains(svc, "new Date()") {
		t.Error("cleanup.service.ts must use Date.now() or new Date() to compute the cutoff timestamp")
	}
}

func TestCleanupServiceCutoffUsesMilliseconds(t *testing.T) {
	svc := readCleanupService(t)
	// 90 * 24 * 60 * 60 * 1000 is the millisecond calculation
	if !strings.Contains(svc, "1000") {
		t.Error("cleanup.service.ts cutoff calculation must convert days to milliseconds (multiply by 1000)")
	}
}

// ---- Prisma deleteMany call ----

func TestCleanupServiceCallsValidationRunDeleteMany(t *testing.T) {
	svc := readCleanupService(t)
	if !strings.Contains(svc, "deleteMany") {
		t.Error("cleanup.service.ts must call prisma.validationRun.deleteMany")
	}
}

func TestCleanupServiceFiltersOnCreatedAt(t *testing.T) {
	svc := readCleanupService(t)
	if !strings.Contains(svc, "createdAt") {
		t.Error("cleanup.service.ts deleteMany must filter on createdAt field")
	}
}

func TestCleanupServiceUsesLtOperatorForCutoff(t *testing.T) {
	svc := readCleanupService(t)
	if !strings.Contains(svc, "lt") {
		t.Error("cleanup.service.ts must use the 'lt' (less than) Prisma operator for the cutoff date")
	}
}

func TestCleanupServiceExcludesRunningStatus(t *testing.T) {
	svc := readCleanupService(t)
	if !strings.Contains(svc, "RUNNING") {
		t.Error("cleanup.service.ts must reference RUNNING status to exclude running runs from deletion")
	}
}

func TestCleanupServiceUsesNotOperatorForRunningExclusion(t *testing.T) {
	svc := readCleanupService(t)
	if !strings.Contains(svc, "not") {
		t.Error("cleanup.service.ts must use the 'not' Prisma operator to exclude RUNNING status")
	}
}

// ---- Cascade delete via Prisma ----

func TestCleanupServiceReliesOnCascadeDelete(t *testing.T) {
	svc := readCleanupService(t)
	// Service should only call deleteMany on validationRun (not manually delete RunStages)
	// because cascade delete handles RunStages automatically
	if !strings.Contains(svc, "validationRun") {
		t.Error("cleanup.service.ts must use prisma.validationRun.deleteMany (cascade handles RunStages)")
	}
}

// ---- Return type ----

func TestCleanupServiceReturnsCleanupResult(t *testing.T) {
	svc := readCleanupService(t)
	if !strings.Contains(svc, "CleanupResult") {
		t.Error("cleanup.service.ts must return a CleanupResult object")
	}
}

func TestCleanupServiceResultContainsDeletedCount(t *testing.T) {
	svc := readCleanupService(t)
	if !strings.Contains(svc, "deletedCount") {
		t.Error("cleanup.service.ts must include deletedCount in the CleanupResult")
	}
}

func TestCleanupServiceResultContainsDuration(t *testing.T) {
	svc := readCleanupService(t)
	if !strings.Contains(svc, "duration") {
		t.Error("cleanup.service.ts must include duration in the CleanupResult")
	}
}

func TestCleanupServiceResultContainsJobId(t *testing.T) {
	svc := readCleanupService(t)
	if !strings.Contains(svc, "jobId") {
		t.Error("cleanup.service.ts must include jobId in the CleanupResult")
	}
}

// ---- Error handling ----

func TestCleanupServiceWrapsInTryCatch(t *testing.T) {
	svc := readCleanupService(t)
	if !strings.Contains(svc, "try") || !strings.Contains(svc, "catch") {
		t.Error("cleanup.service.ts must wrap the deletion in a try/catch block")
	}
}

func TestCleanupServicePassesErrorMessageToFail(t *testing.T) {
	svc := readCleanupService(t)
	if !strings.Contains(svc, "message") {
		t.Error("cleanup.service.ts must pass the error message to cleanupJobModel.fail")
	}
}

// ---- Imports ----

func TestCleanupServiceImportsPrisma(t *testing.T) {
	svc := readCleanupService(t)
	if !strings.Contains(svc, "prisma") {
		t.Error("cleanup.service.ts must import/use the prisma client")
	}
}

func TestCleanupServiceImportsCleanupJobModel(t *testing.T) {
	svc := readCleanupService(t)
	if !strings.Contains(svc, "cleanup-job.model") || !strings.Contains(svc, "cleanupJobModel") {
		if !strings.Contains(svc, "cleanup-job") {
			t.Error("cleanup.service.ts must import from cleanup-job.model")
		}
	}
}

func TestCleanupServiceImportsCleanupJobType(t *testing.T) {
	svc := readCleanupService(t)
	if !strings.Contains(svc, "CleanupJobType") {
		t.Error("cleanup.service.ts must import CleanupJobType")
	}
}

// ---- Test file patterns ----

func TestCleanupServiceTestFileDescribesRunAgeBased(t *testing.T) {
	testFile := readCleanupServiceTestFile(t)
	if !strings.Contains(testFile, "runAgeBased") {
		t.Error("cleanup.service.test.ts must contain tests for runAgeBased")
	}
}

func TestCleanupServiceTestFileMocksPrisma(t *testing.T) {
	testFile := readCleanupServiceTestFile(t)
	if !strings.Contains(testFile, "prisma") {
		t.Error("cleanup.service.test.ts must mock the prisma client")
	}
}

func TestCleanupServiceTestFileMocksCleanupJobModel(t *testing.T) {
	testFile := readCleanupServiceTestFile(t)
	if !strings.Contains(testFile, "cleanupJobModel") || !strings.Contains(testFile, "mock") {
		if !strings.Contains(testFile, "jest.mock") && !strings.Contains(testFile, "vi.mock") {
			t.Error("cleanup.service.test.ts must mock cleanupJobModel")
		}
	}
}

func TestCleanupServiceTestFileTests90DayCutoff(t *testing.T) {
	testFile := readCleanupServiceTestFile(t)
	if !strings.Contains(testFile, "90") {
		t.Error("cleanup.service.test.ts must test the 90-day cutoff date calculation")
	}
}

func TestCleanupServiceTestFileTestsRunningExclusion(t *testing.T) {
	testFile := readCleanupServiceTestFile(t)
	if !strings.Contains(testFile, "RUNNING") {
		t.Error("cleanup.service.test.ts must test that RUNNING runs are not deleted")
	}
}

func TestCleanupServiceTestFileTestsCascadeDelete(t *testing.T) {
	testFile := readCleanupServiceTestFile(t)
	if !strings.Contains(testFile, "cascade") && !strings.Contains(testFile, "RunStage") && !strings.Contains(testFile, "deleteMany") {
		t.Error("cleanup.service.test.ts must test cascade delete behavior for RunStage records")
	}
}

func TestCleanupServiceTestFileTestsCleanupJobLifecycle(t *testing.T) {
	testFile := readCleanupServiceTestFile(t)
	lifecyclePatterns := []string{"start", "complete", "deletedCount"}
	for _, pattern := range lifecyclePatterns {
		if !strings.Contains(testFile, pattern) {
			t.Errorf("cleanup.service.test.ts must test CleanupJob lifecycle: missing '%s'", pattern)
		}
	}
}

func TestCleanupServiceTestFileTestsFailOnError(t *testing.T) {
	testFile := readCleanupServiceTestFile(t)
	if !strings.Contains(testFile, "fail") {
		t.Error("cleanup.service.test.ts must test that fail is called when Prisma throws")
	}
}

func TestCleanupServiceTestFileTestsErrorMessage(t *testing.T) {
	testFile := readCleanupServiceTestFile(t)
	if !strings.Contains(testFile, "message") {
		t.Error("cleanup.service.test.ts must verify the error message is passed to fail")
	}
}
