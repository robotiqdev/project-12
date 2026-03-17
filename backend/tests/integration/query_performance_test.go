// Package integration_test validates the performance index migration for TASK-4622.
//
// These tests verify that backend/prisma/migrations/003_indexes/migration.sql
// exists and contains all required CREATE INDEX statements — including the
// composite indexes on ValidationRun, RunStage, and ValidationPreset, as well
// as the partial index on active runs.
package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot walks up from the current working directory until go.mod is found.
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

func indexesMigrationPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(
		repoRoot(t),
		"backend", "prisma", "migrations", "003_indexes", "migration.sql",
	)
}

func readIndexesMigration(t *testing.T) string {
	t.Helper()
	path := indexesMigrationPath(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf(
			"failed to read 003_indexes/migration.sql (has the migration been created?): %v", err,
		)
	}
	return string(data)
}

// ---- File existence ----

func TestIndexesMigrationFileExists(t *testing.T) {
	path := indexesMigrationPath(t)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf(
			"003_indexes/migration.sql not found at %s — run `npx prisma migrate dev --name add_indexes`",
			path,
		)
	}
}

func TestIndexesMigrationNotEmpty(t *testing.T) {
	sql := readIndexesMigration(t)
	if strings.TrimSpace(sql) == "" {
		t.Error("003_indexes/migration.sql must not be empty")
	}
}

// ---- ValidationRun composite indexes ----

func TestIndexesMigrationValidationRunsBranchNameStatus(t *testing.T) {
	sql := readIndexesMigration(t)
	upper := strings.ToUpper(sql)
	// Must contain a CREATE INDEX on validation_runs covering (branchName, status)
	if !strings.Contains(upper, "CREATE INDEX") {
		t.Fatal("003_indexes/migration.sql must contain CREATE INDEX statements")
	}
	if !strings.Contains(sql, "validation_runs") {
		t.Error("003_indexes/migration.sql must reference validation_runs table")
	}
	if !strings.Contains(sql, `"branchName"`) && !strings.Contains(sql, "branch_name") {
		t.Error(
			"003_indexes/migration.sql must include an index with branchName column on validation_runs",
		)
	}
	if !strings.Contains(sql, `"status"`) && !strings.Contains(sql, "status") {
		t.Error(
			"003_indexes/migration.sql must include an index with status column on validation_runs",
		)
	}
}

func TestIndexesMigrationValidationRunsBranchNameStatusIndexName(t *testing.T) {
	sql := readIndexesMigration(t)
	// Prisma generates index names like validation_runs_branchName_status_idx
	if !strings.Contains(sql, "validation_runs_branchName_status_idx") {
		t.Error(
			"003_indexes/migration.sql must define index named " +
				"validation_runs_branchName_status_idx on validation_runs(branchName, status)",
		)
	}
}

func TestIndexesMigrationValidationRunsBranchNameCommitSha(t *testing.T) {
	sql := readIndexesMigration(t)
	if !strings.Contains(sql, "validation_runs_branchName_commitSha_idx") {
		t.Error(
			"003_indexes/migration.sql must define index named " +
				"validation_runs_branchName_commitSha_idx on validation_runs(branchName, commitSha)",
		)
	}
}

func TestIndexesMigrationValidationRunsBranchNameCreatedAt(t *testing.T) {
	sql := readIndexesMigration(t)
	if !strings.Contains(sql, "validation_runs_branchName_createdAt_idx") {
		t.Error(
			"003_indexes/migration.sql must define index named " +
				"validation_runs_branchName_createdAt_idx on validation_runs(branchName, createdAt DESC)",
		)
	}
}

func TestIndexesMigrationValidationRunsBranchNameCreatedAtDescOrdering(t *testing.T) {
	sql := readIndexesMigration(t)
	upper := strings.ToUpper(sql)
	// The createdAt column in this index must use DESC ordering
	if !strings.Contains(upper, "DESC") {
		t.Error(
			"003_indexes/migration.sql must specify DESC ordering for createdAt " +
				"in the validation_runs(branchName, createdAt DESC) index",
		)
	}
}

func TestIndexesMigrationValidationRunsUserIdCreatedAt(t *testing.T) {
	sql := readIndexesMigration(t)
	if !strings.Contains(sql, "validation_runs_userId_createdAt_idx") {
		t.Error(
			"003_indexes/migration.sql must define index named " +
				"validation_runs_userId_createdAt_idx on validation_runs(userId, createdAt DESC)",
		)
	}
}

// ---- RunStage indexes ----

func TestIndexesMigrationRunStagesRunIdStageIndexIdx(t *testing.T) {
	sql := readIndexesMigration(t)
	if !strings.Contains(sql, "run_stages_runId_stageIndex_idx") {
		t.Error(
			"003_indexes/migration.sql must define index named " +
				"run_stages_runId_stageIndex_idx on run_stages(runId, stageIndex)",
		)
	}
}

func TestIndexesMigrationRunStagesRunIdStageIndexUnique(t *testing.T) {
	sql := readIndexesMigration(t)
	upper := strings.ToUpper(sql)
	// A UNIQUE index on (runId, stageIndex) must be present
	if !strings.Contains(upper, "CREATE UNIQUE INDEX") {
		t.Error(
			"003_indexes/migration.sql must define a UNIQUE index for " +
				"run_stages(runId, stageIndex)",
		)
	}
	if !strings.Contains(sql, "run_stages_runId_stageIndex_key") {
		t.Error(
			"003_indexes/migration.sql must define unique index named " +
				"run_stages_runId_stageIndex_key on run_stages(runId, stageIndex)",
		)
	}
}

// ---- ValidationPreset indexes ----

