/**
 * Integration tests for admin cleanup API endpoints (TASK-4619).
 *
 * These tests verify:
 * - POST /api/admin/cleanup/age   → 200 with CleanupResult
 * - POST /api/admin/cleanup/count → 200 with CleanupResult
 * - GET  /api/admin/cleanup/jobs  → 200 with list of CleanupJob records
 * - Non-admin request             → 403 Forbidden
 *
 * Auth strategy: X-Admin-Token header checked against ADMIN_TOKEN env var.
 * The adminAuthMiddleware is mocked in unit tests; here we test the full
 * integration stack with supertest or a similar HTTP test client.
 */

import request from 'supertest';
import { app } from '../../src/app';
import * as cleanupService from '../../src/services/cleanup.service';
import * as cleanupJobModel from '../../src/models/cleanup-job.model';
import { CleanupJobType, CleanupJobStatus } from '../../src/types/cleanup.types';

// Mock the cleanup service so tests don't hit the database
jest.mock('../../src/services/cleanup.service');
jest.mock('../../src/models/cleanup-job.model');

const ADMIN_TOKEN = 'test-admin-token-secret';

beforeAll(() => {
  process.env.ADMIN_TOKEN = ADMIN_TOKEN;
});

afterAll(() => {
  delete process.env.ADMIN_TOKEN;
});

// ---- POST /api/admin/cleanup/age ----

describe('POST /api/admin/cleanup/age', () => {
  it('returns 200 with CleanupResult when request includes valid admin token', async () => {
    const mockResult = {
      jobId: 'abc-123',
      type: CleanupJobType.AGE_BASED,
      deletedCount: 42,
      duration: 300,
    };
    (cleanupService.runAgeBased as jest.Mock).mockResolvedValue(mockResult);

    const response = await request(app)
      .post('/api/admin/cleanup/age')
      .set('X-Admin-Token', ADMIN_TOKEN);

    expect(response.status).toBe(200);
    expect(response.body).toMatchObject({
      jobId: expect.any(String),
      type: CleanupJobType.AGE_BASED,
      deletedCount: expect.any(Number),
      duration: expect.any(Number),
    });
  });

  it('calls runAgeBased on the cleanup service', async () => {
    const mockResult = {
      jobId: 'abc-123',
      type: CleanupJobType.AGE_BASED,
      deletedCount: 0,
      duration: 100,
    };
    (cleanupService.runAgeBased as jest.Mock).mockResolvedValue(mockResult);

    await request(app)
      .post('/api/admin/cleanup/age')
      .set('X-Admin-Token', ADMIN_TOKEN);

    expect(cleanupService.runAgeBased).toHaveBeenCalledTimes(1);
  });

  it('returns 403 when X-Admin-Token header is missing', async () => {
    const response = await request(app).post('/api/admin/cleanup/age');

    expect(response.status).toBe(403);
  });

  it('returns 403 when X-Admin-Token header is wrong', async () => {
    const response = await request(app)
      .post('/api/admin/cleanup/age')
      .set('X-Admin-Token', 'wrong-token');

    expect(response.status).toBe(403);
  });

  it('returns CleanupResult with jobId field', async () => {
    const mockResult = {
      jobId: 'job-uuid-1234',
      type: CleanupJobType.AGE_BASED,
      deletedCount: 10,
      duration: 50,
    };
    (cleanupService.runAgeBased as jest.Mock).mockResolvedValue(mockResult);

    const response = await request(app)
      .post('/api/admin/cleanup/age')
      .set('X-Admin-Token', ADMIN_TOKEN);

    expect(response.body.jobId).toBeDefined();
  });

  it('returns CleanupResult with deletedCount field', async () => {
    const mockResult = {
      jobId: 'job-uuid-5678',
      type: CleanupJobType.AGE_BASED,
      deletedCount: 99,
      duration: 200,
    };
    (cleanupService.runAgeBased as jest.Mock).mockResolvedValue(mockResult);

    const response = await request(app)
      .post('/api/admin/cleanup/age')
      .set('X-Admin-Token', ADMIN_TOKEN);

    expect(response.body.deletedCount).toBe(99);
  });

  it('returns CleanupResult with duration field', async () => {
    const mockResult = {
      jobId: 'job-uuid-9999',
      type: CleanupJobType.AGE_BASED,
      deletedCount: 5,
      duration: 750,
    };
    (cleanupService.runAgeBased as jest.Mock).mockResolvedValue(mockResult);

    const response = await request(app)
      .post('/api/admin/cleanup/age')
      .set('X-Admin-Token', ADMIN_TOKEN);

    expect(response.body.duration).toBe(750);
  });
});

