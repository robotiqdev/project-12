import { ValidationPresetModel } from '../models/validation-preset.model';
import { CreatePresetInput, UpdatePresetInput, ValidationPresetRecord, PRESET_MAX_PER_USER } from '../types/preset.types';

export class PresetLimitExceededError extends Error {
  statusCode = 422;

  constructor() {
    super(`You may not have more than ${PRESET_MAX_PER_USER} presets`);
    this.name = 'PresetLimitExceededError';
  }
}

export class PresetService {
  async createPreset(input: CreatePresetInput): Promise<ValidationPresetRecord> {
    const count = await ValidationPresetModel.countByUser(input.userId);
    if (count >= PRESET_MAX_PER_USER) {
      throw new PresetLimitExceededError();
    }
    return ValidationPresetModel.create(input);
  }

  async getPresets(userId: string): Promise<ValidationPresetRecord[]> {
    return ValidationPresetModel.findAllByUser(userId);
  }

  async deletePreset(id: string, userId: string): Promise<void> {
    const preset = await ValidationPresetModel.findByIdAndUser(id, userId);
    if (!preset) {
      const err: NodeJS.ErrnoException = new Error('Preset not found');
      (err as any).statusCode = 404;
      throw err;
    }
    await ValidationPresetModel.delete(id, userId);
  }
}

export const presetService = new PresetService();
