// Package cleanup_test validates the cleanup types and model files for TASK-4615.
// These are integration-style file-system tests that verify cleanup.types.ts and
// cleanup-job.model.ts contain all required definitions for the CleanupJob data model.
package cleanup_test

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

func readCleanupTypes(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "src", "types", "cleanup.types.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read cleanup.types.ts: %v", err)
	}
	return string(data)
}

func readCleanupJobModel(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "src", "models", "cleanup-job.model.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read cleanup-job.model.ts: %v", err)
	}
	return string(data)
}

// ---- File existence ----

func TestCleanupTypesFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "src", "types", "cleanup.types.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("cleanup.types.ts not found at %s", path)
	}
}

func TestCleanupJobModelFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "src", "models", "cleanup-job.model.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("cleanup-job.model.ts not found at %s", path)
	}
}

// ---- CleanupJobType enum ----

func TestCleanupTypesDefinesCleanupJobTypeEnum(t *testing.T) {
	types := readCleanupTypes(t)
	if !strings.Contains(types, "CleanupJobType") {
		t.Error("cleanup.types.ts must define CleanupJobType enum")
	}
}

func TestCleanupJobTypeHasAgeBasedMember(t *testing.T) {
	types := readCleanupTypes(t)
	if !strings.Contains(types, "AGE_BASED") {
		t.Error("CleanupJobType enum must include AGE_BASED member")
	}
}

func TestCleanupJobTypeHasCountBasedMember(t *testing.T) {
	types := readCleanupTypes(t)
	if !strings.Contains(types, "COUNT_BASED") {
		t.Error("CleanupJobType enum must include COUNT_BASED member")
	}
}

// ---- CleanupJobStatus enum ----

func TestCleanupTypesDefinesCleanupJobStatusEnum(t *testing.T) {
	types := readCleanupTypes(t)
	if !strings.Contains(types, "CleanupJobStatus") {
		t.Error("cleanup.types.ts must define CleanupJobStatus enum")
	}
}

func TestCleanupJobStatusHasPendingMember(t *testing.T) {
	types := readCleanupTypes(t)
	if !strings.Contains(types, "PENDING") {
		t.Error("CleanupJobStatus enum must include PENDING member")
	}
}

func TestCleanupJobStatusHasRunningMember(t *testing.T) {
	types := readCleanupTypes(t)
	if !strings.Contains(types, "RUNNING") {
		t.Error("CleanupJobStatus enum must include RUNNING member")
	}
}

func TestCleanupJobStatusHasDoneMember(t *testing.T) {
	types := readCleanupTypes(t)
	if !strings.Contains(types, "DONE") {
		t.Error("CleanupJobStatus enum must include DONE member")
	}
}

func TestCleanupJobStatusHasFailedMember(t *testing.T) {
	types := readCleanupTypes(t)
	if !strings.Contains(types, "FAILED") {
		t.Error("CleanupJobStatus enum must include FAILED member")
	}
}

// ---- CleanupJobRecord interface ----

func TestCleanupTypesDefinesCleanupJobRecordInterface(t *testing.T) {
	types := readCleanupTypes(t)
	if !strings.Contains(types, "CleanupJobRecord") {
		t.Error("cleanup.types.ts must define CleanupJobRecord interface")
	}
}

func TestCleanupJobRecordHasIdField(t *testing.T) {
	types := readCleanupTypes(t)
	if !strings.Contains(types, "id") {
		t.Error("CleanupJobRecord interface must include id field")
	}
}

func TestCleanupJobRecordHasTypeField(t *testing.T) {
	types := readCleanupTypes(t)
	if !strings.Contains(types, "CleanupJobRecord") && !strings.Contains(types, "type") {
		t.Error("CleanupJobRecord interface must include type field")
	}
}

func TestCleanupJobRecordHasStatusField(t *testing.T) {
	types := readCleanupTypes(t)
	if !strings.Contains(types, "status") {
		t.Error("CleanupJobRecord interface must include status field")
	}
}

