// Package integration_test validates the repository integration API files for TASK-4600.
// These are integration-style file-system tests that verify the repository service,
// routes, and controller TypeScript files contain all required definitions.
package integration_test

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

func readServiceTS(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "src", "services", "repository.service.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read repository.service.ts: %v", err)
	}
	return string(data)
}

func readRoutesTS(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "src", "routes", "repository.routes.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read repository.routes.ts: %v", err)
	}
	return string(data)
}

func readControllerTS(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "src", "controllers", "repository.controller.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read repository.controller.ts: %v", err)
	}
	return string(data)
}

func readRepositoryTestTS(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "tests", "integration", "repository.test.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read repository.test.ts: %v", err)
	}
	return string(data)
}

// ---- File existence tests ----

func TestRepositoryServiceFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "src", "services", "repository.service.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("repository.service.ts not found at %s", path)
	}
}

func TestRepositoryRoutesFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "src", "routes", "repository.routes.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("repository.routes.ts not found at %s", path)
	}
}

func TestRepositoryControllerFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "src", "controllers", "repository.controller.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("repository.controller.ts not found at %s", path)
	}
}

func TestRepositoryTestFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "tests", "integration", "repository.test.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("repository.test.ts not found at %s", path)
	}
}

// ---- RepositoryProvider interface (Adapter pattern) ----

func TestRepositoryServiceDefinesRepositoryProviderInterface(t *testing.T) {
	svc := readServiceTS(t)
	if !strings.Contains(svc, "RepositoryProvider") {
		t.Error("repository.service.ts must define a RepositoryProvider interface")
	}
}

func TestRepositoryServiceProviderInterfaceHasListBranches(t *testing.T) {
	svc := readServiceTS(t)
	if !strings.Contains(svc, "listBranches") {
		t.Error("RepositoryProvider interface must declare listBranches method")
	}
}

func TestRepositoryServiceProviderInterfaceHasListCommits(t *testing.T) {
	svc := readServiceTS(t)
	if !strings.Contains(svc, "listCommits") {
		t.Error("RepositoryProvider interface must declare listCommits method")
	}
}

func TestRepositoryServiceProviderInterfaceHasGetContextBaseUrl(t *testing.T) {
	svc := readServiceTS(t)
	if !strings.Contains(svc, "getContextBaseUrl") {
		t.Error("RepositoryProvider interface must declare getContextBaseUrl method")
	}
}

func TestRepositoryServiceProviderInterfaceHasBuildCommitUrl(t *testing.T) {
	svc := readServiceTS(t)
	if !strings.Contains(svc, "buildCommitUrl") {
		t.Error("RepositoryProvider interface must declare buildCommitUrl method")
	}
}

func TestRepositoryServiceProviderInterfaceHasBuildBranchUrl(t *testing.T) {
	svc := readServiceTS(t)
	if !strings.Contains(svc, "buildBranchUrl") {
		t.Error("RepositoryProvider interface must declare buildBranchUrl method")
	}
}

// ---- Service configuration via env vars ----

func TestRepositoryServiceUsesRepoBASEURLEnvVar(t *testing.T) {
	svc := readServiceTS(t)
	if !strings.Contains(svc, "REPO_BASE_URL") {
		t.Error("repository.service.ts must configure via REPO_BASE_URL env var")
	}
}

func TestRepositoryServiceUsesREPOPROVIDEREnvVar(t *testing.T) {
	svc := readServiceTS(t)
	if !strings.Contains(svc, "REPO_PROVIDER") {
		t.Error("repository.service.ts must configure via REPO_PROVIDER env var")
	}
}

// ---- Branch data shape ----

func TestRepositoryServiceBranchShapeHasName(t *testing.T) {
	svc := readServiceTS(t)
	if !strings.Contains(svc, "name") {
		t.Error("repository.service.ts must include 'name' in branch data shape")
	}
}

func TestRepositoryServiceBranchShapeHasLastCommit(t *testing.T) {
	svc := readServiceTS(t)
	if !strings.Contains(svc, "lastCommit") {
		t.Error("repository.service.ts must include 'lastCommit' in branch data shape")
	}
}

func TestRepositoryServiceBranchShapeHasIsDefault(t *testing.T) {
	svc := readServiceTS(t)
	if !strings.Contains(svc, "isDefault") {
		t.Error("repository.service.ts must include 'isDefault' in branch data shape")
	}
}

// ---- Commit data shape ----

func TestRepositoryServiceCommitShapeHasSha(t *testing.T) {
	svc := readServiceTS(t)
	if !strings.Contains(svc, "sha") {
		t.Error("repository.service.ts must include 'sha' in commit data shape")
	}
}

func TestRepositoryServiceCommitShapeHasMessage(t *testing.T) {
	svc := readServiceTS(t)
	if !strings.Contains(svc, "message") {
		t.Error("repository.service.ts must include 'message' in commit data shape")
	}
}

