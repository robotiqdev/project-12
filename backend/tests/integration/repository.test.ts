/**
 * Integration tests for Repository API Endpoints (TASK-4600).
 *
 * Tests:
 *   GET /api/repository/branches        -> array of { name, lastCommit, isDefault }
 *   GET /api/repository/branches/:branch/commits -> last 20 commits { sha, message, author, date }
 *   GET /api/repository/context         -> { baseUrl, provider }
 *   GET /api/repository/branches/:branch/commits (invalid branch) -> 404
 *
 * Mocks the underlying git/API calls in repositoryService so no real network
 * or filesystem access is required during tests.
 */

import request from 'supertest';
import { jest } from '@jest/globals';

// Mock the repository service before importing the app so that the factory
// returns the stub provider rather than trying real network calls.
jest.mock('../../src/services/repository.service', () => {
  const mockBranches = [
    { name: 'main', lastCommit: 'abc1234', isDefault: true },
    { name: 'feature/FEAT-1159', lastCommit: 'def5678', isDefault: false },
  ];

  const mockCommits = Array.from({ length: 20 }, (_, i) => ({
    sha: `sha${i.toString().padStart(3, '0')}`,
    message: `commit message ${i}`,
    author: `author-${i}`,
    date: new Date(2024, 0, i + 1).toISOString(),
  }));

  return {
    repositoryService: {
      listBranches: jest.fn().mockResolvedValue(mockBranches),
      listCommits: jest.fn(async (branchName: string, limit = 20) => {
        if (!mockBranches.some((b) => b.name === branchName)) {
          throw Object.assign(new Error('Branch not found'), { status: 404 });
        }
        return mockCommits.slice(0, limit);
      }),
      getContextBaseUrl: jest.fn().mockResolvedValue({
        baseUrl: 'https://github.com/org/repo',
        provider: 'github',
      }),
    },
  };
});

import app from '../../src/app';

// ---------------------------------------------------------------------------
// GET /api/repository/branches
// ---------------------------------------------------------------------------

describe('GET /api/repository/branches', () => {
  it('returns 200 with an array of branch objects', async () => {
    const res = await request(app).get('/api/repository/branches');
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body)).toBe(true);
  });

  it('each branch object has name, lastCommit, and isDefault fields', async () => {
    const res = await request(app).get('/api/repository/branches');
    expect(res.status).toBe(200);
    expect(res.body.length).toBeGreaterThan(0);
    for (const branch of res.body) {
      expect(branch).toHaveProperty('name');
      expect(branch).toHaveProperty('lastCommit');
      expect(branch).toHaveProperty('isDefault');
    }
  });

  it('name is a string', async () => {
    const res = await request(app).get('/api/repository/branches');
    expect(res.status).toBe(200);
    for (const branch of res.body) {
      expect(typeof branch.name).toBe('string');
    }
  });

  it('isDefault is a boolean', async () => {
    const res = await request(app).get('/api/repository/branches');
    expect(res.status).toBe(200);
    for (const branch of res.body) {
      expect(typeof branch.isDefault).toBe('boolean');
    }
  });

  it('lastCommit is a string', async () => {
    const res = await request(app).get('/api/repository/branches');
    expect(res.status).toBe(200);
    for (const branch of res.body) {
      expect(typeof branch.lastCommit).toBe('string');
    }
  });

  it('exactly one branch has isDefault=true', async () => {
    const res = await request(app).get('/api/repository/branches');
    expect(res.status).toBe(200);
    const defaultBranches = res.body.filter((b: { isDefault: boolean }) => b.isDefault === true);
    expect(defaultBranches.length).toBe(1);
  });
});

// ---------------------------------------------------------------------------
// GET /api/repository/branches/:branch/commits
// ---------------------------------------------------------------------------

