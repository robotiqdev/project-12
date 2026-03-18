package version_test

import (
	"regexp"
	"testing"

	"github.com/robotiqdev/project-12/version"
)

// TestVersionIsNonEmpty verifies that the Version constant is not an empty string.
func TestVersionIsNonEmpty(t *testing.T) {
	if version.Version == "" {
		t.Error("version.Version must not be empty")
	}
}

// TestVersionMatchesSemver verifies that the Version constant follows semver format (MAJOR.MINOR.PATCH).
func TestVersionMatchesSemver(t *testing.T) {
	semver := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	if !semver.MatchString(version.Version) {
		t.Errorf("version.Version %q does not match semver format (e.g. 0.1.0)", version.Version)
	}
}