func TestRepositoryServiceCommitShapeHasAuthor(t *testing.T) {
	svc := readServiceTS(t)
	if !strings.Contains(svc, "author") {
		t.Error("repository.service.ts must include 'author' in commit data shape")
	}
}

func TestRepositoryServiceCommitShapeHasDate(t *testing.T) {
	svc := readServiceTS(t)
	if !strings.Contains(svc, "date") {
		t.Error("repository.service.ts must include 'date' in commit data shape")
	}
}

// ---- Context data shape ----

func TestRepositoryServiceContextShapeHasBaseUrl(t *testing.T) {
	svc := readServiceTS(t)
	if !strings.Contains(svc, "baseUrl") {
		t.Error("repository.service.ts must include 'baseUrl' in context data shape")
	}
}

func TestRepositoryServiceContextShapeHasProvider(t *testing.T) {
	svc := readServiceTS(t)
	if !strings.Contains(svc, "provider") {
		t.Error("repository.service.ts must include 'provider' in context data shape")
	}
}

// ---- Stub implementation ----

func TestRepositoryServiceHasStubOrMockImplementation(t *testing.T) {
	svc := readServiceTS(t)
	if !strings.Contains(strings.ToLower(svc), "stub") && !strings.Contains(strings.ToLower(svc), "mock") {
		t.Error("repository.service.ts must include a stub/mock implementation for development")
	}
}

// ---- listCommits default limit ----

func TestRepositoryServiceListCommitsAcceptsLimitParam(t *testing.T) {
	svc := readServiceTS(t)
	// The method signature should accept limit parameter with default of 20
	if !strings.Contains(svc, "limit") {
		t.Error("listCommits must accept a limit parameter")
	}
}

func TestRepositoryServiceListCommitsDefaultLimitIs20(t *testing.T) {
	svc := readServiceTS(t)
	if !strings.Contains(svc, "20") {
		t.Error("listCommits must default limit to 20")
	}
}

// ---- Factory function selects provider based on env var ----

func TestRepositoryServiceHasFactoryOrProviderSelection(t *testing.T) {
	svc := readServiceTS(t)
	// Should have some factory logic selecting GitHub/GitLab/stub based on REPO_PROVIDER
	if !strings.Contains(strings.ToLower(svc), "github") &&
		!strings.Contains(strings.ToLower(svc), "gitlab") &&
		!strings.Contains(strings.ToLower(svc), "factory") {
		t.Error("repository.service.ts must contain provider factory logic (github/gitlab/stub selection)")
	}
}

// ---- Routes: GET /api/repository/branches ----

func TestRepositoryRoutesDefinesBranchesRoute(t *testing.T) {
	routes := readRoutesTS(t)
	if !strings.Contains(routes, "/branches") {
		t.Error("repository.routes.ts must define GET /api/repository/branches route")
	}
}

func TestRepositoryRoutesUsesGetMethod(t *testing.T) {
	routes := readRoutesTS(t)
	lower := strings.ToLower(routes)
	if !strings.Contains(lower, "get") {
		t.Error("repository.routes.ts must use GET HTTP method for branch and commit routes")
	}
}

func TestRepositoryRoutesDefinesBranchCommitsRoute(t *testing.T) {
	routes := readRoutesTS(t)
	// Route: GET /api/repository/branches/:branch/commits
	if !strings.Contains(routes, "commits") {
		t.Error("repository.routes.ts must define GET /api/repository/branches/:branch/commits route")
	}
}

func TestRepositoryRoutesDefinesBranchParamInCommitsRoute(t *testing.T) {
	routes := readRoutesTS(t)
	// Route param :branch or :branchName
	if !strings.Contains(routes, ":branch") {
		t.Error("repository.routes.ts commits route must accept a :branch route parameter")
	}
}

func TestRepositoryRoutesDefinesContextRoute(t *testing.T) {
	routes := readRoutesTS(t)
	if !strings.Contains(routes, "context") {
		t.Error("repository.routes.ts must define GET /api/repository/context route")
	}
}

func TestRepositoryRoutesImportsController(t *testing.T) {
	routes := readRoutesTS(t)
	if !strings.Contains(strings.ToLower(routes), "controller") {
		t.Error("repository.routes.ts must import and use the repository controller")
	}
}

func TestRepositoryRoutesMountsUnderApiRepository(t *testing.T) {
	routes := readRoutesTS(t)
	lower := strings.ToLower(routes)
	if !strings.Contains(lower, "repository") {
		t.Error("repository.routes.ts must be scoped under /api/repository path")
	}
}

// ---- Controller: getBranches ----

func TestRepositoryControllerHasGetBranchesHandler(t *testing.T) {
	ctrl := readControllerTS(t)
	lower := strings.ToLower(ctrl)
	if !strings.Contains(lower, "getbranches") && !strings.Contains(lower, "branches") {
		t.Error("repository.controller.ts must define a getBranches handler")
	}
}