describe('GET /api/repository/branches/:branch/commits', () => {
  it('returns 200 with an array of commit objects for a valid branch', async () => {
    const res = await request(app).get('/api/repository/branches/main/commits');
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body)).toBe(true);
  });

  it('returns exactly 20 commits by default', async () => {
    const res = await request(app).get('/api/repository/branches/main/commits');
    expect(res.status).toBe(200);
    expect(res.body.length).toBe(20);
  });

  it('each commit object has sha, message, author, and date fields', async () => {
    const res = await request(app).get('/api/repository/branches/main/commits');
    expect(res.status).toBe(200);
    expect(res.body.length).toBeGreaterThan(0);
    for (const commit of res.body) {
      expect(commit).toHaveProperty('sha');
      expect(commit).toHaveProperty('message');
      expect(commit).toHaveProperty('author');
      expect(commit).toHaveProperty('date');
    }
  });

  it('sha is a non-empty string', async () => {
    const res = await request(app).get('/api/repository/branches/main/commits');
    expect(res.status).toBe(200);
    for (const commit of res.body) {
      expect(typeof commit.sha).toBe('string');
      expect(commit.sha.length).toBeGreaterThan(0);
    }
  });

  it('message is a string', async () => {
    const res = await request(app).get('/api/repository/branches/main/commits');
    expect(res.status).toBe(200);
    for (const commit of res.body) {
      expect(typeof commit.message).toBe('string');
    }
  });

  it('author is a string', async () => {
    const res = await request(app).get('/api/repository/branches/main/commits');
    expect(res.status).toBe(200);
    for (const commit of res.body) {
      expect(typeof commit.author).toBe('string');
    }
  });

  it('date is a string (ISO format)', async () => {
    const res = await request(app).get('/api/repository/branches/main/commits');
    expect(res.status).toBe(200);
    for (const commit of res.body) {
      expect(typeof commit.date).toBe('string');
      expect(() => new Date(commit.date)).not.toThrow();
    }
  });

  it('returns 404 for an invalid/nonexistent branch name', async () => {
    const res = await request(app).get('/api/repository/branches/nonexistent-branch-xyz/commits');
    expect(res.status).toBe(404);
  });

  it('returns 404 for a branch name with special characters that does not exist', async () => {
    const res = await request(app).get('/api/repository/branches/../../etc-passwd/commits');
    expect(res.status).toBe(404);
  });

  it('handles URL-encoded branch names with slashes', async () => {
    const res = await request(app).get(
      '/api/repository/branches/feature%2FFEAT-1159/commits'
    );
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body)).toBe(true);
  });
});

// ---------------------------------------------------------------------------
// GET /api/repository/context
// ---------------------------------------------------------------------------

describe('GET /api/repository/context', () => {
  it('returns 200 with a context object', async () => {
    const res = await request(app).get('/api/repository/context');
    expect(res.status).toBe(200);
    expect(typeof res.body).toBe('object');
    expect(res.body).not.toBeNull();
    expect(Array.isArray(res.body)).toBe(false);
  });

  it('context object has baseUrl field', async () => {
    const res = await request(app).get('/api/repository/context');
    expect(res.status).toBe(200);
    expect(res.body).toHaveProperty('baseUrl');
  });

  it('context object has provider field', async () => {
    const res = await request(app).get('/api/repository/context');
    expect(res.status).toBe(200);
    expect(res.body).toHaveProperty('provider');
  });

  it('baseUrl is a string', async () => {
    const res = await request(app).get('/api/repository/context');
    expect(res.status).toBe(200);
    expect(typeof res.body.baseUrl).toBe('string');
  });

  it('provider is a string (github | gitlab | bitbucket)', async () => {
    const res = await request(app).get('/api/repository/context');
    expect(res.status).toBe(200);
    expect(typeof res.body.provider).toBe('string');
    expect(['github', 'gitlab', 'bitbucket', 'stub']).toContain(res.body.provider);
  });

  it('baseUrl is a valid URL', async () => {
    const res = await request(app).get('/api/repository/context');
    expect(res.status).toBe(200);
    expect(() => new URL(res.body.baseUrl)).not.toThrow();
  });
});
