/**
 * Integration tests for Preset CRUD API Endpoints (TASK-4602).
 *
 * Tests:
 *   GET /api/presets              -> [] for a new user, array of preset objects otherwise
 *   POST /api/presets             -> 201 with created preset record
 *   POST /api/presets (limit hit) -> 422 with { error: 'PRESET_LIMIT_EXCEEDED' }
 *   PUT /api/presets/:id          -> 200 with updated preset
 *   PUT /api/presets/:id (other)  -> 403/404 for another user's preset
 *   DELETE /api/presets/:id       -> 204 with empty body
 *   Zod validation                -> 422 when name is missing or exceeds 100 chars
 *
 * Mocks the presetService so no real database access is required.
 */

import request from 'supertest';
import { jest } from '@jest/globals';

// Mock the preset service BEFORE importing the app so that the Express routes
// use the mock instead of the real service.
jest.mock('../../src/services/preset.service', () => ({
  presetService: {
    getPresets: jest.fn(),
    createPreset: jest.fn(),
    updatePreset: jest.fn(),
    deletePreset: jest.fn(),
  },
  PresetLimitExceededError: class PresetLimitExceededError extends Error {
    statusCode = 422;
    constructor() {
      super('You may not have more than 5 presets');
      this.name = 'PresetLimitExceededError';
    }
  },
}));

import app from '../../src/app';

// ---------------------------------------------------------------------------
// Test fixtures and mock references
// ---------------------------------------------------------------------------

const TEST_USER_ID = 'user-123';
const OTHER_USER_ID = 'other-user-456';

// Header providing authenticated userId for test requests
const authHeader = { 'x-user-id': TEST_USER_ID };

const testPreset = {
  id: 'preset-001',
  userId: TEST_USER_ID,
  name: 'Test Preset',
  branchName: 'main',
  commitSha: null,
  envVars: { NODE_ENV: 'test' },
  featureFlags: { featureA: true },
  createdAt: new Date('2024-01-01T00:00:00Z').toISOString(),
  updatedAt: new Date('2024-01-01T00:00:00Z').toISOString(),
};

// Get typed references to the mock functions
const { presetService: mockService, PresetLimitExceededError } =
  jest.requireMock('../../src/services/preset.service') as {
    presetService: {
      getPresets: jest.Mock;
      createPreset: jest.Mock;
      updatePreset: jest.Mock;
      deletePreset: jest.Mock;
    };
    PresetLimitExceededError: new () => Error & { statusCode: number };
  };

// ---------------------------------------------------------------------------
// GET /api/presets
// ---------------------------------------------------------------------------

describe('GET /api/presets', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockService.getPresets.mockResolvedValue([]);
  });

  it('returns 200 with an empty array for a new user', async () => {
    const res = await request(app)
      .get('/api/presets')
      .set(authHeader);

    expect(res.status).toBe(200);
    expect(Array.isArray(res.body)).toBe(true);
    expect(res.body).toHaveLength(0);
  });

  it('returns 200 with an array of preset objects when user has presets', async () => {
    mockService.getPresets.mockResolvedValue([testPreset]);

    const res = await request(app)
      .get('/api/presets')
      .set(authHeader);

    expect(res.status).toBe(200);
    expect(Array.isArray(res.body)).toBe(true);
    expect(res.body.length).toBeGreaterThan(0);
  });

  it('each preset object has id, name, userId, envVars, featureFlags fields', async () => {
    mockService.getPresets.mockResolvedValue([testPreset]);

    const res = await request(app)
      .get('/api/presets')
      .set(authHeader);

    expect(res.status).toBe(200);
    for (const preset of res.body) {
      expect(preset).toHaveProperty('id');
      expect(preset).toHaveProperty('name');
      expect(preset).toHaveProperty('userId');
      expect(preset).toHaveProperty('envVars');
      expect(preset).toHaveProperty('featureFlags');
    }
  });

  it('passes the authenticated userId to getPresets', async () => {
    await request(app)
      .get('/api/presets')
      .set(authHeader);

    expect(mockService.getPresets).toHaveBeenCalledWith(TEST_USER_ID);
  });
});

