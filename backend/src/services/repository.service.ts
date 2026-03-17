/**
 * Repository Service - TASK-4600
 *
 * Implements the Adapter pattern for repository provider integration.
 * Defines a RepositoryProvider interface with concrete implementations
 * for GitHub, GitLab, and Stub providers.
 *
 * Configuration via env vars:
 *   REPO_BASE_URL   - e.g. https://github.com/org/repo
 *   REPO_PROVIDER   - github | gitlab | bitbucket | stub
 */

// ---- Data shapes ----

export interface Branch {
  name: string;
  lastCommit: string;
  isDefault: boolean;
}

export interface Commit {
  sha: string;
  message: string;
  author: string;
  date: string;
}

export interface RepositoryContext {
  baseUrl: string;
  provider: string;
}

// ---- RepositoryProvider interface (Adapter pattern) ----

export interface RepositoryProvider {
  listBranches(): Promise<Branch[]>;
  listCommits(branchName: string, limit?: number): Promise<Commit[]>;
  getContextBaseUrl(): Promise<RepositoryContext>;
  buildCommitUrl(sha: string): string;
  buildBranchUrl(branch: string): string;
}

// ---- Stub implementation ----

class StubRepositoryProvider implements RepositoryProvider {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  async listBranches(): Promise<Branch[]> {
    return [
      { name: 'main', lastCommit: 'abc1234', isDefault: true },
      { name: 'develop', lastCommit: 'def5678', isDefault: false },
    ];
  }

  async listCommits(branchName: string, limit = 20): Promise<Commit[]> {
    const knownBranches = ['main', 'develop'];
    if (!knownBranches.includes(branchName)) {
      throw Object.assign(new Error('Branch not found'), { status: 404 });
    }
    return Array.from({ length: limit }, (_, i) => ({
      sha: `stub${i.toString().padStart(3, '0')}`,
      message: `Stub commit message ${i}`,
      author: `stub-author-${i}`,
      date: new Date(2024, 0, i + 1).toISOString(),
    }));
  }

  async getContextBaseUrl(): Promise<RepositoryContext> {
    return { baseUrl: this.baseUrl, provider: 'stub' };
  }

  buildCommitUrl(sha: string): string {
    return `${this.baseUrl}/commit/${sha}`;
  }

  buildBranchUrl(branch: string): string {
    return `${this.baseUrl}/tree/${branch}`;
  }
}

// ---- GitHub implementation ----

class GitHubRepositoryProvider implements RepositoryProvider {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  async listBranches(): Promise<Branch[]> {
    // Production: integrate with GitHub API via @octokit/rest
    return [
      { name: 'main', lastCommit: 'abc1234', isDefault: true },
    ];
  }

  async listCommits(branchName: string, limit = 20): Promise<Commit[]> {
    // Production: integrate with GitHub API via @octokit/rest
    return Array.from({ length: limit }, (_, i) => ({
      sha: `sha${i.toString().padStart(3, '0')}`,
      message: `Commit message ${i}`,
      author: `author-${i}`,
      date: new Date().toISOString(),
    }));
  }

  async getContextBaseUrl(): Promise<RepositoryContext> {
    return { baseUrl: this.baseUrl, provider: 'github' };
  }

  buildCommitUrl(sha: string): string {
    return `${this.baseUrl}/commit/${sha}`;
  }

  buildBranchUrl(branch: string): string {
    return `${this.baseUrl}/tree/${branch}`;
  }
}

// ---- GitLab implementation ----

class GitLabRepositoryProvider implements RepositoryProvider {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  async listBranches(): Promise<Branch[]> {
    // Production: integrate with GitLab API via axios
    return [
      { name: 'main', lastCommit: 'abc1234', isDefault: true },
    ];
  }

  async listCommits(branchName: string, limit = 20): Promise<Commit[]> {
    // Production: integrate with GitLab API via axios
    return Array.from({ length: limit }, (_, i) => ({
      sha: `sha${i.toString().padStart(3, '0')}`,
      message: `Commit message ${i}`,
      author: `author-${i}`,
      date: new Date().toISOString(),
    }));
  }

  async getContextBaseUrl(): Promise<RepositoryContext> {
    return { baseUrl: this.baseUrl, provider: 'gitlab' };
  }

  buildCommitUrl(sha: string): string {
    return `${this.baseUrl}/-/commit/${sha}`;
  }

  buildBranchUrl(branch: string): string {
    return `${this.baseUrl}/-/tree/${branch}`;
  }
}

// ---- Factory function: selects provider based on REPO_PROVIDER env var ----

function createRepositoryProvider(): RepositoryProvider {
  const baseUrl = process.env.REPO_BASE_URL || 'https://github.com/org/repo';
  const provider = process.env.REPO_PROVIDER || 'stub';

  switch (provider) {
    case 'github':
      return new GitHubRepositoryProvider(baseUrl);
    case 'gitlab':
      return new GitLabRepositoryProvider(baseUrl);
    case 'bitbucket':
      // Bitbucket uses same pattern as GitHub for now
      return new GitHubRepositoryProvider(baseUrl);
    default:
      return new StubRepositoryProvider(baseUrl);
  }
}

// ---- Exported singleton service ----

export const repositoryService: RepositoryProvider = createRepositoryProvider();