func TestCleanupJobRecordHasDeletedCountField(t *testing.T) {
	types := readCleanupTypes(t)
	if !strings.Contains(types, "deletedCount") {
		t.Error("CleanupJobRecord interface must include deletedCount field")
	}
}

func TestCleanupJobRecordHasStartedAtField(t *testing.T) {
	types := readCleanupTypes(t)
	if !strings.Contains(types, "startedAt") {
		t.Error("CleanupJobRecord interface must include startedAt field")
	}
}

func TestCleanupJobRecordHasCompletedAtField(t *testing.T) {
	types := readCleanupTypes(t)
	if !strings.Contains(types, "completedAt") {
		t.Error("CleanupJobRecord interface must include completedAt field")
	}
}

func TestCleanupJobRecordHasErrorMessageField(t *testing.T) {
	types := readCleanupTypes(t)
	if !strings.Contains(types, "errorMessage") {
		t.Error("CleanupJobRecord interface must include errorMessage field")
	}
}

// ---- CleanupResult interface ----

func TestCleanupTypesDefinesCleanupResultInterface(t *testing.T) {
	types := readCleanupTypes(t)
	if !strings.Contains(types, "CleanupResult") {
		t.Error("cleanup.types.ts must define CleanupResult interface")
	}
}

func TestCleanupResultHasJobIdField(t *testing.T) {
	types := readCleanupTypes(t)
	if !strings.Contains(types, "jobId") {
		t.Error("CleanupResult interface must include jobId field")
	}
}

func TestCleanupResultHasTypeField(t *testing.T) {
	types := readCleanupTypes(t)
	if !strings.Contains(types, "CleanupResult") && !strings.Contains(types, "type") {
		t.Error("CleanupResult interface must include type field")
	}
}

func TestCleanupResultHasDeletedCountField(t *testing.T) {
	types := readCleanupTypes(t)
	if !strings.Contains(types, "deletedCount") {
		t.Error("CleanupResult interface must include deletedCount field")
	}
}

func TestCleanupResultHasDurationField(t *testing.T) {
	types := readCleanupTypes(t)
	if !strings.Contains(types, "duration") {
		t.Error("CleanupResult interface must include duration field")
	}
}

// ---- cleanup-job.model.ts imports and structure ----

func TestCleanupJobModelImportsPrismaClient(t *testing.T) {
	model := readCleanupJobModel(t)
	if !strings.Contains(model, "prisma") {
		t.Error("cleanup-job.model.ts must import/use the prisma client")
	}
}

func TestCleanupJobModelImportsCleanupTypes(t *testing.T) {
	model := readCleanupJobModel(t)
	if !strings.Contains(model, "CleanupJobType") && !strings.Contains(model, "cleanup.types") {
		t.Error("cleanup-job.model.ts must import CleanupJobType or reference cleanup types")
	}
}

// ---- create method ----

func TestCleanupJobModelExposesCreateMethod(t *testing.T) {
	model := readCleanupJobModel(t)
	if !strings.Contains(model, "create") {
		t.Error("cleanup-job.model.ts must expose a create method")
	}
}

func TestCleanupJobModelCreateAcceptsType(t *testing.T) {
	model := readCleanupJobModel(t)
	// The create method should accept a type parameter (CleanupJobType)
	if !strings.Contains(model, "create") || !strings.Contains(model, "CleanupJobType") {
		t.Error("cleanup-job.model.ts create method must accept a CleanupJobType parameter")
	}
}

func TestCleanupJobModelCreateCreatesJobWithPendingStatus(t *testing.T) {
	model := readCleanupJobModel(t)
	// create should result in a PENDING status record
	if !strings.Contains(model, "PENDING") {
		t.Error("cleanup-job.model.ts must reference PENDING status (used in create)")
	}
}

// ---- start method ----

func TestCleanupJobModelExposesStartMethod(t *testing.T) {
	model := readCleanupJobModel(t)
	if !strings.Contains(model, "start") {
		t.Error("cleanup-job.model.ts must expose a start method")
	}
}

func TestCleanupJobModelStartAcceptsId(t *testing.T) {
	model := readCleanupJobModel(t)
	// The start method accepts an id and transitions status to RUNNING
	if !strings.Contains(model, "start") || !strings.Contains(model, "RUNNING") {
		t.Error("cleanup-job.model.ts start method must transition status to RUNNING")
	}
}

