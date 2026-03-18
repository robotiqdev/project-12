// Package version_test validates TASK-4633: version.go must have godoc comments
// on the exported Version constant and (optionally) a package-level doc comment.
package version_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

// versionFilePath locates version/version.go relative to the repo root.
func versionFilePath(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	// Walk up from current package dir to find go.mod (repo root).
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, "version", "version.go")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find repo root (no go.mod)")
		}
		dir = parent
	}
}

// parseVersionFile parses version.go and returns the AST file node.
func parseVersionFile(t *testing.T) *ast.File {
	t.Helper()
	path := versionFilePath(t)
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse %s: %v", path, err)
	}
	return f
}

// TestVersionFileExists verifies that version/version.go exists.
func TestVersionFileExists(t *testing.T) {
	path := versionFilePath(t)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("version/version.go not found at %s", path)
	}
}

// TestVersionConstantExists verifies that the Version exported constant is declared.
func TestVersionConstantExists(t *testing.T) {
	f := parseVersionFile(t)
	found := false
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}
		for _, spec := range genDecl.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range vs.Names {
				if name.Name == "Version" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("version.go must declare an exported constant named Version")
	}
}

// TestVersionConstantHasDocComment verifies that the Version constant has a
// godoc comment immediately above it, as required by TASK-4633.
func TestVersionConstantHasDocComment(t *testing.T) {
	f := parseVersionFile(t)
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}
		for _, spec := range genDecl.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range vs.Names {
				if name.Name != "Version" {
					continue
				}
				// Check spec-level doc comment first, then group-level.
				doc := vs.Doc
				if doc == nil {
					doc = genDecl.Doc
				}
				if doc == nil || len(doc.List) == 0 {
					t.Error("Version constant must have a godoc comment immediately above it")
					return
				}
				combined := ""
				for _, c := range doc.List {
					combined += c.Text + " "
				}
				if !strings.Contains(combined, "Version") {
					t.Errorf("Version doc comment should mention 'Version'; got: %s", combined)
				}
				return
			}
		}
	}
	t.Error("Version constant not found while checking for doc comment")
}

// TestVersionConstantDocMentionsRelease verifies the doc comment contains
// information about the release version, as specified in the task.
func TestVersionConstantDocMentionsRelease(t *testing.T) {
	f := parseVersionFile(t)
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}
		for _, spec := range genDecl.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range vs.Names {
				if name.Name != "Version" {
					continue
				}
				doc := vs.Doc
				if doc == nil {
					doc = genDecl.Doc
				}
				if doc == nil || len(doc.List) == 0 {
					t.Error("Version constant must have a godoc comment")
					return
				}
				combined := ""
				for _, c := range doc.List {
					combined += c.Text + " "
				}
				lc := strings.ToLower(combined)
				if !strings.Contains(lc, "release") && !strings.Contains(lc, "version") {
					t.Errorf("Version doc comment should describe the release version; got: %s", combined)
				}
				return
			}
		}
	}
}

// TestPackageHasDocComment verifies that the version package itself has a
// package-level doc comment (optional but recommended per the task spec).
func TestPackageHasDocComment(t *testing.T) {
	f := parseVersionFile(t)
	if f.Doc == nil || len(f.Doc.List) == 0 {
		t.Error("version package should have a package-level doc comment above 'package version'")
	}
}
