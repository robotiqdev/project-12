/**
 * Unit tests for concurrency.service.ts — TASK-4592
 *
 * Tests the public interface of:
 *   - checkAndLockBranch(branchName): atomically checks whether a new validation
 *     run is allowed on a branch and returns { allowed } or { allowed, conflictingRunId }
 *   - createRun(input): persists a new ValidationRun with status PENDING
 *   - getRunWithStages(id): returns the run with its stages sorted by stageIndex
 */

import {
  checkAndLockBranch,
  createRun,
  getRunWithStages,
} from '../../../src/services/concurrency.service';
import { prisma } from '../../../src/db/client';

jest.mock('../../../src/db/client');

const mockPrisma = prisma as jest.Mocked<typeof prisma>;

describe('ConcurrencyService', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  // ---------------------------------------------------------------------------
  // checkAndLockBranch
  // ---------------------------------------------------------------------------

  describe('checkAndLockBranch', () => {
    it('returns { allowed: true } when no active run exists on the branch', async () => {
      // No PENDING or RUNNING rows found — lock is available
      (mockPrisma.$transaction as jest.Mock).mockImplementation(async (fn: Function) =>
        fn(mockPrisma),
      );
      (mockPrisma.$queryRaw as jest.Mock).mockResolvedValue([]);

      const result = await checkAndLockBranch('feature/my-branch');

      expect(result.allowed).toBe(true);
      expect((result as any).conflictingRunId).toBeUndefined();
    });

    it('returns { allowed: false, conflictingRunId } when a RUNNING run exists on the branch', async () => {
      const runningRunId = 'run-abc-123';

      (mockPrisma.$transaction as jest.Mock).mockImplementation(async (fn: Function) =>
        fn(mockPrisma),
      );
      (mockPrisma.$queryRaw as jest.Mock).mockResolvedValue([
        { id: runningRunId, status: 'RUNNING', branch_name: 'feature/my-branch' },
      ]);

      const result = await checkAndLockBranch('feature/my-branch');

      expect(result.allowed).toBe(false);
      expect((result as any).conflictingRunId).toBe(runningRunId);
    });

    it('returns { allowed: false } when a PENDING run exists on the branch', async () => {
      const pendingRunId = 'run-def-456';

      (mockPrisma.$transaction as jest.Mock).mockImplementation(async (fn: Function) =>
        fn(mockPrisma),
      );
      (mockPrisma.$queryRaw as jest.Mock).mockResolvedValue([
        { id: pendingRunId, status: 'PENDING', branch_name: 'feature/my-branch' },
      ]);

      const result = await checkAndLockBranch('feature/my-branch');

      expect(result.allowed).toBe(false);
    });

    it('returns { allowed: true } when only FAILED/SUCCESS/CANCELLED runs exist on the branch', async () => {
      // Terminal-state runs do NOT block new runs
      (mockPrisma.$transaction as jest.Mock).mockImplementation(async (fn: Function) =>
        fn(mockPrisma),
      );
      // The query filters to only PENDING/RUNNING rows — so the result is empty
      (mockPrisma.$queryRaw as jest.Mock).mockResolvedValue([]);

      const result = await checkAndLockBranch('feature/my-branch');

      expect(result.allowed).toBe(true);
    });

    it('wraps the concurrency check inside a prisma.$transaction call', async () => {
      (mockPrisma.$transaction as jest.Mock).mockImplementation(async (fn: Function) =>
        fn(mockPrisma),
      );
      (mockPrisma.$queryRaw as jest.Mock).mockResolvedValue([]);

      await checkAndLockBranch('feature/my-branch');

      expect(mockPrisma.$transaction).toHaveBeenCalled();
    });

    it('uses a raw SQL query to perform the SELECT ... FOR UPDATE SKIP LOCKED check', async () => {
      (mockPrisma.$transaction as jest.Mock).mockImplementation(async (fn: Function) =>
        fn(mockPrisma),
      );
      (mockPrisma.$queryRaw as jest.Mock).mockResolvedValue([]);

      await checkAndLockBranch('feature/my-branch');

      expect(mockPrisma.$queryRaw).toHaveBeenCalled();
    });
  });

  // ---------------------------------------------------------------------------
  // createRun
  // ---------------------------------------------------------------------------

  describe('createRun', () => {
    it('persists a ValidationRun with status PENDING and all provided fields', async () => {
      const input = {
        branchName: 'feature/my-branch',
        commitSha: 'abc1234567890',
        userId: 'user-001',
        config: { stages: ['SETUP', 'LINT', 'TEST'] },
      };
      const persistedRun = {
        id: 'run-new-001',
        ...input,
        status: 'PENDING',
        createdAt: new Date(),
        updatedAt: new Date(),
        completedAt: null,
        cancelledAt: null,
      };

      (mockPrisma.validationRun.create as jest.Mock).mockResolvedValue(persistedRun);

      const result = await createRun(input);

      expect(result.status).toBe('PENDING');
      expect(result.branchName).toBe(input.branchName);
      expect(result.commitSha).toBe(input.commitSha);
      expect(result.userId).toBe(input.userId);
      expect(result.config).toEqual(input.config);
    });

    it('calls prisma.validationRun.create with data containing status PENDING', async () => {
      const input = {
        branchName: 'main',
        commitSha: 'deadbeef',
        userId: 'user-002',
        config: {},
      };
      const persistedRun = {
        id: 'run-new-002',
        ...input,
        status: 'PENDING',
        createdAt: new Date(),
        updatedAt: new Date(),
        completedAt: null,
        cancelledAt: null,
      };

      (mockPrisma.validationRun.create as jest.Mock).mockResolvedValue(persistedRun);

      await createRun(input);

      expect(mockPrisma.validationRun.create).toHaveBeenCalledWith(
        expect.objectContaining({
          data: expect.objectContaining({ status: 'PENDING' }),
        }),
      );
    });

    it('returns the newly created run record with an id', async () => {
      const input = {
        branchName: 'feature/new-feature',
        commitSha: 'cafebabe',
        userId: 'user-003',
        config: { stages: ['BUILD'] },
      };
      const persistedRun = {
        id: 'run-new-003',
        ...input,
        status: 'PENDING',
        createdAt: new Date(),
        updatedAt: new Date(),
        completedAt: null,
        cancelledAt: null,
      };

      (mockPrisma.validationRun.create as jest.Mock).mockResolvedValue(persistedRun);

      const result = await createRun(input);

      expect(result.id).toBeDefined();
      expect(typeof result.id).toBe('string');
    });
  });

  // ---------------------------------------------------------------------------
  // getRunWithStages
  // ---------------------------------------------------------------------------

  describe('getRunWithStages', () => {
    it('returns run with stages sorted by stageIndex ascending', async () => {
      const runId = 'run-with-stages-001';
      // Stages deliberately out of order to confirm sorting
      const unsortedStages = [
        { id: 'stage-c', stageIndex: 2, stageName: 'TEST', status: 'PENDING', logs: '', startedAt: null, completedAt: null },
        { id: 'stage-a', stageIndex: 0, stageName: 'SETUP', status: 'PENDING', logs: '', startedAt: null, completedAt: null },
        { id: 'stage-b', stageIndex: 1, stageName: 'LINT', status: 'PENDING', logs: '', startedAt: null, completedAt: null },
      ];
      const run = {
        id: runId,
        branchName: 'feature/sorted-stages',
        commitSha: 'abc',
        status: 'RUNNING',
        config: {},
        userId: 'user-001',
        createdAt: new Date(),
        updatedAt: new Date(),
        completedAt: null,
        cancelledAt: null,
        stages: unsortedStages,
      };

      (mockPrisma.validationRun.findUnique as jest.Mock).mockResolvedValue(run);

      const result = await getRunWithStages(runId);

      expect(result).not.toBeNull();
      expect(result!.stages).toHaveLength(3);
      expect(result!.stages[0].stageIndex).toBe(0);
      expect(result!.stages[1].stageIndex).toBe(1);
      expect(result!.stages[2].stageIndex).toBe(2);
    });

    it('returns null when the run does not exist', async () => {
      (mockPrisma.validationRun.findUnique as jest.Mock).mockResolvedValue(null);

      const result = await getRunWithStages('non-existent-run-id');

      expect(result).toBeNull();
    });

    it('returns run with an empty stages array when the run has no stages', async () => {
      const runId = 'run-no-stages';
      const run = {
        id: runId,
        branchName: 'feature/no-stages',
        commitSha: 'def',
        status: 'PENDING',
        config: {},
        userId: 'user-001',
        createdAt: new Date(),
        updatedAt: new Date(),
        completedAt: null,
        cancelledAt: null,
        stages: [],
      };

      (mockPrisma.validationRun.findUnique as jest.Mock).mockResolvedValue(run);

      const result = await getRunWithStages(runId);

      expect(result).not.toBeNull();
      expect(result!.stages).toHaveLength(0);
    });

    it('queries using the provided run id', async () => {
      const runId = 'run-query-check';

      (mockPrisma.validationRun.findUnique as jest.Mock).mockResolvedValue(null);

      await getRunWithStages(runId);

      expect(mockPrisma.validationRun.findUnique).toHaveBeenCalledWith(
        expect.objectContaining({
          where: expect.objectContaining({ id: runId }),
        }),
      );
    });

    it('includes stages relation in the query', async () => {
      const runId = 'run-include-stages';

      (mockPrisma.validationRun.findUnique as jest.Mock).mockResolvedValue(null);

      await getRunWithStages(runId);

      expect(mockPrisma.validationRun.findUnique).toHaveBeenCalledWith(
        expect.objectContaining({
          include: expect.objectContaining({ stages: expect.anything() }),
        }),
      );
    });
  });
});