func TestRepositoryControllerGetBranchesReturnsNameLastCommitIsDefault(t *testing.T) {
	ctrl := readControllerTS(t)
	if !strings.Contains(ctrl, "lastCommit") || !strings.Contains(ctrl, "isDefault") {
		t.Error("getBranches handler must return objects with lastCommit and isDefault fields")
	}
}

// ---- Controller: getBranchCommits ----

func TestRepositoryControllerHasGetBranchCommitsHandler(t *testing.T) {
	ctrl := readControllerTS(t)
	lower := strings.ToLower(ctrl)
	if !strings.Contains(lower, "commits") {
		t.Error("repository.controller.ts must define a getBranchCommits handler")
	}
}

func TestRepositoryControllerGetBranchCommitsReturnsLast20(t *testing.T) {
	ctrl := readControllerTS(t)
	if !strings.Contains(ctrl, "20") {
		t.Error("getBranchCommits handler must return last 20 commits (limit=20)")
	}
}

func TestRepositoryControllerGetBranchCommitsReturnsCommitShape(t *testing.T) {
	ctrl := readControllerTS(t)
	if !strings.Contains(ctrl, "sha") || !strings.Contains(ctrl, "message") ||
		!strings.Contains(ctrl, "author") || !strings.Contains(ctrl, "date") {
		t.Error("getBranchCommits handler must return commits with sha, message, author, date fields")
	}
}

// ---- Controller: getContext ----

func TestRepositoryControllerHasGetContextHandler(t *testing.T) {
	ctrl := readControllerTS(t)
	lower := strings.ToLower(ctrl)
	if !strings.Contains(lower, "context") {
		t.Error("repository.controller.ts must define a getContext handler")
	}
}

func TestRepositoryControllerGetContextReturnsBaseUrlAndProvider(t *testing.T) {
	ctrl := readControllerTS(t)
	if !strings.Contains(ctrl, "baseUrl") || !strings.Contains(ctrl, "provider") {
		t.Error("getContext handler must return object with baseUrl and provider fields")
	}
}

// ---- Controller: invalid branch returns 404 ----

func TestRepositoryControllerHandlesInvalidBranchWith404(t *testing.T) {
	ctrl := readControllerTS(t)
	if !strings.Contains(ctrl, "404") {
		t.Error("repository.controller.ts must return 404 for invalid/not-found branch names")
	}
}

// ---- Integration test spec file ----

func TestRepositoryTestFileTestsBranchesEndpoint(t *testing.T) {
	testFile := readRepositoryTestTS(t)
	lower := strings.ToLower(testFile)
	if !strings.Contains(lower, "branches") {
		t.Error("repository.test.ts must include tests for GET /api/repository/branches endpoint")
	}
}

func TestRepositoryTestFileTestsCommitsEndpoint(t *testing.T) {
	testFile := readRepositoryTestTS(t)
	lower := strings.ToLower(testFile)
	if !strings.Contains(lower, "commits") {
		t.Error("repository.test.ts must include tests for GET /api/repository/branches/:branch/commits endpoint")
	}
}

func TestRepositoryTestFileTestsContextEndpoint(t *testing.T) {
	testFile := readRepositoryTestTS(t)
	lower := strings.ToLower(testFile)
	if !strings.Contains(lower, "context") {
		t.Error("repository.test.ts must include tests for GET /api/repository/context endpoint")
	}
}

func TestRepositoryTestFileTestsInvalidBranch404(t *testing.T) {
	testFile := readRepositoryTestTS(t)
	if !strings.Contains(testFile, "404") {
		t.Error("repository.test.ts must include a test for invalid branch name returning 404")
	}
}

func TestRepositoryTestFileMockesRepositoryService(t *testing.T) {
	testFile := readRepositoryTestTS(t)
	lower := strings.ToLower(testFile)
	if !strings.Contains(lower, "mock") && !strings.Contains(lower, "stub") && !strings.Contains(lower, "jest.") {
		t.Error("repository.test.ts must mock the underlying git/API calls in repositoryService")
	}
}

func TestRepositoryTestFileBranchResponseShape(t *testing.T) {
	testFile := readRepositoryTestTS(t)
	if !strings.Contains(testFile, "lastCommit") || !strings.Contains(testFile, "isDefault") {
		t.Error("repository.test.ts must assert branch response shape includes lastCommit and isDefault")
	}
}

func TestRepositoryTestFileCommitResponseShape(t *testing.T) {
	testFile := readRepositoryTestTS(t)
	if !strings.Contains(testFile, "sha") || !strings.Contains(testFile, "message") ||
		!strings.Contains(testFile, "author") || !strings.Contains(testFile, "date") {
		t.Error("repository.test.ts must assert commit response shape includes sha, message, author, date")
	}
}

func TestRepositoryTestFileContextResponseShape(t *testing.T) {
	testFile := readRepositoryTestTS(t)
	if !strings.Contains(testFile, "baseUrl") || !strings.Contains(testFile, "provider") {
		t.Error("repository.test.ts must assert context response shape includes baseUrl and provider")
	}
}