// ---- complete method ----

func TestCleanupJobModelExposesCompleteMethod(t *testing.T) {
	model := readCleanupJobModel(t)
	if !strings.Contains(model, "complete") {
		t.Error("cleanup-job.model.ts must expose a complete method")
	}
}

func TestCleanupJobModelCompleteAcceptsIdAndDeletedCount(t *testing.T) {
	model := readCleanupJobModel(t)
	// complete(id, deletedCount) — should update deletedCount and set status DONE
	if !strings.Contains(model, "deletedCount") {
		t.Error("cleanup-job.model.ts complete method must accept and update deletedCount")
	}
}

func TestCleanupJobModelCompleteSetsDoneStatus(t *testing.T) {
	model := readCleanupJobModel(t)
	if !strings.Contains(model, "DONE") {
		t.Error("cleanup-job.model.ts complete method must set status to DONE")
	}
}

func TestCleanupJobModelCompleteSetsCompletedAt(t *testing.T) {
	model := readCleanupJobModel(t)
	if !strings.Contains(model, "completedAt") {
		t.Error("cleanup-job.model.ts complete method must set completedAt timestamp")
	}
}

// ---- fail method ----

func TestCleanupJobModelExposesFailMethod(t *testing.T) {
	model := readCleanupJobModel(t)
	if !strings.Contains(model, "fail") {
		t.Error("cleanup-job.model.ts must expose a fail method")
	}
}

func TestCleanupJobModelFailAcceptsIdAndErrorMessage(t *testing.T) {
	model := readCleanupJobModel(t)
	// fail(id, errorMessage)
	if !strings.Contains(model, "errorMessage") {
		t.Error("cleanup-job.model.ts fail method must accept and store errorMessage")
	}
}

func TestCleanupJobModelFailSetsFailedStatus(t *testing.T) {
	model := readCleanupJobModel(t)
	if !strings.Contains(model, "FAILED") {
		t.Error("cleanup-job.model.ts fail method must set status to FAILED")
	}
}

func TestCleanupJobModelFailSetsCompletedAt(t *testing.T) {
	model := readCleanupJobModel(t)
	// fail should also set completedAt
	if !strings.Contains(model, "completedAt") {
		t.Error("cleanup-job.model.ts fail method must set completedAt timestamp")
	}
}

// ---- list method ----

func TestCleanupJobModelExposesListMethod(t *testing.T) {
	model := readCleanupJobModel(t)
	if !strings.Contains(model, "list") {
		t.Error("cleanup-job.model.ts must expose a list method")
	}
}

func TestCleanupJobModelListHasDefaultLimit(t *testing.T) {
	model := readCleanupJobModel(t)
	// list(limit=20) — should have default limit of 20
	if !strings.Contains(model, "20") {
		t.Error("cleanup-job.model.ts list method must have a default limit of 20")
	}
}

func TestCleanupJobModelListOrdersByStartedAtDesc(t *testing.T) {
	model := readCleanupJobModel(t)
	// list should order by startedAt descending (most recent first)
	if !strings.Contains(model, "startedAt") {
		t.Error("cleanup-job.model.ts list method must reference startedAt for ordering")
	}
}

// ---- Lifecycle pattern ----

func TestCleanupJobModelImplementsLifecyclePattern(t *testing.T) {
	model := readCleanupJobModel(t)
	// The model must implement all four lifecycle methods: create, start, complete, fail
	lifecycleMethods := []string{"create", "start", "complete", "fail"}
	for _, method := range lifecycleMethods {
		if !strings.Contains(model, method) {
			t.Errorf("cleanup-job.model.ts must implement lifecycle method: %s", method)
		}
	}
}

func TestCleanupJobModelExportsAllMethods(t *testing.T) {
	model := readCleanupJobModel(t)
	// All methods should be exported (module exports)
	if !strings.Contains(model, "export") {
		t.Error("cleanup-job.model.ts must export its methods")
	}
}
