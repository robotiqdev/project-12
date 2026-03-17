/**
 * Integration tests for POST /api/validation-runs — TASK-4593
 *
 * Tests the public HTTP interface for creating validation runs:
 *   - Returns 201 with run record (id, status=PENDING) on success
 *   - Returns 409 BRANCH_LOCKED when same branch already has PENDING/RUNNING run
 *   - Returns 400 on invalid commitSha (not 40-char hex)
 *   - Returns 400 on missing branchName or commitSha
 *   - Allows new run when previous run on branch is FAILED
 *   - Calls pipelineService.enqueue asynchronously (fire-and-forget)
 */

import request from 'supertest';
import { app } from '../../src/app';
import { prisma } from '../../src/db/client';

// Mock the pipeline service to intercept fire-and-forget calls
jest.mock('../../src/services/pipeline.service');
import * as pipelineService from '../../src/services/pipeline.service';
const mockEnqueue = pipelineService.enqueue as jest.Mock;

/** A valid 40-character hex commit SHA for use across tests */
const VALID_COMMIT_SHA = 'a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2';

describe('POST /api/validation-runs', () => {
  beforeEach(async () => {
    // Truncate all tables before each test to ensure isolation
    await prisma.$executeRaw`TRUNCATE TABLE run_stages, validation_runs CASCADE`;
    jest.clearAllMocks();
    // Default: pipeline enqueue resolves successfully
    mockEnqueue.mockResolvedValue(undefined);
  });

  afterAll(async () => {
    await prisma.$disconnect();
  });

  // ── Happy path ────────────────────────────────────────────────────────────

  it('returns 201 with a run record containing id and status=PENDING for a valid request', async () => {
    const response = await request(app)
      .post('/api/validation-runs')
      .send({
        branchName: 'feature/my-branch',
        commitSha: VALID_COMMIT_SHA,
      })
      .expect(201);

    expect(response.body).toMatchObject({
      id: expect.any(String),
      status: 'PENDING',
      branchName: 'feature/my-branch',
      commitSha: VALID_COMMIT_SHA,
    });
    expect(response.body.id).toBeTruthy();
  });

  it('persists the new run with status PENDING in the database', async () => {
    const response = await request(app)
      .post('/api/validation-runs')
      .send({
        branchName: 'feature/store-check',
        commitSha: VALID_COMMIT_SHA,
      })
      .expect(201);

    const runId = response.body.id;
    const savedRun = await prisma.validationRun.findUnique({ where: { id: runId } });

    expect(savedRun).not.toBeNull();
    expect(savedRun!.status).toBe('PENDING');
  });

  it('accepts optional config with envVars and featureFlags and returns 201', async () => {
    const response = await request(app)
      .post('/api/validation-runs')
      .send({
        branchName: 'feature/with-config',
        commitSha: VALID_COMMIT_SHA,
        config: {
          envVars: { NODE_ENV: 'test', API_KEY: 'secret' },
          featureFlags: { darkMode: true, betaFeature: false },
        },
      })
      .expect(201);

    expect(response.body.id).toBeTruthy();
    expect(response.body.status).toBe('PENDING');
  });

  it('returns 201 without config when config field is omitted', async () => {
    const response = await request(app)
      .post('/api/validation-runs')
      .send({
        branchName: 'feature/no-config',
        commitSha: VALID_COMMIT_SHA,
      })
      .expect(201);

    expect(response.body.status).toBe('PENDING');
    expect(response.body.id).toBeTruthy();
  });

  // ── Concurrency / branch lock ────────────────────────────────────────────

  it('returns 409 with BRANCH_LOCKED error when a PENDING run already exists on the same branch', async () => {
    const branchName = 'feature/locked-branch';

    // Create the first run (will be PENDING)
    const firstResponse = await request(app)
      .post('/api/validation-runs')
      .send({ branchName, commitSha: VALID_COMMIT_SHA })
      .expect(201);

    const firstRunId = firstResponse.body.id;

    // Attempt a second run on the same branch while first is PENDING
    const secondResponse = await request(app)
      .post('/api/validation-runs')
      .send({ branchName, commitSha: VALID_COMMIT_SHA })
      .expect(409);

    expect(secondResponse.body).toMatchObject({
      error: 'BRANCH_LOCKED',
      conflictingRunId: firstRunId,
    });
  });

  it('returns 409 conflictingRunId that matches the blocking run id', async () => {
    const branchName = 'feature/conflict-id-check';

    const firstResponse = await request(app)
      .post('/api/validation-runs')
      .send({ branchName, commitSha: VALID_COMMIT_SHA })
      .expect(201);

    const secondResponse = await request(app)
      .post('/api/validation-runs')
      .send({ branchName, commitSha: VALID_COMMIT_SHA })
      .expect(409);

    expect(secondResponse.body.conflictingRunId).toBe(firstResponse.body.id);
    expect(secondResponse.body.error).toBe('BRANCH_LOCKED');
  });

  it('returns 409 when a RUNNING run exists on the branch', async () => {
    const branchName = 'feature/running-branch';

    const firstResponse = await request(app)
      .post('/api/validation-runs')
      .send({ branchName, commitSha: VALID_COMMIT_SHA })
      .expect(201);

    // Simulate the run being picked up and set to RUNNING
    await prisma.validationRun.update({
      where: { id: firstResponse.body.id },
      data: { status: 'RUNNING' },
    });

    const secondResponse = await request(app)
      .post('/api/validation-runs')
      .send({ branchName, commitSha: VALID_COMMIT_SHA })
      .expect(409);

    expect(secondResponse.body.error).toBe('BRANCH_LOCKED');
  });

  it('allows a new run on a branch that previously had a FAILED run', async () => {
    const branchName = 'feature/previously-failed';

    // Create and mark as failed
    const firstResponse = await request(app)
      .post('/api/validation-runs')
      .send({ branchName, commitSha: VALID_COMMIT_SHA })
      .expect(201);

    await prisma.validationRun.update({
      where: { id: firstResponse.body.id },
      data: { status: 'FAILED' },
    });

    // New run on the same branch should succeed
    const secondResponse = await request(app)
      .post('/api/validation-runs')
      .send({ branchName, commitSha: VALID_COMMIT_SHA })
      .expect(201);

    expect(secondResponse.body.status).toBe('PENDING');
    expect(secondResponse.body.id).not.toBe(firstResponse.body.id);
  });

  it('allows a new run on a branch that previously had a SUCCESS run', async () => {
    const branchName = 'feature/previously-succeeded';

    const firstResponse = await request(app)
      .post('/api/validation-runs')
      .send({ branchName, commitSha: VALID_COMMIT_SHA })
      .expect(201);

    await prisma.validationRun.update({
      where: { id: firstResponse.body.id },
      data: { status: 'SUCCESS' },
    });

    const secondResponse = await request(app)
      .post('/api/validation-runs')
      .send({ branchName, commitSha: VALID_COMMIT_SHA })
      .expect(201);

    expect(secondResponse.body.status).toBe('PENDING');
  });

  it('allows a new run on a branch that previously had a CANCELLED run', async () => {
    const branchName = 'feature/previously-cancelled';

    const firstResponse = await request(app)
      .post('/api/validation-runs')
      .send({ branchName, commitSha: VALID_COMMIT_SHA })
      .expect(201);

    await prisma.validationRun.update({
      where: { id: firstResponse.body.id },
      data: { status: 'CANCELLED' },
    });

    const secondResponse = await request(app)
      .post('/api/validation-runs')
      .send({ branchName, commitSha: VALID_COMMIT_SHA })
      .expect(201);

    expect(secondResponse.body.status).toBe('PENDING');
  });

  // ── Validation errors (400) ───────────────────────────────────────────────

  it('returns 400 when commitSha is not a 40-character hex string (too short)', async () => {
    const response = await request(app)
      .post('/api/validation-runs')
      .send({
        branchName: 'feature/bad-sha',
        commitSha: 'abc123',
      })
      .expect(400);

    expect(response.body).toHaveProperty('error');
  });

  it('returns 400 when commitSha contains non-hex characters', async () => {
    const response = await request(app)
      .post('/api/validation-runs')
      .send({
        branchName: 'feature/bad-sha',
        commitSha: 'z'.repeat(40), // 'z' is not a valid hex character
      })
      .expect(400);

    expect(response.body).toHaveProperty('error');
  });

  it('returns 400 when commitSha is 40 chars but uppercase (must be lowercase hex)', async () => {
    const response = await request(app)
      .post('/api/validation-runs')
      .send({
        branchName: 'feature/uppercase-sha',
        commitSha: 'A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2', // uppercase
      })
      .expect(400);

    expect(response.body).toHaveProperty('error');
  });

  it('returns 400 when commitSha is exactly 39 characters (one too short)', async () => {
    const response = await request(app)
      .post('/api/validation-runs')
      .send({
        branchName: 'feature/short-sha',
        commitSha: 'a'.repeat(39),
      })
      .expect(400);

    expect(response.body).toHaveProperty('error');
  });

  it('returns 400 when commitSha is exactly 41 characters (one too long)', async () => {
    const response = await request(app)
      .post('/api/validation-runs')
      .send({
        branchName: 'feature/long-sha',
        commitSha: 'a'.repeat(41),
      })
      .expect(400);

    expect(response.body).toHaveProperty('error');
  });

  it('returns 400 when branchName is missing from the request body', async () => {
    const response = await request(app)
      .post('/api/validation-runs')
      .send({
        commitSha: VALID_COMMIT_SHA,
      })
      .expect(400);

    expect(response.body).toHaveProperty('error');
  });

  it('returns 400 when branchName is an empty string', async () => {
    const response = await request(app)
      .post('/api/validation-runs')
      .send({
        branchName: '',
        commitSha: VALID_COMMIT_SHA,
      })
      .expect(400);

    expect(response.body).toHaveProperty('error');
  });

  it('returns 400 when commitSha is missing from the request body', async () => {
    const response = await request(app)
      .post('/api/validation-runs')
      .send({
        branchName: 'feature/no-sha',
      })
      .expect(400);

    expect(response.body).toHaveProperty('error');
  });

  it('returns 400 when the entire request body is empty', async () => {
    const response = await request(app)
      .post('/api/validation-runs')
      .send({})
      .expect(400);

    expect(response.body).toHaveProperty('error');
  });

  // ── Pipeline service async integration ───────────────────────────────────

  it('calls pipelineService.enqueue with the new runId after creating the run', async () => {
    const response = await request(app)
      .post('/api/validation-runs')
      .send({
        branchName: 'feature/pipeline-check',
        commitSha: VALID_COMMIT_SHA,
      })
      .expect(201);

    // Allow the event loop a tick for fire-and-forget to be scheduled
    await new Promise((resolve) => process.nextTick(resolve));

    expect(mockEnqueue).toHaveBeenCalledWith(response.body.id);
    expect(mockEnqueue).toHaveBeenCalledTimes(1);
  });

  it('returns the 201 response immediately without waiting for pipeline execution to complete', async () => {
    // Pipeline enqueue is deliberately slow — HTTP response must still be fast
    mockEnqueue.mockImplementation(
      () => new Promise((resolve) => setTimeout(resolve, 5000)),
    );

    const start = Date.now();
    await request(app)
      .post('/api/validation-runs')
      .send({
        branchName: 'feature/fast-response',
        commitSha: VALID_COMMIT_SHA,
      })
      .expect(201);
    const elapsed = Date.now() - start;

    // Should return well before the 5-second pipeline delay (fire-and-forget)
    expect(elapsed).toBeLessThan(1000);
  });

  it('does NOT call pipelineService.enqueue when the request is rejected with 409', async () => {
    const branchName = 'feature/no-enqueue-on-conflict';

    // First run
    await request(app)
      .post('/api/validation-runs')
      .send({ branchName, commitSha: VALID_COMMIT_SHA })
      .expect(201);

    jest.clearAllMocks();
    mockEnqueue.mockResolvedValue(undefined);

    // Second run is rejected
    await request(app)
      .post('/api/validation-runs')
      .send({ branchName, commitSha: VALID_COMMIT_SHA })
      .expect(409);

    await new Promise((resolve) => process.nextTick(resolve));

    // No enqueue should be called for the rejected run
    expect(mockEnqueue).not.toHaveBeenCalled();
  });

  it('does NOT call pipelineService.enqueue when the request is rejected with 400', async () => {
    await request(app)
      .post('/api/validation-runs')
      .send({
        branchName: 'feature/invalid',
        commitSha: 'invalid-sha',
      })
      .expect(400);

    await new Promise((resolve) => process.nextTick(resolve));

    expect(mockEnqueue).not.toHaveBeenCalled();
  });
});

