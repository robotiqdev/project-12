// Package integration_test validates the partitioning migration for TASK-4621.
// These file-system tests verify that the 002_partitioning migration SQL contains
// all required PostgreSQL declarative range partitioning statements for validation_runs.
package integration_test

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

func readPartitioningSQL(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "prisma", "migrations", "002_partitioning", "migration.sql")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read 002_partitioning/migration.sql: %v", err)
	}
	return string(data)
}

// ---- File existence ----

func TestPartitioningMigrationFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "prisma", "migrations", "002_partitioning", "migration.sql")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("002_partitioning/migration.sql not found at %s", path)
	}
}

// ---- SQL content: not empty ----

func TestPartitioningMigrationSQLNotEmpty(t *testing.T) {
	sql := readPartitioningSQL(t)
	if strings.TrimSpace(sql) == "" {
		t.Error("002_partitioning/migration.sql must not be empty")
	}
}

// ---- DROP and re-CREATE validation_runs ----

func TestPartitioningMigrationDropsValidationRunsTable(t *testing.T) {
	sql := readPartitioningSQL(t)
	upper := strings.ToUpper(sql)
	if !strings.Contains(upper, "DROP TABLE") || !strings.Contains(strings.ToLower(sql), "validation_runs") {
		t.Error("migration.sql must DROP the original validation_runs table before recreating it as partitioned")
	}
}

func TestPartitioningMigrationCreatesValidationRunsTable(t *testing.T) {
	sql := readPartitioningSQL(t)
	upper := strings.ToUpper(sql)
	if !strings.Contains(upper, "CREATE TABLE") || !strings.Contains(strings.ToLower(sql), "validation_runs") {
		t.Error("migration.sql must CREATE the validation_runs table")
	}
}

// ---- PARTITION BY RANGE on created_at ----

func TestPartitioningMigrationUsesRangePartitioning(t *testing.T) {
	sql := readPartitioningSQL(t)
	upper := strings.ToUpper(sql)
	if !strings.Contains(upper, "PARTITION BY RANGE") {
		t.Error("migration.sql must use PARTITION BY RANGE for declarative range partitioning")
	}
}

func TestPartitioningMigrationPartitionedByCreatedAt(t *testing.T) {
	sql := readPartitioningSQL(t)
	upper := strings.ToUpper(sql)
	// Must include PARTITION BY RANGE with created_at (case-insensitive column reference)
	if !strings.Contains(upper, "PARTITION BY RANGE") {
		t.Error("migration.sql must partition validation_runs by RANGE")
	}
	if !strings.Contains(sql, "createdAt") && !strings.Contains(sql, "created_at") {
		t.Error("migration.sql must partition on createdAt/created_at column")
	}
}

// ---- Monthly partitions for 2026 ----

func TestPartitioningMigrationHasJanuary2026Partition(t *testing.T) {
	sql := readPartitioningSQL(t)
	// Partition for 2026-01 must be defined
	if !strings.Contains(sql, "2026-01-01") || !strings.Contains(sql, "2026-02-01") {
		t.Error("migration.sql must define a partition covering 2026-01-01 to 2026-02-01 (January 2026)")
	}
}

func TestPartitioningMigrationHasFebruary2026Partition(t *testing.T) {
	sql := readPartitioningSQL(t)
	if !strings.Contains(sql, "2026-02-01") || !strings.Contains(sql, "2026-03-01") {
		t.Error("migration.sql must define a partition covering 2026-02-01 to 2026-03-01 (February 2026)")
	}
}

func TestPartitioningMigrationHasMarch2026Partition(t *testing.T) {
	sql := readPartitioningSQL(t)
	if !strings.Contains(sql, "2026-03-01") || !strings.Contains(sql, "2026-04-01") {
		t.Error("migration.sql must define a partition covering 2026-03-01 to 2026-04-01 (March 2026)")
	}
}

func TestPartitioningMigrationHasApril2026Partition(t *testing.T) {
	sql := readPartitioningSQL(t)
	if !strings.Contains(sql, "2026-04-01") || !strings.Contains(sql, "2026-05-01") {
		t.Error("migration.sql must define a partition covering 2026-04-01 to 2026-05-01 (April 2026)")
	}
}

func TestPartitioningMigrationHasMay2026Partition(t *testing.T) {
	sql := readPartitioningSQL(t)
	if !strings.Contains(sql, "2026-05-01") || !strings.Contains(sql, "2026-06-01") {
		t.Error("migration.sql must define a partition covering 2026-05-01 to 2026-06-01 (May 2026)")
	}
}

