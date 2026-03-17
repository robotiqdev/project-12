// Package preset_test validates the preset service implementation files for TASK-4601.
// These are integration-style file-system tests that verify the types, model, and service
// TypeScript files exist and contain all required definitions.
//
// These tests SHOULD FAIL until the implementation is created.
package preset_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot resolves the repository root from the current working directory.
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

func readPresetTypesTS(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "src", "types", "preset.types.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read preset.types.ts: %v", err)
	}
	return string(data)
}

func readValidationPresetModelTS(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "src", "models", "validation-preset.model.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read validation-preset.model.ts: %v", err)
	}
	return string(data)
}

func readPresetServiceTS(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "src", "services", "preset.service.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read preset.service.ts: %v", err)
	}
	return string(data)
}

func readPresetServiceTestTS(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "tests", "unit", "services", "preset.service.test.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read preset.service.test.ts: %v", err)
	}
	return string(data)
}

// ---- File existence ----

func TestPresetTypesFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "src", "types", "preset.types.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("preset.types.ts not found at %s", path)
	}
}

func TestValidationPresetModelFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "src", "models", "validation-preset.model.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("validation-preset.model.ts not found at %s", path)
	}
}

func TestPresetServiceFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "src", "services", "preset.service.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("preset.service.ts not found at %s", path)
	}
}

func TestPresetServiceTestFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "tests", "unit", "services", "preset.service.test.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("preset.service.test.ts not found at %s", path)
	}
}

// ---- preset.types.ts ----

func TestPresetTypesDefinesCreatePresetInput(t *testing.T) {
	content := readPresetTypesTS(t)
	if !strings.Contains(content, "CreatePresetInput") {
		t.Error("preset.types.ts must define CreatePresetInput interface")
	}
}

func TestPresetTypesDefinesUpdatePresetInput(t *testing.T) {
	content := readPresetTypesTS(t)
	if !strings.Contains(content, "UpdatePresetInput") {
		t.Error("preset.types.ts must define UpdatePresetInput interface")
	}
}

func TestPresetTypesDefinesValidationPresetRecord(t *testing.T) {
	content := readPresetTypesTS(t)
	if !strings.Contains(content, "ValidationPresetRecord") {
		t.Error("preset.types.ts must define ValidationPresetRecord interface")
	}
}

func TestPresetTypesDefinesPresetMaxPerUserConstant(t *testing.T) {
	content := readPresetTypesTS(t)
	if !strings.Contains(content, "PRESET_MAX_PER_USER") {
		t.Error("preset.types.ts must define PRESET_MAX_PER_USER constant")
	}
}

func TestPresetMaxPerUserIsExported(t *testing.T) {
	content := readPresetTypesTS(t)
	if !strings.Contains(content, "export") || !strings.Contains(content, "PRESET_MAX_PER_USER") {
		t.Error("PRESET_MAX_PER_USER must be exported from preset.types.ts")
	}
}

func TestPresetTypesHasUserIdField(t *testing.T) {
	content := readPresetTypesTS(t)
	if !strings.Contains(content, "userId") {
		t.Error("preset.types.ts interfaces must include userId field")
	}
}

func TestPresetTypesHasNameField(t *testing.T) {
	content := readPresetTypesTS(t)
	if !strings.Contains(content, "name") {
		t.Error("preset.types.ts interfaces must include name field")
	}
}

func TestPresetTypesHasEnvVarsField(t *testing.T) {
	content := readPresetTypesTS(t)
	if !strings.Contains(content, "envVars") {
		t.Error("preset.types.ts interfaces must include envVars field")
	}
}

func TestPresetTypesHasFeatureFlagsField(t *testing.T) {
	content := readPresetTypesTS(t)
	if !strings.Contains(content, "featureFlags") {
		t.Error("preset.types.ts interfaces must include featureFlags field")
	}
}

// ---- validation-preset.model.ts ----

func TestValidationPresetModelDefinesCountByUser(t *testing.T) {
	content := readValidationPresetModelTS(t)
	if !strings.Contains(content, "countByUser") {
		t.Error("validation-preset.model.ts must define countByUser method")
	}
}

func TestValidationPresetModelDefinesCreate(t *testing.T) {
	content := readValidationPresetModelTS(t)
	if !strings.Contains(content, "create") {
		t.Error("validation-preset.model.ts must define create method")
	}
}

func TestValidationPresetModelDefinesFindAllByUser(t *testing.T) {
	content := readValidationPresetModelTS(t)
	if !strings.Contains(content, "findAllByUser") {
		t.Error("validation-preset.model.ts must define findAllByUser method")
	}
}

func TestValidationPresetModelDefinesFindByIdAndUser(t *testing.T) {
	content := readValidationPresetModelTS(t)
	if !strings.Contains(content, "findByIdAndUser") {
		t.Error("validation-preset.model.ts must define findByIdAndUser method")
	}
}

func TestValidationPresetModelDefinesUpdate(t *testing.T) {
	content := readValidationPresetModelTS(t)
	if !strings.Contains(content, "update") {
		t.Error("validation-preset.model.ts must define update method")
	}
}

func TestValidationPresetModelDefinesDelete(t *testing.T) {
	content := readValidationPresetModelTS(t)
	if !strings.Contains(content, "delete") {
		t.Error("validation-preset.model.ts must define delete method")
	}
}

func TestValidationPresetModelExportsValidationPresetModel(t *testing.T) {
	content := readValidationPresetModelTS(t)
	if !strings.Contains(content, "export") || !strings.Contains(content, "ValidationPresetModel") {
		t.Error("validation-preset.model.ts must export ValidationPresetModel")
	}
}