// ---------------------------------------------------------------------------
// POST /api/presets
// ---------------------------------------------------------------------------

describe('POST /api/presets', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockService.createPreset.mockResolvedValue(testPreset);
  });

  it('creates a preset and returns 201 with the preset record', async () => {
    const res = await request(app)
      .post('/api/presets')
      .set(authHeader)
      .send({ name: 'Test Preset', branchName: 'main', envVars: {}, featureFlags: {} });

    expect(res.status).toBe(201);
    expect(res.body).toHaveProperty('id');
    expect(res.body).toHaveProperty('name');
  });

  it('returned preset record contains expected fields', async () => {
    const res = await request(app)
      .post('/api/presets')
      .set(authHeader)
      .send({ name: 'Test Preset' });

    expect(res.status).toBe(201);
    expect(res.body).toHaveProperty('id');
    expect(res.body).toHaveProperty('userId');
    expect(res.body).toHaveProperty('name');
    expect(res.body).toHaveProperty('envVars');
    expect(res.body).toHaveProperty('featureFlags');
  });

  it('returns 422 with { error: PRESET_LIMIT_EXCEEDED } when user already has 5 presets', async () => {
    mockService.createPreset.mockRejectedValueOnce(new PresetLimitExceededError());

    const res = await request(app)
      .post('/api/presets')
      .set(authHeader)
      .send({ name: 'New Preset' });

    expect(res.status).toBe(422);
    expect(res.body).toHaveProperty('error');
    expect(res.body.error).toBe('PRESET_LIMIT_EXCEEDED');
  });

  it('response body contains PRESET_LIMIT_EXCEEDED key (not a generic message) when limit hit', async () => {
    mockService.createPreset.mockRejectedValueOnce(new PresetLimitExceededError());

    const res = await request(app)
      .post('/api/presets')
      .set(authHeader)
      .send({ name: 'Sixth Preset' });

    expect(res.status).toBe(422);
    expect(res.body.error).toBe('PRESET_LIMIT_EXCEEDED');
  });

  // ---- Zod validation ----

  it('returns 422 when name field is missing (zod validation)', async () => {
    const res = await request(app)
      .post('/api/presets')
      .set(authHeader)
      .send({ branchName: 'main', envVars: {} });

    expect(res.status).toBe(422);
  });

  it('returns 422 when name is an empty string (min 1 char zod validation)', async () => {
    const res = await request(app)
      .post('/api/presets')
      .set(authHeader)
      .send({ name: '', branchName: 'main' });

    expect(res.status).toBe(422);
  });

  it('returns 422 when name exceeds 100 characters (zod max length validation)', async () => {
    const longName = 'a'.repeat(101);

    const res = await request(app)
      .post('/api/presets')
      .set(authHeader)
      .send({ name: longName });

    expect(res.status).toBe(422);
  });

  it('accepts a name of exactly 100 characters (boundary - zod max)', async () => {
    const maxName = 'a'.repeat(100);

    const res = await request(app)
      .post('/api/presets')
      .set(authHeader)
      .send({ name: maxName });

    expect(res.status).toBe(201);
  });

  it('accepts optional branchName and commitSha fields', async () => {
    const res = await request(app)
      .post('/api/presets')
      .set(authHeader)
      .send({ name: 'Preset with branch', branchName: 'feature/x', commitSha: 'abc123' });

    expect(res.status).toBe(201);
  });

  it('accepts optional envVars as a key-value record', async () => {
    const res = await request(app)
      .post('/api/presets')
      .set(authHeader)
      .send({ name: 'Preset', envVars: { KEY: 'VALUE', ANOTHER: 'val' } });

    expect(res.status).toBe(201);
  });
});

// ---------------------------------------------------------------------------
// PUT /api/presets/:id
// ---------------------------------------------------------------------------

