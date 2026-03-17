// Package cancellation_test validates the TypeScript layer for TASK-4605.
// These file-system tests verify that validation-run.types.ts and
// validation-run.model.ts correctly expose the cancellation fields and methods.
package cancellation_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot resolves the repository root from the current working directory
// by walking up to find go.mod.
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

func readValidationRunTypes(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "src", "models", "validation-run.types.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read validation-run.types.ts: %v", err)
	}
	return string(data)
}

func readValidationRunModel(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "src", "models", "validation-run.model.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read validation-run.model.ts: %v", err)
	}
	return string(data)
}

// ---- File existence ----

func TestValidationRunTypesFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "src", "models", "validation-run.types.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("validation-run.types.ts not found at %s", path)
	}
}

func TestValidationRunModelFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "src", "models", "validation-run.model.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("validation-run.model.ts not found at %s", path)
	}
}

// ---- ValidationRunRecord interface ----

func TestValidationRunRecordInterfaceExists(t *testing.T) {
	content := readValidationRunTypes(t)
	if !strings.Contains(content, "ValidationRunRecord") {
		t.Error("validation-run.types.ts must define the ValidationRunRecord interface")
	}
}

func TestValidationRunRecordHasCancelledAtField(t *testing.T) {
	content := readValidationRunTypes(t)
	if !strings.Contains(content, "cancelledAt") {
		t.Error("ValidationRunRecord interface must include a cancelledAt field")
	}
}

func TestValidationRunRecordCancelledAtIsDateOrNull(t *testing.T) {
	content := readValidationRunTypes(t)
	// The field must be typed as Date | null (nullable)
	hasDateOrNull := strings.Contains(content, "cancelledAt: Date | null") ||
		strings.Contains(content, "cancelledAt: null | Date")
	if !hasDateOrNull {
		t.Error("ValidationRunRecord.cancelledAt must be typed as 'Date | null'")
	}
}

func TestValidationRunRecordHasIdField(t *testing.T) {
	content := readValidationRunTypes(t)
	if !strings.Contains(content, "id") {
		t.Error("ValidationRunRecord interface must include an id field")
	}
}

func TestValidationRunRecordHasStatusField(t *testing.T) {
	content := readValidationRunTypes(t)
	if !strings.Contains(content, "status") {
		t.Error("ValidationRunRecord interface must include a status field")
	}
}

func TestValidationRunRecordHasCreatedAtField(t *testing.T) {
	content := readValidationRunTypes(t)
	if !strings.Contains(content, "createdAt") {
		t.Error("ValidationRunRecord interface must include a createdAt field")
	}
}

func TestValidationRunRecordHasUpdatedAtField(t *testing.T) {
	content := readValidationRunTypes(t)
	if !strings.Contains(content, "updatedAt") {
		t.Error("ValidationRunRecord interface must include an updatedAt field")
	}
}

func TestValidationRunRecordHasCompletedAtField(t *testing.T) {
	content := readValidationRunTypes(t)
	if !strings.Contains(content, "completedAt") {
		t.Error("ValidationRunRecord interface must include a completedAt field")
	}
}

// ---- ValidationRunStatus enum/type in types file ----

func TestValidationRunStatusEnumExists(t *testing.T) {
	content := readValidationRunTypes(t)
	if !strings.Contains(content, "ValidationRunStatus") {
		t.Error("validation-run.types.ts must define or reference ValidationRunStatus")
	}
}

func TestValidationRunStatusIncludesCancelled(t *testing.T) {
	content := readValidationRunTypes(t)
	if !strings.Contains(content, "CANCELLED") {
		t.Error("ValidationRunStatus must include CANCELLED status")
	}
}

// ---- validation-run.model.ts: cancel method ----

func TestValidationRunModelHasCancelMethod(t *testing.T) {
	content := readValidationRunModel(t)
	if !strings.Contains(content, "cancel") {
		t.Error("validation-run.model.ts must define a cancel method")
	}
}

func TestValidationRunModelCancelAcceptsId(t *testing.T) {
	content := readValidationRunModel(t)
	// The cancel method must accept an id parameter
	hasIdParam := strings.Contains(content, "cancel(id") ||
		strings.Contains(content, "cancel(id:")
	if !hasIdParam {
		t.Error("cancel method must accept an id parameter")
	}
}

func TestValidationRunModelCancelReturnsPromiseValidationRunRecord(t *testing.T) {
	content := readValidationRunModel(t)
	// The cancel method should return Promise<ValidationRunRecord>
	hasReturnType := strings.Contains(content, "Promise<ValidationRunRecord>")
	if !hasReturnType {
		t.Error("cancel method must return Promise<ValidationRunRecord>")
	}
}

func TestValidationRunModelCancelSetsCancelledAt(t *testing.T) {
	content := readValidationRunModel(t)
	if !strings.Contains(content, "cancelledAt") {
		t.Error("cancel method must set cancelledAt field")
	}
}

func TestValidationRunModelCancelSetsStatusCancelled(t *testing.T) {
	content := readValidationRunModel(t)
	if !strings.Contains(content, "CANCELLED") {
		t.Error("cancel method must set status to CANCELLED")
	}
}

func TestValidationRunModelCancelSetsCancelledAtToNewDate(t *testing.T) {
	content := readValidationRunModel(t)
	// The cancel method must assign cancelledAt: new Date()
	if !strings.Contains(content, "new Date()") {
		t.Error("cancel method must set cancelledAt to new Date()")
	}
}

// ---- validation-run.model.ts: updateStatus method accepts cancelledAt ----

func TestValidationRunModelHasUpdateStatusMethod(t *testing.T) {
	content := readValidationRunModel(t)
	if !strings.Contains(content, "updateStatus") {
		t.Error("validation-run.model.ts must define an updateStatus method")
	}
}

func TestValidationRunModelUpdateStatusAcceptsCancelledAtOptional(t *testing.T) {
	content := readValidationRunModel(t)
	// updateStatus must accept cancelledAt as an optional parameter
	// This can be expressed as cancelledAt?: or in an options object with cancelledAt?:
	if !strings.Contains(content, "cancelledAt") {
		t.Error("updateStatus method must accept cancelledAt as an optional parameter")
	}
}

func TestValidationRunModelUpdateStatusCancelledAtIsOptional(t *testing.T) {
	content := readValidationRunModel(t)
	// The cancelledAt parameter must be marked optional with ?
	hasCancelledAtOptional := strings.Contains(content, "cancelledAt?") ||
		strings.Contains(content, "cancelledAt?: Date") ||
		strings.Contains(content, "cancelledAt?: Date | null")
	if !hasCancelledAtOptional {
		t.Error("cancelledAt parameter in updateStatus must be optional (marked with '?')")
	}
}

// ---- validation-run.model.ts: general structure ----

func TestValidationRunModelImportsOrReferencesValidationRunRecord(t *testing.T) {
	content := readValidationRunModel(t)
	if !strings.Contains(content, "ValidationRunRecord") {
		t.Error("validation-run.model.ts must reference ValidationRunRecord type")
	}
}

func TestValidationRunModelImportsOrReferencesValidationRunStatus(t *testing.T) {
	content := readValidationRunModel(t)
	if !strings.Contains(content, "ValidationRunStatus") {
		t.Error("validation-run.model.ts must reference ValidationRunStatus enum")
	}
}
