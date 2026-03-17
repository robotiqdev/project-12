/**
 * Unit tests for CleanupService.runAgeBased (TASK-4616)
 *
 * Tests are written against the public interface of cleanup.service.ts.
 * Prisma and cleanupJobModel are mocked so no real database is needed.
 */

import { runAgeBased } from '../../../src/services/cleanup.service';
import { CleanupJobType } from '../../../src/types/cleanup.types';

// ---- Mocks ----

jest.mock('../../../src/db/client', () => ({
  prisma: {
    validationRun: {
      deleteMany: jest.fn(),
    },
  },
}));

jest.mock('../../../src/models/cleanup-job.model');

import { prisma } from '../../../src/db/client';
import * as cleanupJobModel from '../../../src/models/cleanup-job.model';

// ---- Test setup ----

const mockJob = {
  id: 'job-abc-123',
  type: CleanupJobType.AGE_BASED,
  status: 'PENDING',
  deletedCount: 0,
  startedAt: new Date(),
  completedAt: null,
  errorMessage: null,
};

describe('CleanupService', () => {
  describe('runAgeBased', () => {
    beforeEach(() => {
      jest.clearAllMocks();

      (cleanupJobModel.create as jest.Mock).mockResolvedValue(mockJob);
      (cleanupJobModel.start as jest.Mock).mockResolvedValue({
        ...mockJob,
        status: 'RUNNING',
      });
      (cleanupJobModel.complete as jest.Mock).mockResolvedValue({
        ...mockJob,
        status: 'DONE',
        deletedCount: 5,
      });
      (cleanupJobModel.fail as jest.Mock).mockResolvedValue({
        ...mockJob,
        status: 'FAILED',
      });
      (prisma.validationRun.deleteMany as jest.Mock).mockResolvedValue({
        count: 5,
      });
    });

    // ---- Cutoff date calculation ----

    it('deletes validation runs with createdAt older than 90 days', async () => {
      const beforeCall = Date.now();
      await runAgeBased();

      expect(prisma.validationRun.deleteMany).toHaveBeenCalledWith(
        expect.objectContaining({
          where: expect.objectContaining({
            createdAt: expect.objectContaining({
              lt: expect.any(Date),
            }),
          }),
        }),
      );

      const callArgs = (prisma.validationRun.deleteMany as jest.Mock).mock
        .calls[0][0];
      const cutoff = callArgs.where.createdAt.lt as Date;
      const ninetyDaysMs = 90 * 24 * 60 * 60 * 1000;
      const elapsed = beforeCall - cutoff.getTime();

      // Cutoff should be approximately 90 days before the call time
      expect(elapsed).toBeGreaterThanOrEqual(ninetyDaysMs - 1000);
      expect(elapsed).toBeLessThanOrEqual(ninetyDaysMs + 1000);
    });

    // ---- RUNNING runs must be excluded ----

    it('does NOT delete RUNNING runs even if they are older than 90 days', async () => {
      await runAgeBased();

      const callArgs = (prisma.validationRun.deleteMany as jest.Mock).mock
        .calls[0][0];
      expect(callArgs.where.status).toEqual(
        expect.objectContaining({ not: 'RUNNING' }),
      );
    });

    // ---- Cascade delete of RunStages ----

    it('cascade-deletes associated RunStage records via a single deleteMany call', async () => {
      await runAgeBased();

      // Prisma's onDelete: Cascade on RunStage → ValidationRun relation
      // means deleting ValidationRun automatically removes RunStages.
      // The service must NOT manually delete RunStage records.
      expect(prisma.validationRun.deleteMany).toHaveBeenCalledTimes(1);
    });

    // ---- CleanupJob lifecycle ----

    it('creates a CleanupJob of type AGE_BASED', async () => {
      await runAgeBased();

      expect(cleanupJobModel.create).toHaveBeenCalledWith(
        CleanupJobType.AGE_BASED,
      );
    });

    it('calls start on the CleanupJob before running the deletion', async () => {
      await runAgeBased();

      expect(cleanupJobModel.start).toHaveBeenCalledWith(mockJob.id);

      const startOrder = (cleanupJobModel.start as jest.Mock).mock
        .invocationCallOrder[0];
      const deleteOrder = (prisma.validationRun.deleteMany as jest.Mock).mock
        .invocationCallOrder[0];
      expect(startOrder).toBeLessThan(deleteOrder);
    });

    it('calls complete with the correct deletedCount after deletion', async () => {
      (prisma.validationRun.deleteMany as jest.Mock).mockResolvedValue({
        count: 7,
      });

      await runAgeBased();

      expect(cleanupJobModel.complete).toHaveBeenCalledWith(mockJob.id, 7);
    });

    it('calls complete with deletedCount of 0 when no runs are deleted', async () => {
      (prisma.validationRun.deleteMany as jest.Mock).mockResolvedValue({
        count: 0,
      });

      await runAgeBased();

      expect(cleanupJobModel.complete).toHaveBeenCalledWith(mockJob.id, 0);
    });

    // ---- Return value ----

    it('returns a CleanupResult with jobId, type, deletedCount, and duration', async () => {
      (prisma.validationRun.deleteMany as jest.Mock).mockResolvedValue({
        count: 3,
      });

      const result = await runAgeBased();

      expect(result.jobId).toBe(mockJob.id);
      expect(result.type).toBe(CleanupJobType.AGE_BASED);
      expect(result.deletedCount).toBe(3);
      expect(typeof result.duration).toBe('number');
      expect(result.duration).toBeGreaterThanOrEqual(0);
    });

    // ---- Error handling ----

    it('calls cleanupJobModel.fail with the error message when Prisma delete throws', async () => {
      const error = new Error('Database connection failed');
      (prisma.validationRun.deleteMany as jest.Mock).mockRejectedValue(error);

      await expect(runAgeBased()).rejects.toThrow();

      expect(cleanupJobModel.fail).toHaveBeenCalledWith(
        mockJob.id,
        error.message,
      );
    });

    it('does not call complete when an error occurs during deletion', async () => {
      const error = new Error('DB error');
      (prisma.validationRun.deleteMany as jest.Mock).mockRejectedValue(error);

      try {
        await runAgeBased();
      } catch (_) {
        // expected
      }

      expect(cleanupJobModel.complete).not.toHaveBeenCalled();
    });

    it('re-throws the error after calling fail so callers can handle it', async () => {
      const error = new Error('Unexpected DB failure');
      (prisma.validationRun.deleteMany as jest.Mock).mockRejectedValue(error);

      await expect(runAgeBased()).rejects.toThrow('Unexpected DB failure');
    });
  });
});