describe('PUT /api/presets/:id', () => {
  const updatedPreset = { ...testPreset, name: 'Updated Preset', envVars: { KEY: 'value' } };

  beforeEach(() => {
    jest.clearAllMocks();
    mockService.updatePreset.mockResolvedValue(updatedPreset);
  });

  it('updates a preset and returns 200 with the updated record', async () => {
    const res = await request(app)
      .put('/api/presets/preset-001')
      .set(authHeader)
      .send({ name: 'Updated Preset', envVars: { KEY: 'value' } });

    expect(res.status).toBe(200);
    expect(res.body).toHaveProperty('name');
  });

  it('returns the updated name in the response body', async () => {
    const res = await request(app)
      .put('/api/presets/preset-001')
      .set(authHeader)
      .send({ name: 'Updated Preset' });

    expect(res.status).toBe(200);
    expect(res.body.name).toBe('Updated Preset');
  });

  it('returns the updated envVars in the response body', async () => {
    const res = await request(app)
      .put('/api/presets/preset-001')
      .set(authHeader)
      .send({ envVars: { KEY: 'value' } });

    expect(res.status).toBe(200);
    expect(res.body).toHaveProperty('envVars');
  });

  it("returns 403 or 404 when trying to update another user's preset", async () => {
    mockService.updatePreset.mockRejectedValueOnce(
      Object.assign(new Error('Preset not found'), { statusCode: 404 })
    );

    const res = await request(app)
      .put('/api/presets/other-users-preset')
      .set({ 'x-user-id': OTHER_USER_ID })
      .send({ name: 'Hacked Name' });

    expect([403, 404]).toContain(res.status);
  });

  it('returns 403 or 404 for a nonexistent preset id', async () => {
    mockService.updatePreset.mockRejectedValueOnce(
      Object.assign(new Error('Preset not found'), { statusCode: 404 })
    );

    const res = await request(app)
      .put('/api/presets/nonexistent-id')
      .set(authHeader)
      .send({ name: 'New Name' });

    expect([403, 404]).toContain(res.status);
  });

  // ---- Zod validation on update ----

  it('returns 422 when name exceeds 100 characters on update (zod validation)', async () => {
    const longName = 'a'.repeat(101);

    const res = await request(app)
      .put('/api/presets/preset-001')
      .set(authHeader)
      .send({ name: longName });

    expect(res.status).toBe(422);
  });
});

// ---------------------------------------------------------------------------
// DELETE /api/presets/:id
// ---------------------------------------------------------------------------

describe('DELETE /api/presets/:id', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockService.deletePreset.mockResolvedValue(undefined);
  });

  it('deletes a preset and returns 204', async () => {
    const res = await request(app)
      .delete('/api/presets/preset-001')
      .set(authHeader);

    expect(res.status).toBe(204);
  });

  it('returns an empty body on successful deletion', async () => {
    const res = await request(app)
      .delete('/api/presets/preset-001')
      .set(authHeader);

    expect(res.status).toBe(204);
    expect(res.body).toEqual({});
  });

  it('passes the preset id and userId to deletePreset', async () => {
    await request(app)
      .delete('/api/presets/preset-001')
      .set(authHeader);

    expect(mockService.deletePreset).toHaveBeenCalledWith('preset-001', TEST_USER_ID);
  });

  it("returns 403 or 404 when trying to delete another user's preset", async () => {
    mockService.deletePreset.mockRejectedValueOnce(
      Object.assign(new Error('Preset not found'), { statusCode: 404 })
    );

    const res = await request(app)
      .delete('/api/presets/other-users-preset')
      .set({ 'x-user-id': OTHER_USER_ID });

    expect([403, 404]).toContain(res.status);
  });

  it('returns 403 or 404 for a nonexistent preset id', async () => {
    mockService.deletePreset.mockRejectedValueOnce(
      Object.assign(new Error('Preset not found'), { statusCode: 404 })
    );

    const res = await request(app)
      .delete('/api/presets/nonexistent-id')
      .set(authHeader);

    expect([403, 404]).toContain(res.status);
  });
});