// ---- POST /api/admin/cleanup/count ----

describe('POST /api/admin/cleanup/count', () => {
  it('returns 200 with CleanupResult when request includes valid admin token', async () => {
    const mockResult = {
      jobId: 'def-456',
      type: CleanupJobType.COUNT_BASED,
      deletedCount: 17,
      duration: 180,
    };
    (cleanupService.runCountBased as jest.Mock).mockResolvedValue(mockResult);

    const response = await request(app)
      .post('/api/admin/cleanup/count')
      .set('X-Admin-Token', ADMIN_TOKEN);

    expect(response.status).toBe(200);
    expect(response.body).toMatchObject({
      jobId: expect.any(String),
      type: CleanupJobType.COUNT_BASED,
      deletedCount: expect.any(Number),
      duration: expect.any(Number),
    });
  });

  it('calls runCountBased on the cleanup service', async () => {
    const mockResult = {
      jobId: 'def-456',
      type: CleanupJobType.COUNT_BASED,
      deletedCount: 0,
      duration: 100,
    };
    (cleanupService.runCountBased as jest.Mock).mockResolvedValue(mockResult);

    await request(app)
      .post('/api/admin/cleanup/count')
      .set('X-Admin-Token', ADMIN_TOKEN);

    expect(cleanupService.runCountBased).toHaveBeenCalledTimes(1);
  });

  it('returns 403 when X-Admin-Token header is missing', async () => {
    const response = await request(app).post('/api/admin/cleanup/count');

    expect(response.status).toBe(403);
  });

  it('returns 403 when X-Admin-Token header is wrong', async () => {
    const response = await request(app)
      .post('/api/admin/cleanup/count')
      .set('X-Admin-Token', 'invalid-token');

    expect(response.status).toBe(403);
  });

  it('returns CleanupResult with type COUNT_BASED', async () => {
    const mockResult = {
      jobId: 'ghi-789',
      type: CleanupJobType.COUNT_BASED,
      deletedCount: 25,
      duration: 400,
    };
    (cleanupService.runCountBased as jest.Mock).mockResolvedValue(mockResult);

    const response = await request(app)
      .post('/api/admin/cleanup/count')
      .set('X-Admin-Token', ADMIN_TOKEN);

    expect(response.body.type).toBe(CleanupJobType.COUNT_BASED);
  });
});

// ---- GET /api/admin/cleanup/jobs ----