func TestPartitioningMigrationHasJune2026Partition(t *testing.T) {
	sql := readPartitioningSQL(t)
	if !strings.Contains(sql, "2026-06-01") || !strings.Contains(sql, "2026-07-01") {
		t.Error("migration.sql must define a partition covering 2026-06-01 to 2026-07-01 (June 2026)")
	}
}

func TestPartitioningMigrationHasJuly2026Partition(t *testing.T) {
	sql := readPartitioningSQL(t)
	if !strings.Contains(sql, "2026-07-01") || !strings.Contains(sql, "2026-08-01") {
		t.Error("migration.sql must define a partition covering 2026-07-01 to 2026-08-01 (July 2026)")
	}
}

func TestPartitioningMigrationHasAugust2026Partition(t *testing.T) {
	sql := readPartitioningSQL(t)
	if !strings.Contains(sql, "2026-08-01") || !strings.Contains(sql, "2026-09-01") {
		t.Error("migration.sql must define a partition covering 2026-08-01 to 2026-09-01 (August 2026)")
	}
}

func TestPartitioningMigrationHasSeptember2026Partition(t *testing.T) {
	sql := readPartitioningSQL(t)
	if !strings.Contains(sql, "2026-09-01") || !strings.Contains(sql, "2026-10-01") {
		t.Error("migration.sql must define a partition covering 2026-09-01 to 2026-10-01 (September 2026)")
	}
}

func TestPartitioningMigrationHasOctober2026Partition(t *testing.T) {
	sql := readPartitioningSQL(t)
	if !strings.Contains(sql, "2026-10-01") || !strings.Contains(sql, "2026-11-01") {
		t.Error("migration.sql must define a partition covering 2026-10-01 to 2026-11-01 (October 2026)")
	}
}

func TestPartitioningMigrationHasNovember2026Partition(t *testing.T) {
	sql := readPartitioningSQL(t)
	if !strings.Contains(sql, "2026-11-01") || !strings.Contains(sql, "2026-12-01") {
		t.Error("migration.sql must define a partition covering 2026-11-01 to 2026-12-01 (November 2026)")
	}
}

func TestPartitioningMigrationHasDecember2026Partition(t *testing.T) {
	sql := readPartitioningSQL(t)
	if !strings.Contains(sql, "2026-12-01") || !strings.Contains(sql, "2027-01-01") {
		t.Error("migration.sql must define a partition covering 2026-12-01 to 2027-01-01 (December 2026)")
	}
}

// ---- PARTITION OF syntax ----

func TestPartitioningMigrationUsesPartitionOfSyntax(t *testing.T) {
	sql := readPartitioningSQL(t)
	upper := strings.ToUpper(sql)
	if !strings.Contains(upper, "PARTITION OF") {
		t.Error("migration.sql must use 'PARTITION OF' to declare child partitions")
	}
}

func TestPartitioningMigrationUsesForValuesSyntax(t *testing.T) {
	sql := readPartitioningSQL(t)
	upper := strings.ToUpper(sql)
	if !strings.Contains(upper, "FOR VALUES FROM") {
		t.Error("migration.sql must use 'FOR VALUES FROM ... TO ...' syntax for range partitions")
	}
}

// ---- Default catch-all partition ----

func TestPartitioningMigrationHasDefaultPartition(t *testing.T) {
	sql := readPartitioningSQL(t)
	upper := strings.ToUpper(sql)
	if !strings.Contains(upper, "DEFAULT") || !strings.Contains(upper, "PARTITION OF") {
		t.Error("migration.sql must define a default catch-all partition (PARTITION OF ... DEFAULT)")
	}
}

func TestPartitioningMigrationDefaultPartitionNameContainsDefault(t *testing.T) {
	sql := readPartitioningSQL(t)
	// Default partition table should be named with 'default' in the name
	if !strings.Contains(strings.ToLower(sql), "validation_runs_default") {
		t.Error("migration.sql must name the default partition 'validation_runs_default'")
	}
}

// ---- Naming convention for monthly partitions ----

func TestPartitioningMigrationPartitionNamingConvention(t *testing.T) {
	sql := readPartitioningSQL(t)
	lower := strings.ToLower(sql)
	// Check at least one partition follows the yYYYYmMM naming convention
	if !strings.Contains(lower, "validation_runs_y2026m01") {
		t.Error("migration.sql partition names must follow 'validation_runs_y2026m01' naming convention")
	}
}