func TestValidationPresetModelImportsPrismaClient(t *testing.T) {
	content := readValidationPresetModelTS(t)
	if !strings.Contains(content, "prisma") {
		t.Error("validation-preset.model.ts must use the prisma client")
	}
}

// ---- preset.service.ts ----

func TestPresetServiceDefinesPresetLimitExceededError(t *testing.T) {
	content := readPresetServiceTS(t)
	if !strings.Contains(content, "PresetLimitExceededError") {
		t.Error("preset.service.ts must define PresetLimitExceededError")
	}
}

func TestPresetLimitExceededErrorExtendsError(t *testing.T) {
	content := readPresetServiceTS(t)
	if !strings.Contains(content, "PresetLimitExceededError") || !strings.Contains(content, "extends Error") {
		t.Error("PresetLimitExceededError must extend Error")
	}
}

func TestPresetLimitExceededErrorHasStatusCode(t *testing.T) {
	content := readPresetServiceTS(t)
	if !strings.Contains(content, "statusCode") {
		t.Error("PresetLimitExceededError must have a statusCode property")
	}
}

func TestPresetServiceDefinesPresetServiceClass(t *testing.T) {
	content := readPresetServiceTS(t)
	if !strings.Contains(content, "PresetService") {
		t.Error("preset.service.ts must define PresetService class or export")
	}
}

func TestPresetServiceDefinesCreatePreset(t *testing.T) {
	content := readPresetServiceTS(t)
	if !strings.Contains(content, "createPreset") {
		t.Error("preset.service.ts must define createPreset function/method")
	}
}

func TestPresetServiceDefinesGetPresets(t *testing.T) {
	content := readPresetServiceTS(t)
	if !strings.Contains(content, "getPresets") {
		t.Error("preset.service.ts must define getPresets function/method")
	}
}

func TestPresetServiceDefinesDeletePreset(t *testing.T) {
	content := readPresetServiceTS(t)
	if !strings.Contains(content, "deletePreset") {
		t.Error("preset.service.ts must define deletePreset function/method")
	}
}

func TestPresetServiceEnforcesMaxPresetLimit(t *testing.T) {
	content := readPresetServiceTS(t)
	if !strings.Contains(content, "countByUser") {
		t.Error("preset.service.ts createPreset must call countByUser to enforce the max-5 limit")
	}
}

func TestPresetServiceThrowsPresetLimitExceededError(t *testing.T) {
	content := readPresetServiceTS(t)
	if !strings.Contains(content, "PresetLimitExceededError") || !strings.Contains(content, "throw") {
		t.Error("preset.service.ts must throw PresetLimitExceededError when limit exceeded")
	}
}

func TestPresetServiceImportsCountFromModel(t *testing.T) {
	content := readPresetServiceTS(t)
	if !strings.Contains(content, "ValidationPresetModel") {
		t.Error("preset.service.ts must use ValidationPresetModel for data access")
	}
}

func TestPresetServiceExportsPresetLimitExceededError(t *testing.T) {
	content := readPresetServiceTS(t)
	if !strings.Contains(content, "export") || !strings.Contains(content, "PresetLimitExceededError") {
		t.Error("preset.service.ts must export PresetLimitExceededError")
	}
}

// ---- TypeScript test file content ----

func TestPresetServiceTestFileContainsCreatePresetTests(t *testing.T) {
	content := readPresetServiceTestTS(t)
	if !strings.Contains(content, "createPreset") {
		t.Error("preset.service.test.ts must contain tests for createPreset")
	}
}

func TestPresetServiceTestFileContainsGetPresetsTests(t *testing.T) {
	content := readPresetServiceTestTS(t)
	if !strings.Contains(content, "getPresets") {
		t.Error("preset.service.test.ts must contain tests for getPresets")
	}
}

func TestPresetServiceTestFileContainsDeletePresetTests(t *testing.T) {
	content := readPresetServiceTestTS(t)
	if !strings.Contains(content, "deletePreset") {
		t.Error("preset.service.test.ts must contain tests for deletePreset")
	}
}

func TestPresetServiceTestFileContainsPresetLimitExceededErrorTest(t *testing.T) {
	content := readPresetServiceTestTS(t)
	if !strings.Contains(content, "PresetLimitExceededError") {
		t.Error("preset.service.test.ts must test PresetLimitExceededError behavior")
	}
}

func TestPresetServiceTestFileMocksPrisma(t *testing.T) {
	content := readPresetServiceTestTS(t)
	if !strings.Contains(content, "jest.mock") && !strings.Contains(content, "Mock") {
		t.Error("preset.service.test.ts must mock the Prisma/model layer")
	}
}

func TestPresetServiceTestFileTestsMaxFivePresets(t *testing.T) {
	content := readPresetServiceTestTS(t)
	if !strings.Contains(content, "5") {
		t.Error("preset.service.test.ts must test the max-5 preset limit")
	}
}

func TestPresetServiceTestFileTestsDeletePreset404(t *testing.T) {
	content := readPresetServiceTestTS(t)
	if !strings.Contains(content, "404") {
		t.Error("preset.service.test.ts must test 404 when preset does not belong to the user")
	}
}

func TestPresetServiceTestFileTestsOrderedByCreatedAt(t *testing.T) {
	content := readPresetServiceTestTS(t)
	if !strings.Contains(content, "createdAt") {
		t.Error("preset.service.test.ts must test that presets are ordered by createdAt")
	}
}
