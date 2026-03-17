// Package services_test validates the TypeScript source files for TASK-4592.
// These are integration-style file-system tests that verify the types,
// models, and concurrency service files contain all required definitions.
//
// Tests WILL FAIL until the implementation files are created — this is the
// intended TDD behaviour.
package services_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot resolves the repository root from the current working directory.
// Tests are run with `go test ./...` from the repo root, so we walk up to
// find go.mod.
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

// readFile reads a file relative to the repo root or fails the test.
func readFile(t *testing.T, relPath string) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), relPath)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", relPath, err)
	}
	return string(data)
}

// fileExists returns true if the file exists at the given repo-root-relative path.
func fileExists(t *testing.T, relPath string) bool {
	t.Helper()
	path := filepath.Join(repoRoot(t), relPath)
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// ── File existence ────────────────────────────────────────────────────────────

func TestValidationRunTypesFileExists(t *testing.T) {
	if !fileExists(t, "backend/src/types/validation-run.types.ts") {
		t.Fatal("backend/src/types/validation-run.types.ts does not exist")
	}
}

func TestValidationRunModelFileExists(t *testing.T) {
	if !fileExists(t, "backend/src/models/validation-run.model.ts") {
		t.Fatal("backend/src/models/validation-run.model.ts does not exist")
	}
}

func TestRunStageModelFileExists(t *testing.T) {
	if !fileExists(t, "backend/src/models/run-stage.model.ts") {
		t.Fatal("backend/src/models/run-stage.model.ts does not exist")
	}
}

func TestConcurrencyServiceFileExists(t *testing.T) {
	if !fileExists(t, "backend/src/services/concurrency.service.ts") {
		t.Fatal("backend/src/services/concurrency.service.ts does not exist")
	}
}

func TestConcurrencyServiceTestFileExists(t *testing.T) {
	if !fileExists(t, "backend/tests/unit/services/concurrency.service.test.ts") {
		t.Fatal("backend/tests/unit/services/concurrency.service.test.ts does not exist")
	}
}

// ── validation-run.types.ts ───────────────────────────────────────────────────

func TestValidationRunStatusEnumDefined(t *testing.T) {
	src := readFile(t, "backend/src/types/validation-run.types.ts")
	if !strings.Contains(src, "ValidationRunStatus") {
		t.Error("validation-run.types.ts must define ValidationRunStatus enum")
	}
}

func TestValidationRunStatusEnumMembers(t *testing.T) {
	src := readFile(t, "backend/src/types/validation-run.types.ts")
	for _, member := range []string{"PENDING", "RUNNING", "SUCCESS", "FAILED", "CANCELLED"} {
		if !strings.Contains(src, member) {
			t.Errorf("ValidationRunStatus must include member %s", member)
		}
	}
}

func TestStageTypeEnumDefined(t *testing.T) {
	src := readFile(t, "backend/src/types/validation-run.types.ts")
	if !strings.Contains(src, "StageType") {
		t.Error("validation-run.types.ts must define StageType enum")
	}
}

func TestStageTypeEnumMembers(t *testing.T) {
	src := readFile(t, "backend/src/types/validation-run.types.ts")
	for _, member := range []string{"SETUP", "LINT", "TEST", "BUILD", "DEPLOY"} {
		if !strings.Contains(src, member) {
			t.Errorf("StageType enum must include member %s", member)
		}
	}
}

func TestStageStatusEnumDefined(t *testing.T) {
	src := readFile(t, "backend/src/types/validation-run.types.ts")
	if !strings.Contains(src, "StageStatus") {
		t.Error("validation-run.types.ts must define StageStatus enum")
	}
}

func TestStageStatusEnumMembers(t *testing.T) {
	src := readFile(t, "backend/src/types/validation-run.types.ts")
	for _, member := range []string{"PENDING", "RUNNING", "SUCCESS", "FAILED", "CANCELLED"} {
		if !strings.Contains(src, member) {
			t.Errorf("StageStatus enum must include member %s", member)
		}
	}
}

func TestStageOrderConstantDefined(t *testing.T) {
	src := readFile(t, "backend/src/types/validation-run.types.ts")
	if !strings.Contains(src, "STAGE_ORDER") {
		t.Error("validation-run.types.ts must export STAGE_ORDER constant")
	}
}

func TestCreateValidationRunInputInterfaceDefined(t *testing.T) {
	src := readFile(t, "backend/src/types/validation-run.types.ts")
	if !strings.Contains(src, "CreateValidationRunInput") {
		t.Error("validation-run.types.ts must export CreateValidationRunInput interface")
	}
}

func TestCreateValidationRunInputHasRequiredFields(t *testing.T) {
	src := readFile(t, "backend/src/types/validation-run.types.ts")
	for _, field := range []string{"branchName", "commitSha", "userId", "config"} {
		if !strings.Contains(src, field) {
			t.Errorf("CreateValidationRunInput must include field %s", field)
		}
	}
}

func TestRunConfigInterfaceDefined(t *testing.T) {
	src := readFile(t, "backend/src/types/validation-run.types.ts")
	if !strings.Contains(src, "RunConfig") {
		t.Error("validation-run.types.ts must export RunConfig interface")
	}
}

func TestValidationRunRecordInterfaceDefined(t *testing.T) {
	src := readFile(t, "backend/src/types/validation-run.types.ts")
	if !strings.Contains(src, "ValidationRunRecord") {
		t.Error("validation-run.types.ts must export ValidationRunRecord interface")
	}
}

func TestValidationRunRecordHasRequiredFields(t *testing.T) {
	src := readFile(t, "backend/src/types/validation-run.types.ts")
	for _, field := range []string{"id", "branchName", "commitSha", "status", "userId", "config", "createdAt", "updatedAt", "completedAt", "cancelledAt"} {
		if !strings.Contains(src, field) {
			t.Errorf("ValidationRunRecord must include field %s", field)
		}
	}
}

func TestRunStageRecordInterfaceDefined(t *testing.T) {
	src := readFile(t, "backend/src/types/validation-run.types.ts")
	if !strings.Contains(src, "RunStageRecord") {
		t.Error("validation-run.types.ts must export RunStageRecord interface")
	}
}

func TestRunStageRecordHasRequiredFields(t *testing.T) {
	src := readFile(t, "backend/src/types/validation-run.types.ts")
	for _, field := range []string{"id", "runId", "stageName", "stageIndex", "status", "logs", "startedAt", "completedAt"} {
		if !strings.Contains(src, field) {
			t.Errorf("RunStageRecord must include field %s", field)
		}
	}
}

// ── validation-run.model.ts ───────────────────────────────────────────────────

func TestValidationRunModelImportsPrisma(t *testing.T) {
	src := readFile(t, "backend/src/models/validation-run.model.ts")
	if !strings.Contains(src, "prisma") {
		t.Error("validation-run.model.ts must import/use the prisma client")
	}
}

func TestValidationRunModelCreateMethod(t *testing.T) {
	src := readFile(t, "backend/src/models/validation-run.model.ts")
	if !strings.Contains(src, "create") {
		t.Error("validation-run.model.ts must export a create method")
	}
}

func TestValidationRunModelFindByIdMethod(t *testing.T) {
	src := readFile(t, "backend/src/models/validation-run.model.ts")
	if !strings.Contains(src, "findById") {
		t.Error("validation-run.model.ts must export a findById method")
	}
}

func TestValidationRunModelFindByIdWithStagesMethod(t *testing.T) {
	src := readFile(t, "backend/src/models/validation-run.model.ts")
	if !strings.Contains(src, "findByIdWithStages") {
		t.Error("validation-run.model.ts must export a findByIdWithStages method")
	}
}

func TestValidationRunModelFindActiveOnBranchMethod(t *testing.T) {
	src := readFile(t, "backend/src/models/validation-run.model.ts")
	if !strings.Contains(src, "findActiveOnBranch") {
		t.Error("validation-run.model.ts must export a findActiveOnBranch method")
	}
}

func TestValidationRunModelUpdateStatusMethod(t *testing.T) {
	src := readFile(t, "backend/src/models/validation-run.model.ts")
	if !strings.Contains(src, "updateStatus") {
		t.Error("validation-run.model.ts must export an updateStatus method")
	}
}

func TestValidationRunModelListMethod(t *testing.T) {
	src := readFile(t, "backend/src/models/validation-run.model.ts")
	if !strings.Contains(src, "list") {
		t.Error("validation-run.model.ts must export a list method")
	}
}

func TestValidationRunModelFindActiveOnBranchChecksPendingStatus(t *testing.T) {
	src := readFile(t, "backend/src/models/validation-run.model.ts")
	if !strings.Contains(src, "PENDING") {
		t.Error("validation-run.model.ts findActiveOnBranch must check for PENDING status")
	}
}

func TestValidationRunModelFindActiveOnBranchChecksRunningStatus(t *testing.T) {
	src := readFile(t, "backend/src/models/validation-run.model.ts")
	if !strings.Contains(src, "RUNNING") {
		t.Error("validation-run.model.ts findActiveOnBranch must check for RUNNING status")
	}
}

// ── run-stage.model.ts ────────────────────────────────────────────────────────

func TestRunStageModelImportsPrisma(t *testing.T) {
	src := readFile(t, "backend/src/models/run-stage.model.ts")
	if !strings.Contains(src, "prisma") {
		t.Error("run-stage.model.ts must import/use the prisma client")
	}
}

func TestRunStageModelCreateAllMethod(t *testing.T) {
	src := readFile(t, "backend/src/models/run-stage.model.ts")
	if !strings.Contains(src, "createAll") {
		t.Error("run-stage.model.ts must export a createAll method")
	}
}

func TestRunStageModelUpdateStageMethod(t *testing.T) {
	src := readFile(t, "backend/src/models/run-stage.model.ts")
	if !strings.Contains(src, "updateStage") {
		t.Error("run-stage.model.ts must export an updateStage method")
	}
}

func TestRunStageModelFindByRunIdMethod(t *testing.T) {
	src := readFile(t, "backend/src/models/run-stage.model.ts")
	if !strings.Contains(src, "findByRunId") {
		t.Error("run-stage.model.ts must export a findByRunId method")
	}
}

func TestRunStageModelCreateAllAcceptsRunId(t *testing.T) {
	src := readFile(t, "backend/src/models/run-stage.model.ts")
	if !strings.Contains(src, "runId") {
		t.Error("run-stage.model.ts createAll must accept a runId parameter")
	}
}

// ── concurrency.service.ts ────────────────────────────────────────────────────

func TestConcurrencyServiceImportsPrisma(t *testing.T) {
	src := readFile(t, "backend/src/services/concurrency.service.ts")
	if !strings.Contains(src, "prisma") {
		t.Error("concurrency.service.ts must import/use the prisma client")
	}
}

func TestConcurrencyServiceExportsCheckAndLockBranch(t *testing.T) {
	src := readFile(t, "backend/src/services/concurrency.service.ts")
	if !strings.Contains(src, "checkAndLockBranch") {
		t.Error("concurrency.service.ts must export checkAndLockBranch function")
	}
}

func TestConcurrencyServiceExportsCreateRun(t *testing.T) {
	src := readFile(t, "backend/src/services/concurrency.service.ts")
	if !strings.Contains(src, "createRun") {
		t.Error("concurrency.service.ts must export createRun function")
	}
}

func TestConcurrencyServiceExportsGetRunWithStages(t *testing.T) {
	src := readFile(t, "backend/src/services/concurrency.service.ts")
	if !strings.Contains(src, "getRunWithStages") {
		t.Error("concurrency.service.ts must export getRunWithStages function")
	}
}

func TestConcurrencyServiceUsesTransaction(t *testing.T) {
	src := readFile(t, "backend/src/services/concurrency.service.ts")
	if !strings.Contains(src, "$transaction") {
		t.Error("concurrency.service.ts checkAndLockBranch must use prisma.$transaction for atomicity")
	}
}

func TestConcurrencyServiceUsesRawSQLForLocking(t *testing.T) {
	src := readFile(t, "backend/src/services/concurrency.service.ts")
	hasQueryRaw := strings.Contains(src, "$queryRaw") || strings.Contains(src, "$executeRaw")
	hasAdvisoryLock := strings.Contains(src, "pg_try_advisory") || strings.Contains(src, "FOR UPDATE")
	if !hasQueryRaw && !hasAdvisoryLock {
		t.Error("concurrency.service.ts must use raw SQL ($queryRaw or advisory lock) to prevent TOCTOU race conditions")
	}
}

func TestConcurrencyServiceCheckAndLockBranchReturnShape(t *testing.T) {
	src := readFile(t, "backend/src/services/concurrency.service.ts")
	if !strings.Contains(src, "allowed") {
		t.Error("concurrency.service.ts checkAndLockBranch must return an object with 'allowed' property")
	}
}

func TestConcurrencyServiceCheckAndLockBranchReturnsConflictingRunId(t *testing.T) {
	src := readFile(t, "backend/src/services/concurrency.service.ts")
	if !strings.Contains(src, "conflictingRunId") {
		t.Error("concurrency.service.ts checkAndLockBranch must return 'conflictingRunId' when a conflict is found")
	}
}

func TestConcurrencyServiceChecksPendingStatus(t *testing.T) {
	src := readFile(t, "backend/src/services/concurrency.service.ts")
	if !strings.Contains(src, "PENDING") {
		t.Error("concurrency.service.ts must check for PENDING status when locking branch")
	}
}

func TestConcurrencyServiceChecksRunningStatus(t *testing.T) {
	src := readFile(t, "backend/src/services/concurrency.service.ts")
	if !strings.Contains(src, "RUNNING") {
		t.Error("concurrency.service.ts must check for RUNNING status when locking branch")
	}
}

func TestConcurrencyServiceCreateRunSetsPendingStatus(t *testing.T) {
	src := readFile(t, "backend/src/services/concurrency.service.ts")
	// createRun must set status to PENDING when creating a new run
	if !strings.Contains(src, "PENDING") {
		t.Error("concurrency.service.ts createRun must set status to PENDING")
	}
}

func TestConcurrencyServiceGetRunWithStagesSortsByStageIndex(t *testing.T) {
	src := readFile(t, "backend/src/services/concurrency.service.ts")
	if !strings.Contains(src, "stageIndex") {
		t.Error("concurrency.service.ts getRunWithStages must sort/order stages by stageIndex")
	}
}
