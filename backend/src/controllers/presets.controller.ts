/**
 * Presets Controller - TASK-4602
 *
 * Handles HTTP requests for preset CRUD endpoints:
 *   GET    /api/presets
 *   POST   /api/presets
 *   PUT    /api/presets/:id
 *   DELETE /api/presets/:id
 *
 * All operations are scoped to the authenticated userId extracted from
 * the request context (set by auth middleware via x-user-id header).
 */

import { Request, Response } from 'express';
import { z } from 'zod';
import { presetService, PresetLimitExceededError } from '../services/preset.service';

// ---------------------------------------------------------------------------
// Zod validation schemas
// ---------------------------------------------------------------------------

const createPresetSchema = z.object({
  name: z.string().min(1).max(100),
  branchName: z.string().optional(),
  commitSha: z.string().nullable().optional(),
  envVars: z.record(z.string()).optional(),
  featureFlags: z.record(z.boolean()).optional(),
});

const updatePresetSchema = z.object({
  name: z.string().min(1).max(100).optional(),
  branchName: z.string().optional(),
  commitSha: z.string().nullable().optional(),
  envVars: z.record(z.string()).optional(),
  featureFlags: z.record(z.boolean()).optional(),
});

// ---------------------------------------------------------------------------
// Helper: extract userId from request (set by auth middleware)
// ---------------------------------------------------------------------------

function getUserId(req: Request): string {
  return (req.headers['x-user-id'] as string) || '';
}

// ---------------------------------------------------------------------------
// GET /api/presets
// ---------------------------------------------------------------------------

export async function getPresets(req: Request, res: Response): Promise<void> {
  try {
    const userId = getUserId(req);
    const presets = await presetService.getPresets(userId);
    res.json(presets);
  } catch (err: any) {
    const status = err.statusCode || err.status || 500;
    res.status(status).json({ error: err.message || 'Internal server error' });
  }
}

// ---------------------------------------------------------------------------
// POST /api/presets
// ---------------------------------------------------------------------------

export async function createPreset(req: Request, res: Response): Promise<void> {
  const parsed = createPresetSchema.safeParse(req.body);
  if (!parsed.success) {
    res.status(422).json({ error: 'VALIDATION_ERROR', details: parsed.error.issues });
    return;
  }

  try {
    const userId = getUserId(req);
    const preset = await presetService.createPreset({
      userId,
      name: parsed.data.name,
      branchName: parsed.data.branchName ?? '',
      commitSha: parsed.data.commitSha ?? null,
      envVars: parsed.data.envVars ?? {},
      featureFlags: parsed.data.featureFlags ?? {},
    });
    res.status(201).json(preset);
  } catch (err: any) {
    if (err instanceof PresetLimitExceededError) {
      res.status(422).json({ error: 'PRESET_LIMIT_EXCEEDED' });
      return;
    }
    const status = err.statusCode || err.status || 500;
    res.status(status).json({ error: err.message || 'Internal server error' });
  }
}

// ---------------------------------------------------------------------------
// PUT /api/presets/:id
// ---------------------------------------------------------------------------

export async function updatePreset(req: Request, res: Response): Promise<void> {
  const parsed = updatePresetSchema.safeParse(req.body);
  if (!parsed.success) {
    res.status(422).json({ error: 'VALIDATION_ERROR', details: parsed.error.issues });
    return;
  }

  try {
    const userId = getUserId(req);
    const { id } = req.params;
    const preset = await presetService.updatePreset(id, userId, parsed.data);
    res.status(200).json(preset);
  } catch (err: any) {
    const status = err.statusCode || err.status || 500;
    if (status === 403 || status === 404) {
      res.status(status).json({ error: err.message || 'Not found' });
      return;
    }
    res.status(status).json({ error: err.message || 'Internal server error' });
  }
}

// ---------------------------------------------------------------------------
// DELETE /api/presets/:id
// ---------------------------------------------------------------------------

export async function deletePreset(req: Request, res: Response): Promise<void> {
  try {
    const userId = getUserId(req);
    const { id } = req.params;
    await presetService.deletePreset(id, userId);
    res.status(204).send();
  } catch (err: any) {
    const status = err.statusCode || err.status || 500;
    if (status === 403 || status === 404) {
      res.status(status).json({ error: err.message || 'Not found' });
      return;
    }
    res.status(status).json({ error: err.message || 'Internal server error' });
  }
}