// ── GET /api/validation-runs (list) ─────────────────────────────────────────

describe('GET /api/validation-runs', () => {
  beforeEach(async () => {
    await prisma.$executeRaw`TRUNCATE TABLE run_stages, validation_runs CASCADE`;
    jest.clearAllMocks();
    mockEnqueue.mockResolvedValue(undefined);
  });

  afterAll(async () => {
    // prisma.$disconnect() already called in POST suite afterAll — no-op here
  });

  // ── Empty state ─────────────────────────────────────────────────────────

  it('returns 200 with empty paginated response when no runs exist', async () => {
    const response = await request(app)
      .get('/api/validation-runs')
      .expect(200);

    expect(response.body).toMatchObject({
      data: [],
      total: 0,
      page: 1,
      limit: 20,
    });
  });

  it('returns data array as an empty array when no runs exist', async () => {
    const response = await request(app)
      .get('/api/validation-runs')
      .expect(200);

    expect(Array.isArray(response.body.data)).toBe(true);
    expect(response.body.data).toHaveLength(0);
  });

  // ── Pagination with 25 seeded runs ───────────────────────────────────────

  it('returns first page of 20 runs with correct metadata when 25 runs are seeded', async () => {
    // Seed 25 runs on different branches to avoid branch-lock conflicts
    for (let i = 0; i < 25; i++) {
      await prisma.validationRun.create({
        data: {
          branchName: `feature/seed-branch-${i}`,
          commitSha: VALID_COMMIT_SHA,
          userId: '',
          config: {},
          status: 'PENDING',
        },
      });
    }

    const response = await request(app)
      .get('/api/validation-runs')
      .expect(200);

    expect(response.body.total).toBe(25);
    expect(response.body.page).toBe(1);
    expect(response.body.limit).toBe(20);
    expect(response.body.data).toHaveLength(20);
  });

  it('returns second page with remaining 5 runs when page=2 and limit=20 with 25 runs seeded', async () => {
    for (let i = 0; i < 25; i++) {
      await prisma.validationRun.create({
        data: {
          branchName: `feature/page-test-branch-${i}`,
          commitSha: VALID_COMMIT_SHA,
          userId: '',
          config: {},
          status: 'PENDING',
        },
      });
    }

    const response = await request(app)
      .get('/api/validation-runs?page=2')
      .expect(200);

    expect(response.body.total).toBe(25);
    expect(response.body.page).toBe(2);
    expect(response.body.data).toHaveLength(5);
  });

  it('returns runs ordered by createdAt descending (newest first)', async () => {
    for (let i = 0; i < 3; i++) {
      await prisma.validationRun.create({
        data: {
          branchName: `feature/order-test-${i}`,
          commitSha: VALID_COMMIT_SHA,
          userId: '',
          config: {},
          status: 'PENDING',
        },
      });
    }

    const response = await request(app)
      .get('/api/validation-runs')
      .expect(200);

    const dates = response.body.data.map((r: { createdAt: string }) => new Date(r.createdAt).getTime());
    for (let i = 1; i < dates.length; i++) {
      expect(dates[i - 1]).toBeGreaterThanOrEqual(dates[i]);
    }
  });

  // ── Filtering ────────────────────────────────────────────────────────────

  it('returns only matching runs when filtering by branch=main', async () => {
    await prisma.validationRun.create({
      data: { branchName: 'main', commitSha: VALID_COMMIT_SHA, userId: '', config: {}, status: 'PENDING' },
    });
    await prisma.validationRun.create({
      data: { branchName: 'feature/other', commitSha: VALID_COMMIT_SHA, userId: '', config: {}, status: 'PENDING' },
    });

    const response = await request(app)
      .get('/api/validation-runs?branch=main')
      .expect(200);

    expect(response.body.total).toBe(1);
    expect(response.body.data).toHaveLength(1);
    expect(response.body.data[0].branchName).toBe('main');
  });

  it('returns only RUNNING runs when filtering by status=RUNNING', async () => {
    await prisma.validationRun.create({
      data: { branchName: 'main', commitSha: VALID_COMMIT_SHA, userId: '', config: {}, status: 'RUNNING' },
    });
    await prisma.validationRun.create({
      data: { branchName: 'feature/other', commitSha: VALID_COMMIT_SHA, userId: '', config: {}, status: 'PENDING' },
    });

    const response = await request(app)
      .get('/api/validation-runs?status=RUNNING')
      .expect(200);

    expect(response.body.total).toBe(1);
    expect(response.body.data).toHaveLength(1);
    expect(response.body.data[0].status).toBe('RUNNING');
  });

  it('returns only runs matching both branch=main and status=RUNNING when both filters applied', async () => {
    // main + RUNNING — should be returned
    await prisma.validationRun.create({
      data: { branchName: 'main', commitSha: VALID_COMMIT_SHA, userId: '', config: {}, status: 'RUNNING' },
    });
    // main + PENDING — should NOT be returned
    await prisma.validationRun.create({
      data: { branchName: 'main', commitSha: VALID_COMMIT_SHA, userId: '', config: {}, status: 'PENDING' },
    });
    // feature + RUNNING — should NOT be returned
    await prisma.validationRun.create({
      data: { branchName: 'feature/other', commitSha: VALID_COMMIT_SHA, userId: '', config: {}, status: 'RUNNING' },
    });

    const response = await request(app)
      .get('/api/validation-runs?branch=main&status=RUNNING')
      .expect(200);

    expect(response.body.total).toBe(1);
    expect(response.body.data).toHaveLength(1);
    expect(response.body.data[0].branchName).toBe('main');
    expect(response.body.data[0].status).toBe('RUNNING');
  });

  it('returns runs matching a partial commitSha prefix when filtering by commitSha', async () => {
    const sha1 = 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa';
    const sha2 = 'bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb';

    await prisma.validationRun.create({
      data: { branchName: 'feature/sha-a', commitSha: sha1, userId: '', config: {}, status: 'PENDING' },
    });
    await prisma.validationRun.create({
      data: { branchName: 'feature/sha-b', commitSha: sha2, userId: '', config: {}, status: 'PENDING' },
    });

    const response = await request(app)
      .get('/api/validation-runs?commitSha=aaaa')
      .expect(200);

    expect(response.body.total).toBe(1);
    expect(response.body.data[0].commitSha).toBe(sha1);
  });

  // ── Limit param / pagination boundaries ─────────────────────────────────

  it('respects limit param and returns only that many results', async () => {
    for (let i = 0; i < 10; i++) {
      await prisma.validationRun.create({
        data: { branchName: `feature/limit-${i}`, commitSha: VALID_COMMIT_SHA, userId: '', config: {}, status: 'PENDING' },
      });
    }

    const response = await request(app)
      .get('/api/validation-runs?limit=5')
      .expect(200);

    expect(response.body.limit).toBe(5);
    expect(response.body.data).toHaveLength(5);
    expect(response.body.total).toBe(10);
  });

  it('clamps limit to maximum of 100 when limit=200 is passed', async () => {
    const response = await request(app)
      .get('/api/validation-runs?limit=200')
      .expect(200);

    expect(response.body.limit).toBeLessThanOrEqual(100);
  });

  it('clamps limit to minimum of 1 when limit=0 is passed', async () => {
    const response = await request(app)
      .get('/api/validation-runs?limit=0')
      .expect(200);

    expect(response.body.limit).toBeGreaterThanOrEqual(1);
  });

  it('uses default limit of 20 when limit param is not provided', async () => {
    const response = await request(app)
      .get('/api/validation-runs')
      .expect(200);

    expect(response.body.limit).toBe(20);
  });

  it('uses default page of 1 when page param is not provided', async () => {
    const response = await request(app)
      .get('/api/validation-runs')
      .expect(200);

    expect(response.body.page).toBe(1);
  });
});