describe('GET /api/admin/cleanup/jobs', () => {
  it('returns 200 with list of recent CleanupJob records when request includes valid admin token', async () => {
    const mockJobs = [
      {
        id: 'job-1',
        type: CleanupJobType.AGE_BASED,
        status: CleanupJobStatus.DONE,
        deletedCount: 10,
        startedAt: new Date('2026-03-10T10:00:00Z'),
        completedAt: new Date('2026-03-10T10:01:00Z'),
        errorMessage: null,
      },
      {
        id: 'job-2',
        type: CleanupJobType.COUNT_BASED,
        status: CleanupJobStatus.FAILED,
        deletedCount: 0,
        startedAt: new Date('2026-03-11T10:00:00Z'),
        completedAt: new Date('2026-03-11T10:00:05Z'),
        errorMessage: 'Database error',
      },
    ];
    (cleanupJobModel.list as jest.Mock).mockResolvedValue(mockJobs);

    const response = await request(app)
      .get('/api/admin/cleanup/jobs')
      .set('X-Admin-Token', ADMIN_TOKEN);

    expect(response.status).toBe(200);
    expect(Array.isArray(response.body)).toBe(true);
  });

  it('returns an array of CleanupJob records', async () => {
    const mockJobs = [
      {
        id: 'job-3',
        type: CleanupJobType.AGE_BASED,
        status: CleanupJobStatus.DONE,
        deletedCount: 50,
        startedAt: new Date('2026-03-15T08:00:00Z'),
        completedAt: new Date('2026-03-15T08:00:30Z'),
        errorMessage: null,
      },
    ];
    (cleanupJobModel.list as jest.Mock).mockResolvedValue(mockJobs);

    const response = await request(app)
      .get('/api/admin/cleanup/jobs')
      .set('X-Admin-Token', ADMIN_TOKEN);

    expect(response.body).toHaveLength(1);
    expect(response.body[0].id).toBe('job-3');
  });

  it('calls cleanupJobModel.list to retrieve recent jobs', async () => {
    (cleanupJobModel.list as jest.Mock).mockResolvedValue([]);

    await request(app)
      .get('/api/admin/cleanup/jobs')
      .set('X-Admin-Token', ADMIN_TOKEN);

    expect(cleanupJobModel.list).toHaveBeenCalled();
  });

  it('returns 403 when X-Admin-Token header is missing', async () => {
    const response = await request(app).get('/api/admin/cleanup/jobs');

    expect(response.status).toBe(403);
  });

  it('returns 403 when X-Admin-Token header is wrong', async () => {
    const response = await request(app)
      .get('/api/admin/cleanup/jobs')
      .set('X-Admin-Token', 'not-the-right-token');

    expect(response.status).toBe(403);
  });

  it('returns an empty array when no cleanup jobs exist', async () => {
    (cleanupJobModel.list as jest.Mock).mockResolvedValue([]);

    const response = await request(app)
      .get('/api/admin/cleanup/jobs')
      .set('X-Admin-Token', ADMIN_TOKEN);

    expect(response.status).toBe(200);
    expect(response.body).toEqual([]);
  });

  it('each job record includes id, type, status, deletedCount, startedAt', async () => {
    const mockJobs = [
      {
        id: 'job-4',
        type: CleanupJobType.AGE_BASED,
        status: CleanupJobStatus.DONE,
        deletedCount: 7,
        startedAt: new Date('2026-03-16T09:00:00Z'),
        completedAt: new Date('2026-03-16T09:00:10Z'),
        errorMessage: null,
      },
    ];
    (cleanupJobModel.list as jest.Mock).mockResolvedValue(mockJobs);

    const response = await request(app)
      .get('/api/admin/cleanup/jobs')
      .set('X-Admin-Token', ADMIN_TOKEN);

    const job = response.body[0];
    expect(job.id).toBeDefined();
    expect(job.type).toBeDefined();
    expect(job.status).toBeDefined();
    expect(job.deletedCount).toBeDefined();
    expect(job.startedAt).toBeDefined();
  });
});

// ---- Admin auth middleware ----

describe('Admin auth middleware', () => {
  it('rejects requests to all admin endpoints when no token is provided', async () => {
    const [ageRes, countRes, jobsRes] = await Promise.all([
      request(app).post('/api/admin/cleanup/age'),
      request(app).post('/api/admin/cleanup/count'),
      request(app).get('/api/admin/cleanup/jobs'),
    ]);

    expect(ageRes.status).toBe(403);
    expect(countRes.status).toBe(403);
    expect(jobsRes.status).toBe(403);
  });

  it('rejects requests with incorrect admin token', async () => {
    const badToken = 'bad-token-value';
    const [ageRes, countRes, jobsRes] = await Promise.all([
      request(app).post('/api/admin/cleanup/age').set('X-Admin-Token', badToken),
      request(app).post('/api/admin/cleanup/count').set('X-Admin-Token', badToken),
      request(app).get('/api/admin/cleanup/jobs').set('X-Admin-Token', badToken),
    ]);

    expect(ageRes.status).toBe(403);
    expect(countRes.status).toBe(403);
    expect(jobsRes.status).toBe(403);
  });

  it('allows requests with correct ADMIN_TOKEN env var value', async () => {
    const mockResult = {
      jobId: 'auth-test-job',
      type: CleanupJobType.AGE_BASED,
      deletedCount: 0,
      duration: 10,
    };
    (cleanupService.runAgeBased as jest.Mock).mockResolvedValue(mockResult);

    const response = await request(app)
      .post('/api/admin/cleanup/age')
      .set('X-Admin-Token', ADMIN_TOKEN);

    expect(response.status).toBe(200);
  });

  it('rejects when ADMIN_TOKEN env var is not set', async () => {
    const savedToken = process.env.ADMIN_TOKEN;
    delete process.env.ADMIN_TOKEN;

    const response = await request(app)
      .post('/api/admin/cleanup/age')
      .set('X-Admin-Token', ADMIN_TOKEN);

    process.env.ADMIN_TOKEN = savedToken;

    // Without ADMIN_TOKEN env var configured, all requests should be rejected
    expect(response.status).toBe(403);
  });
});