func TestIndexesMigrationValidationPresetsUserIdCreatedAt(t *testing.T) {
	sql := readIndexesMigration(t)
	if !strings.Contains(sql, "validation_presets_userId_createdAt_idx") {
		t.Error(
			"003_indexes/migration.sql must define index named " +
				"validation_presets_userId_createdAt_idx on validation_presets(userId, createdAt DESC)",
		)
	}
}

// ---- Partial index for active runs ----

func TestIndexesMigrationPartialIndexActiveRunsExists(t *testing.T) {
	sql := readIndexesMigration(t)
	if !strings.Contains(sql, "idx_runs_active") {
		t.Error(
			"003_indexes/migration.sql must define partial index idx_runs_active " +
				"on validation_runs(branch_name, created_at DESC) WHERE status IN ('PENDING', 'RUNNING')",
		)
	}
}

func TestIndexesMigrationPartialIndexActiveRunsHasWhereClause(t *testing.T) {
	sql := readIndexesMigration(t)
	upper := strings.ToUpper(sql)
	// The partial index must have a WHERE clause
	if !strings.Contains(upper, "WHERE") {
		t.Error(
			"003_indexes/migration.sql partial index idx_runs_active must include a WHERE clause " +
				"to filter on active statuses (PENDING, RUNNING)",
		)
	}
}

func TestIndexesMigrationPartialIndexActiveRunsFiltersPendingAndRunning(t *testing.T) {
	sql := readIndexesMigration(t)
	if !strings.Contains(sql, "PENDING") || !strings.Contains(sql, "RUNNING") {
		t.Error(
			"003_indexes/migration.sql partial index idx_runs_active must filter on " +
				"status IN ('PENDING', 'RUNNING')",
		)
	}
}

func TestIndexesMigrationPartialIndexActiveRunsOnValidationRuns(t *testing.T) {
	sql := readIndexesMigration(t)
	// The partial index must reference validation_runs table
	linesWithActiveIdx := []string{}
	for _, line := range strings.Split(sql, "\n") {
		if strings.Contains(line, "idx_runs_active") {
			linesWithActiveIdx = append(linesWithActiveIdx, line)
		}
	}
	if len(linesWithActiveIdx) == 0 {
		t.Fatal("idx_runs_active not found in migration SQL")
	}
	found := false
	for _, line := range linesWithActiveIdx {
		if strings.Contains(line, "validation_runs") {
			found = true
			break
		}
	}
	if !found {
		t.Error(
			"003_indexes/migration.sql partial index idx_runs_active must be on the validation_runs table",
		)
	}
}

func TestIndexesMigrationPartialIndexActiveRunsCoversCreatedAtDesc(t *testing.T) {
	sql := readIndexesMigration(t)
	// Find the CREATE INDEX line for idx_runs_active
	for _, line := range strings.Split(sql, "\n") {
		if strings.Contains(line, "idx_runs_active") {
			upper := strings.ToUpper(line)
			if !strings.Contains(upper, "DESC") {
				t.Errorf(
					"idx_runs_active index should include DESC ordering for created_at; got: %s", line,
				)
			}
			return
		}
	}
	// idx_runs_active not found — already caught by TestIndexesMigrationPartialIndexActiveRunsExists
}

// ---- Column ordering (highest-selectivity column first) ----

func TestIndexesMigrationBranchNamePrecedesStatusInIndex(t *testing.T) {
	sql := readIndexesMigration(t)
	// In the branchName+status index definition, branchName must come before status
	for _, line := range strings.Split(sql, "\n") {
		if strings.Contains(line, "validation_runs_branchName_status_idx") {
			branchPos := strings.Index(line, "branchName")
			statusPos := strings.Index(line, "status")
			if branchPos == -1 || statusPos == -1 {
				t.Errorf(
					"validation_runs_branchName_status_idx definition missing expected columns: %s", line,
				)
				return
			}
			if branchPos > statusPos {
				t.Errorf(
					"branchName must precede status in validation_runs_branchName_status_idx; got: %s",
					line,
				)
			}
			return
		}
	}
	t.Error("validation_runs_branchName_status_idx line not found in migration SQL")
}

func TestIndexesMigrationBranchNamePrecedesCreatedAtInIndex(t *testing.T) {
	sql := readIndexesMigration(t)
	for _, line := range strings.Split(sql, "\n") {
		if strings.Contains(line, "validation_runs_branchName_createdAt_idx") {
			branchPos := strings.Index(line, "branchName")
			createdAtPos := strings.Index(line, "createdAt")
			if branchPos == -1 || createdAtPos == -1 {
				t.Errorf(
					"validation_runs_branchName_createdAt_idx definition missing expected columns: %s",
					line,
				)
				return
			}
			if branchPos > createdAtPos {
				t.Errorf(
					"branchName must precede createdAt in validation_runs_branchName_createdAt_idx; got: %s",
					line,
				)
			}
			return
		}
	}
	t.Error("validation_runs_branchName_createdAt_idx line not found in migration SQL")
}

func TestIndexesMigrationUserIdPrecedesCreatedAtInIndex(t *testing.T) {
	sql := readIndexesMigration(t)
	for _, line := range strings.Split(sql, "\n") {
		if strings.Contains(line, "validation_runs_userId_createdAt_idx") {
			userIdPos := strings.Index(line, "userId")
			createdAtPos := strings.Index(line, "createdAt")
			if userIdPos == -1 || createdAtPos == -1 {
				t.Errorf(
					"validation_runs_userId_createdAt_idx definition missing expected columns: %s", line,
				)
				return
			}
			if userIdPos > createdAtPos {
				t.Errorf(
					"userId must precede createdAt in validation_runs_userId_createdAt_idx; got: %s",
					line,
				)
			}
			return
		}
	}
	t.Error("validation_runs_userId_createdAt_idx line not found in migration SQL")
}
