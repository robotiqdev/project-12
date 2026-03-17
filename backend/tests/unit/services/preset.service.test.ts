/**
 * Unit tests for preset.service.ts
 * Tests the PresetService business logic including:
 * - createPreset: enforces max-5 presets per user
 * - getPresets: returns ordered presets for a user
 * - deletePreset: removes a preset owned by the user
 *
 * Prisma is mocked to isolate service logic from the database.
 */

import { PresetService, PresetLimitExceededError } from '../../../src/services/preset.service';
import { ValidationPresetModel } from '../../../src/models/validation-preset.model';
import type { CreatePresetInput, ValidationPresetRecord } from '../../../src/types/preset.types';
import { PRESET_MAX_PER_USER } from '../../../src/types/preset.types';

// Mock the model layer
jest.mock('../../../src/models/validation-preset.model');

const MockedValidationPresetModel = ValidationPresetModel as jest.Mocked<typeof ValidationPresetModel>;

function makePreset(overrides: Partial<ValidationPresetRecord> = {}): ValidationPresetRecord {
  return {
    id: 'preset-uuid-1',
    userId: 'user-123',
    name: 'My Preset',
    branchName: 'main',
    commitSha: null,
    envVars: {},
    featureFlags: {},
    createdAt: new Date('2024-01-01T00:00:00Z'),
    updatedAt: new Date('2024-01-01T00:00:00Z'),
    ...overrides,
  };
}