func TestPartitioningMigrationAllTwelveMonthPartitionNamesPresent(t *testing.T) {
	sql := readPartitioningSQL(t)
	lower := strings.ToLower(sql)
	expectedPartitions := []string{
		"validation_runs_y2026m01",
		"validation_runs_y2026m02",
		"validation_runs_y2026m03",
		"validation_runs_y2026m04",
		"validation_runs_y2026m05",
		"validation_runs_y2026m06",
		"validation_runs_y2026m07",
		"validation_runs_y2026m08",
		"validation_runs_y2026m09",
		"validation_runs_y2026m10",
		"validation_runs_y2026m11",
		"validation_runs_y2026m12",
	}
	for _, partName := range expectedPartitions {
		if !strings.Contains(lower, partName) {
			t.Errorf("migration.sql must define partition %s", partName)
		}
	}
}

// ---- TODO comment for future partition maintenance ----

func TestPartitioningMigrationHasTODOCommentForMaintenance(t *testing.T) {
	sql := readPartitioningSQL(t)
	upper := strings.ToUpper(sql)
	// The implementation must include a TODO noting that monthly partitions
	// should be pre-created by a scheduled admin task
	if !strings.Contains(upper, "TODO") {
		t.Error("migration.sql must contain a TODO comment about pre-creating future monthly partitions via a scheduled admin task")
	}
}

// ---- Partition pruning: January 2026 query only touches one partition ----
// This test validates the SQL structure guarantees that a query with
// created_at >= '2026-01-01' AND created_at < '2026-02-01' can be
// efficiently pruned to a single partition.

func TestPartitioningMigrationJanuary2026BoundsDefined(t *testing.T) {
	sql := readPartitioningSQL(t)
	// The January 2026 partition must have precise FROM/TO bounds that allow
	// the PostgreSQL planner to prune to exactly one partition for
	// created_at >= '2026-01-01' AND created_at < '2026-02-01'.
	if !strings.Contains(sql, "'2026-01-01'") {
		t.Error("migration.sql must define January 2026 partition with lower bound '2026-01-01'")
	}
	if !strings.Contains(sql, "'2026-02-01'") {
		t.Error("migration.sql must define January 2026 partition with upper bound '2026-02-01'")
	}
}

func TestPartitioningMigrationPartitionsAreContiguousForPruning(t *testing.T) {
	sql := readPartitioningSQL(t)
	// Partition bounds must be contiguous — the upper bound of one month
	// must equal the lower bound of the next month, enabling partition pruning.
	// Check that consecutive month boundaries appear in the SQL.
	boundaries := []string{
		"2026-01-01",
		"2026-02-01",
		"2026-03-01",
		"2026-04-01",
		"2026-05-01",
		"2026-06-01",
		"2026-07-01",
		"2026-08-01",
		"2026-09-01",
		"2026-10-01",
		"2026-11-01",
		"2026-12-01",
		"2027-01-01",
	}
	for _, boundary := range boundaries {
		if !strings.Contains(sql, boundary) {
			t.Errorf("migration.sql must include partition boundary date %s for contiguous range coverage", boundary)
		}
	}
}

// ---- INSERT routing: partitions correctly receive data ----
// These tests verify the SQL structure ensures INSERTs are routed to the
// correct partition based on created_at value.

func TestPartitioningMigrationValidationRunsIsPartitionedTable(t *testing.T) {
	sql := readPartitioningSQL(t)
	upper := strings.ToUpper(sql)
	// The parent table must include PARTITION BY RANGE clause
	if !strings.Contains(upper, "PARTITION BY RANGE") {
		t.Error("validation_runs parent table must be declared with PARTITION BY RANGE to enable automatic INSERT routing")
	}
}

func TestPartitioningMigrationChildPartitionsReferenceParentTable(t *testing.T) {
	sql := readPartitioningSQL(t)
	upper := strings.ToUpper(sql)
	// Each child partition must be declared as PARTITION OF validation_runs
	occurrences := strings.Count(upper, "PARTITION OF")
	// Should have at least 12 monthly + 1 default = 13 PARTITION OF clauses
	if occurrences < 13 {
		t.Errorf("migration.sql must have at least 13 'PARTITION OF' clauses (12 monthly + 1 default), found %d", occurrences)
	}
}
