// Package cleanup_job_test validates the cleanup job files for TASK-4618.
// These are integration-style file-system tests that verify cleanup.job.ts
// and its test file contain all required patterns for the scheduled cleanup job.
package cleanup_job_test

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

func readCleanupJob(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "src", "jobs", "cleanup.job.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read cleanup.job.ts: %v", err)
	}
	return string(data)
}

func readCleanupJobTestFile(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "tests", "unit", "jobs", "cleanup.job.test.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read cleanup.job.test.ts: %v", err)
	}
	return string(data)
}

// ---- File existence ----

func TestCleanupJobFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "src", "jobs", "cleanup.job.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("cleanup.job.ts not found at %s", path)
	}
}

func TestCleanupJobTestFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "tests", "unit", "jobs", "cleanup.job.test.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("cleanup.job.test.ts not found at %s", path)
	}
}

// ---- Cron schedule ----

func TestCleanupJobUsesCorrectCronSchedule(t *testing.T) {
	job := readCleanupJob(t)
	if !strings.Contains(job, "0 2 * * *") {
		t.Error("cleanup.job.ts must register cron job with schedule '0 2 * * *' (daily at 02:00 UTC)")
	}
}

func TestCleanupJobImportsNodeCron(t *testing.T) {
	job := readCleanupJob(t)
	if !strings.Contains(job, "node-cron") {
		t.Error("cleanup.job.ts must import node-cron")
	}
}

func TestCleanupJobCallsCronSchedule(t *testing.T) {
	job := readCleanupJob(t)
	if !strings.Contains(job, "cron.schedule") && !strings.Contains(job, "schedule(") {
		t.Error("cleanup.job.ts must call cron.schedule to register the cron job")
	}
}

// ---- Service calls ----

func TestCleanupJobCallsRunAgeBased(t *testing.T) {
	job := readCleanupJob(t)
	if !strings.Contains(job, "runAgeBased") {
		t.Error("cleanup.job.ts must call cleanupService.runAgeBased()")
	}
}

func TestCleanupJobCallsRunCountBased(t *testing.T) {
	job := readCleanupJob(t)
	if !strings.Contains(job, "runCountBased") {
		t.Error("cleanup.job.ts must call cleanupService.runCountBased()")
	}
}

func TestCleanupJobImportsCleanupService(t *testing.T) {
	job := readCleanupJob(t)
	if !strings.Contains(job, "cleanup.service") && !strings.Contains(job, "cleanupService") {
		t.Error("cleanup.job.ts must import from cleanup.service")
	}
}

// ---- Error handling ----

func TestCleanupJobWrapsInTryCatch(t *testing.T) {
	job := readCleanupJob(t)
	if !strings.Contains(job, "try") || !strings.Contains(job, "catch") {
		t.Error("cleanup.job.ts must wrap cleanup calls in try/catch to prevent job crashes")
	}
}

func TestCleanupJobLogsErrors(t *testing.T) {
	job := readCleanupJob(t)
	if !strings.Contains(job, "error") && !strings.Contains(job, "logger") && !strings.Contains(job, "console") {
		t.Error("cleanup.job.ts must log errors when cleanup services fail")
	}
}

// ---- Export ----

func TestCleanupJobExportsStartCleanupJob(t *testing.T) {
	job := readCleanupJob(t)
	if !strings.Contains(job, "startCleanupJob") {
		t.Error("cleanup.job.ts must define and export a startCleanupJob function")
	}
}

func TestCleanupJobUsesExportKeyword(t *testing.T) {
	job := readCleanupJob(t)
	if !strings.Contains(job, "export") {
		t.Error("cleanup.job.ts must export startCleanupJob with the export keyword")
	}
}

// ---- NODE_ENV guard ----

func TestCleanupJobHasNodeEnvGuard(t *testing.T) {
	job := readCleanupJob(t)
	if !strings.Contains(job, "NODE_ENV") {
		t.Error("cleanup.job.ts must guard against running during tests using NODE_ENV check")
	}
}

// ---- Test file content patterns ----

func TestCleanupJobTestFileMocksNodeCron(t *testing.T) {
	testFile := readCleanupJobTestFile(t)
	if !strings.Contains(testFile, "node-cron") {
		t.Error("cleanup.job.test.ts must mock node-cron")
	}
}

func TestCleanupJobTestFileMocksCleanupService(t *testing.T) {
	testFile := readCleanupJobTestFile(t)
	if !strings.Contains(testFile, "cleanupService") {
		t.Error("cleanup.job.test.ts must mock cleanupService")
	}
}

func TestCleanupJobTestFileUsesMockFramework(t *testing.T) {
	testFile := readCleanupJobTestFile(t)
	if !strings.Contains(testFile, "jest.mock") {
		t.Error("cleanup.job.test.ts must use jest.mock to mock dependencies")
	}
}

func TestCleanupJobTestFileTestsCronSchedule(t *testing.T) {
	testFile := readCleanupJobTestFile(t)
	if !strings.Contains(testFile, "0 2 * * *") {
		t.Error("cleanup.job.test.ts must test that cron job is registered with schedule '0 2 * * *'")
	}
}

func TestCleanupJobTestFileTestsRunAgeBased(t *testing.T) {
	testFile := readCleanupJobTestFile(t)
	if !strings.Contains(testFile, "runAgeBased") {
		t.Error("cleanup.job.test.ts must test that runAgeBased is called when cron fires")
	}
}

func TestCleanupJobTestFileTestsRunCountBased(t *testing.T) {
	testFile := readCleanupJobTestFile(t)
	if !strings.Contains(testFile, "runCountBased") {
		t.Error("cleanup.job.test.ts must test that runCountBased is called when cron fires")
	}
}

func TestCleanupJobTestFileTestsErrorHandling(t *testing.T) {
	testFile := readCleanupJobTestFile(t)
	if !strings.Contains(testFile, "error") && !strings.Contains(testFile, "Error") {
		t.Error("cleanup.job.test.ts must test error handling behavior")
	}
}

func TestCleanupJobTestFileTestsJobDoesNotCrash(t *testing.T) {
	testFile := readCleanupJobTestFile(t)
	if !strings.Contains(testFile, "resolves") && !strings.Contains(testFile, "not.toThrow") {
		t.Error("cleanup.job.test.ts must test that the cron job does not crash on errors")
	}
}

func TestCleanupJobTestFileImportsStartCleanupJob(t *testing.T) {
	testFile := readCleanupJobTestFile(t)
	if !strings.Contains(testFile, "startCleanupJob") {
		t.Error("cleanup.job.test.ts must import and test startCleanupJob")
	}
}