describe('PresetService', () => {
  let service: PresetService;

  beforeEach(() => {
    jest.clearAllMocks();
    service = new PresetService();
  });

  // ---- createPreset ----

  describe('createPreset', () => {
    const validInput: CreatePresetInput = {
      userId: 'user-123',
      name: 'Test Preset',
      branchName: 'feature/test',
      commitSha: null,
      envVars: { NODE_ENV: 'test' },
      featureFlags: { newDashboard: true },
    };

    it('should create a preset when user has 0 existing presets', async () => {
      MockedValidationPresetModel.countByUser = jest.fn().mockResolvedValue(0);
      MockedValidationPresetModel.create = jest.fn().mockResolvedValue(makePreset());

      const result = await service.createPreset(validInput);

      expect(MockedValidationPresetModel.countByUser).toHaveBeenCalledWith('user-123');
      expect(MockedValidationPresetModel.create).toHaveBeenCalledWith(validInput);
      expect(result).toBeDefined();
    });

    it('should create a preset when user has 1 existing preset', async () => {
      MockedValidationPresetModel.countByUser = jest.fn().mockResolvedValue(1);
      MockedValidationPresetModel.create = jest.fn().mockResolvedValue(makePreset());

      await expect(service.createPreset(validInput)).resolves.toBeDefined();
      expect(MockedValidationPresetModel.create).toHaveBeenCalledTimes(1);
    });

    it('should create a preset when user has 2 existing presets', async () => {
      MockedValidationPresetModel.countByUser = jest.fn().mockResolvedValue(2);
      MockedValidationPresetModel.create = jest.fn().mockResolvedValue(makePreset());

      await expect(service.createPreset(validInput)).resolves.toBeDefined();
      expect(MockedValidationPresetModel.create).toHaveBeenCalledTimes(1);
    });

    it('should create a preset when user has 3 existing presets', async () => {
      MockedValidationPresetModel.countByUser = jest.fn().mockResolvedValue(3);
      MockedValidationPresetModel.create = jest.fn().mockResolvedValue(makePreset());

      await expect(service.createPreset(validInput)).resolves.toBeDefined();
      expect(MockedValidationPresetModel.create).toHaveBeenCalledTimes(1);
    });

    it('should create a preset when user has 4 existing presets (boundary)', async () => {
      MockedValidationPresetModel.countByUser = jest.fn().mockResolvedValue(4);
      MockedValidationPresetModel.create = jest.fn().mockResolvedValue(makePreset());

      await expect(service.createPreset(validInput)).resolves.toBeDefined();
      expect(MockedValidationPresetModel.create).toHaveBeenCalledTimes(1);
    });

    it('should throw PresetLimitExceededError when user already has 5 presets', async () => {
      MockedValidationPresetModel.countByUser = jest.fn().mockResolvedValue(5);
      MockedValidationPresetModel.create = jest.fn();

      await expect(service.createPreset(validInput)).rejects.toThrow(PresetLimitExceededError);
      expect(MockedValidationPresetModel.create).not.toHaveBeenCalled();
    });

    it('should throw PresetLimitExceededError when user has more than 5 presets', async () => {
      MockedValidationPresetModel.countByUser = jest.fn().mockResolvedValue(7);
      MockedValidationPresetModel.create = jest.fn();

      await expect(service.createPreset(validInput)).rejects.toThrow(PresetLimitExceededError);
      expect(MockedValidationPresetModel.create).not.toHaveBeenCalled();
    });

    it('PresetLimitExceededError should have a status code property', async () => {
      MockedValidationPresetModel.countByUser = jest.fn().mockResolvedValue(5);
      MockedValidationPresetModel.create = jest.fn();

      let thrownError: unknown;
      try {
        await service.createPreset(validInput);
      } catch (err) {
        thrownError = err;
      }

      expect(thrownError).toBeInstanceOf(PresetLimitExceededError);
      expect((thrownError as PresetLimitExceededError).statusCode).toBeDefined();
    });

    it('PRESET_MAX_PER_USER constant should equal 5', () => {
      expect(PRESET_MAX_PER_USER).toBe(5);
    });
  });

  // ---- getPresets ----

  describe('getPresets', () => {
    it('should return all presets for a userId', async () => {
      const presets = [
        makePreset({ id: 'preset-1', createdAt: new Date('2024-01-03T00:00:00Z') }),
        makePreset({ id: 'preset-2', createdAt: new Date('2024-01-02T00:00:00Z') }),
        makePreset({ id: 'preset-3', createdAt: new Date('2024-01-01T00:00:00Z') }),
      ];
      MockedValidationPresetModel.findAllByUser = jest.fn().mockResolvedValue(presets);

      const result = await service.getPresets('user-123');

      expect(MockedValidationPresetModel.findAllByUser).toHaveBeenCalledWith('user-123');
      expect(result).toHaveLength(3);
    });

    it('should return presets ordered by createdAt descending (newest first)', async () => {
      const presets = [
        makePreset({ id: 'preset-newest', createdAt: new Date('2024-01-03T00:00:00Z') }),
        makePreset({ id: 'preset-middle', createdAt: new Date('2024-01-02T00:00:00Z') }),
        makePreset({ id: 'preset-oldest', createdAt: new Date('2024-01-01T00:00:00Z') }),
      ];
      MockedValidationPresetModel.findAllByUser = jest.fn().mockResolvedValue(presets);

      const result = await service.getPresets('user-123');

      // The model is expected to return them in desc order; verify the order is preserved
      expect(result[0].id).toBe('preset-newest');
      expect(result[1].id).toBe('preset-middle');
      expect(result[2].id).toBe('preset-oldest');
    });

    it('should return empty array when user has no presets', async () => {
      MockedValidationPresetModel.findAllByUser = jest.fn().mockResolvedValue([]);

      const result = await service.getPresets('user-123');

      expect(result).toEqual([]);
    });

    it('should only return presets for the specified userId', async () => {
      const userPresets = [makePreset({ userId: 'user-123' })];
      MockedValidationPresetModel.findAllByUser = jest.fn().mockResolvedValue(userPresets);

      const result = await service.getPresets('user-123');

      expect(MockedValidationPresetModel.findAllByUser).toHaveBeenCalledWith('user-123');
      expect(result.every(p => p.userId === 'user-123')).toBe(true);
    });
  });

  // ---- deletePreset ----

  describe('deletePreset', () => {
    it('should delete a preset that belongs to the user', async () => {
      MockedValidationPresetModel.findByIdAndUser = jest.fn().mockResolvedValue(makePreset({ id: 'preset-uuid-1' }));
      MockedValidationPresetModel.delete = jest.fn().mockResolvedValue(undefined);

      await service.deletePreset('preset-uuid-1', 'user-123');

      expect(MockedValidationPresetModel.findByIdAndUser).toHaveBeenCalledWith('preset-uuid-1', 'user-123');
      expect(MockedValidationPresetModel.delete).toHaveBeenCalledWith('preset-uuid-1', 'user-123');
    });

    it('should throw a 404 error when preset does not belong to the user', async () => {
      MockedValidationPresetModel.findByIdAndUser = jest.fn().mockResolvedValue(null);
      MockedValidationPresetModel.delete = jest.fn();

      let thrownError: unknown;
      try {
        await service.deletePreset('preset-uuid-1', 'other-user');
      } catch (err) {
        thrownError = err;
      }

      expect(thrownError).toBeDefined();
      // The error should have a status code of 404
      expect((thrownError as { statusCode?: number }).statusCode).toBe(404);
      expect(MockedValidationPresetModel.delete).not.toHaveBeenCalled();
    });

    it('should throw a 404 error when preset does not exist', async () => {
      MockedValidationPresetModel.findByIdAndUser = jest.fn().mockResolvedValue(null);
      MockedValidationPresetModel.delete = jest.fn();

      await expect(
        service.deletePreset('nonexistent-preset', 'user-123')
      ).rejects.toMatchObject({ statusCode: 404 });

      expect(MockedValidationPresetModel.delete).not.toHaveBeenCalled();
    });

    it('should verify ownership before deletion via findByIdAndUser', async () => {
      MockedValidationPresetModel.findByIdAndUser = jest.fn().mockResolvedValue(makePreset());
      MockedValidationPresetModel.delete = jest.fn().mockResolvedValue(undefined);

      await service.deletePreset('preset-uuid-1', 'user-123');

      // Ownership check must happen before delete
      const findOrder = (MockedValidationPresetModel.findByIdAndUser as jest.Mock).mock.invocationCallOrder[0];
      const deleteOrder = (MockedValidationPresetModel.delete as jest.Mock).mock.invocationCallOrder[0];
      expect(findOrder).toBeLessThan(deleteOrder);
    });
  });
});
