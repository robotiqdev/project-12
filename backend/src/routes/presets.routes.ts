/**
 * Presets Routes - TASK-4602
 *
 * Defines CRUD routes for the preset API,
 * scoped under /api/presets.
 *
 * Routes:
 *   GET    /api/presets
 *   POST   /api/presets
 *   PUT    /api/presets/:id
 *   DELETE /api/presets/:id
 */

import { Router } from 'express';
import {
  getPresets,
  createPreset,
  updatePreset,
  deletePreset,
} from '../controllers/presets.controller';

const router = Router();

// GET /api/presets
router.get('/', getPresets);

// POST /api/presets
router.post('/', createPreset);

// PUT /api/presets/:id
router.put('/:id', updatePreset);

// DELETE /api/presets/:id
router.delete('/:id', deletePreset);

export default router;