// ── GET /api/validation-runs/:id (getById) ───────────────────────────────────

describe('GET /api/validation-runs/:id', () => {
  beforeEach(async () => {
    await prisma.$executeRaw`TRUNCATE TABLE run_stages, validation_runs CASCADE`;
    jest.clearAllMocks();
    mockEnqueue.mockResolvedValue(undefined);
  });

  // ── Happy path ─────────────────────────────────────────────────────────

  it('returns 200 with run and stages for a valid run id', async () => {
    const run = await prisma.validationRun.create({
      data: {
        branchName: 'feature/get-by-id',
        commitSha: VALID_COMMIT_SHA,
        userId: '',
        config: {},
        status: 'PENDING',
      },
    });

    await prisma.runStage.createMany({
      data: [
        { runId: run.id, stageName: 'SETUP', stageIndex: 0 },
        { runId: run.id, stageName: 'LINT', stageIndex: 1 },
        { runId: run.id, stageName: 'TEST', stageIndex: 2 },
        { runId: run.id, stageName: 'BUILD', stageIndex: 3 },
        { runId: run.id, stageName: 'DEPLOY', stageIndex: 4 },
      ],
    });

    const response = await request(app)
      .get(`/api/validation-runs/${run.id}`)
      .expect(200);

    expect(response.body).toHaveProperty('run');
    expect(response.body).toHaveProperty('stages');
    expect(response.body.run.id).toBe(run.id);
    expect(Array.isArray(response.body.stages)).toBe(true);
  });

  it('returns stages array with 5 items (one per stage type) for a fully staged run', async () => {
    const run = await prisma.validationRun.create({
      data: {
        branchName: 'feature/five-stages',
        commitSha: VALID_COMMIT_SHA,
        userId: '',
        config: {},
        status: 'PENDING',
      },
    });

    await prisma.runStage.createMany({
      data: [
        { runId: run.id, stageName: 'SETUP', stageIndex: 0 },
        { runId: run.id, stageName: 'LINT', stageIndex: 1 },
        { runId: run.id, stageName: 'TEST', stageIndex: 2 },
        { runId: run.id, stageName: 'BUILD', stageIndex: 3 },
        { runId: run.id, stageName: 'DEPLOY', stageIndex: 4 },
      ],
    });

    const response = await request(app)
      .get(`/api/validation-runs/${run.id}`)
      .expect(200);

    expect(response.body.stages).toHaveLength(5);
  });

  it('returns correct run fields (id, branchName, commitSha, status) in the run object', async () => {
    const run = await prisma.validationRun.create({
      data: {
        branchName: 'feature/field-check',
        commitSha: VALID_COMMIT_SHA,
        userId: '',
        config: {},
        status: 'RUNNING',
      },
    });

    const response = await request(app)
      .get(`/api/validation-runs/${run.id}`)
      .expect(200);

    expect(response.body.run).toMatchObject({
      id: run.id,
      branchName: 'feature/field-check',
      commitSha: VALID_COMMIT_SHA,
      status: 'RUNNING',
    });
  });

  it('returns empty stages array when no stages have been created for the run', async () => {
    const run = await prisma.validationRun.create({
      data: {
        branchName: 'feature/no-stages',
        commitSha: VALID_COMMIT_SHA,
        userId: '',
        config: {},
        status: 'PENDING',
      },
    });

    const response = await request(app)
      .get(`/api/validation-runs/${run.id}`)
      .expect(200);

    expect(response.body.stages).toHaveLength(0);
  });

  it('returns stages with correct fields (id, runId, stageName, stageIndex, status)', async () => {
    const run = await prisma.validationRun.create({
      data: {
        branchName: 'feature/stage-fields',
        commitSha: VALID_COMMIT_SHA,
        userId: '',
        config: {},
        status: 'PENDING',
      },
    });

    await prisma.runStage.createMany({
      data: [
        { runId: run.id, stageName: 'SETUP', stageIndex: 0 },
        { runId: run.id, stageName: 'LINT', stageIndex: 1 },
        { runId: run.id, stageName: 'TEST', stageIndex: 2 },
        { runId: run.id, stageName: 'BUILD', stageIndex: 3 },
        { runId: run.id, stageName: 'DEPLOY', stageIndex: 4 },
      ],
    });

    const response = await request(app)
      .get(`/api/validation-runs/${run.id}`)
      .expect(200);

    const firstStage = response.body.stages[0];
    expect(firstStage).toMatchObject({
      id: expect.any(String),
      runId: run.id,
      stageName: expect.any(String),
      stageIndex: expect.any(Number),
      status: expect.any(String),
    });
  });

  // ── Not found ────────────────────────────────────────────────────────────

  it('returns 404 for a run id that does not exist', async () => {
    const unknownId = '00000000-0000-0000-0000-000000000000';

    await request(app)
      .get(`/api/validation-runs/${unknownId}`)
      .expect(404);
  });

  it('returns a JSON body with an error field on 404', async () => {
    const unknownId = '00000000-0000-0000-0000-000000000001';

    const response = await request(app)
      .get(`/api/validation-runs/${unknownId}`)
      .expect(404);

    expect(response.body).toHaveProperty('error');
  });
});
