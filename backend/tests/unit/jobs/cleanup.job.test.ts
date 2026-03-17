/**
 * Unit tests for cleanup.job.ts (TASK-4618)
 *
 * Tests verify:
 * 1. Cron job is registered with schedule '0 2 * * *' (daily at 02:00 UTC)
 * 2. When cron fires, both cleanupService.runAgeBased() and cleanupService.runCountBased() are called
 * 3. Errors from cleanup services are caught and logged (job does not crash)
 */

// ---- Mocks ----

jest.mock('node-cron');
jest.mock('../../../src/services/cleanup.service');

import cron from 'node-cron';
import * as cleanupService from '../../../src/services/cleanup.service';
import { startCleanupJob } from '../../../src/jobs/cleanup.job';

// ---- Test setup ----

const mockCronSchedule = cron.schedule as jest.Mock;
const mockRunAgeBased = cleanupService.runAgeBased as jest.Mock;
const mockRunCountBased = cleanupService.runCountBased as jest.Mock;

describe('CleanupJob', () => {
  let cronCallback: () => Promise<void>;

  beforeEach(() => {
    jest.clearAllMocks();

    // Capture the callback passed to cron.schedule so we can invoke it manually
    mockCronSchedule.mockImplementation(
      (_schedule: string, callback: () => Promise<void>) => {
        cronCallback = callback;
        return { destroy: jest.fn() };
      },
    );

    mockRunAgeBased.mockResolvedValue({
      jobId: 'age-job-1',
      deletedCount: 5,
      duration: 100,
    });
    mockRunCountBased.mockResolvedValue({
      jobId: 'count-job-1',
      deletedCount: 3,
      duration: 80,
    });
  });

  // ---- Cron schedule registration ----

  it('registers a cron job with schedule "0 2 * * *" (daily at 02:00 UTC)', () => {
    startCleanupJob();

    expect(mockCronSchedule).toHaveBeenCalledWith(
      '0 2 * * *',
      expect.any(Function),
    );
  });

  it('calls cron.schedule exactly once when startCleanupJob is invoked', () => {
    startCleanupJob();

    expect(mockCronSchedule).toHaveBeenCalledTimes(1);
  });

  // ---- Cleanup service calls when cron fires ----

  it('calls cleanupService.runAgeBased() when the cron callback fires', async () => {
    startCleanupJob();

    await cronCallback();

    expect(mockRunAgeBased).toHaveBeenCalledTimes(1);
  });

  it('calls cleanupService.runCountBased() when the cron callback fires', async () => {
    startCleanupJob();

    await cronCallback();

    expect(mockRunCountBased).toHaveBeenCalledTimes(1);
  });

  it('calls both runAgeBased and runCountBased on each cron execution', async () => {
    startCleanupJob();

    await cronCallback();

    expect(mockRunAgeBased).toHaveBeenCalledTimes(1);
    expect(mockRunCountBased).toHaveBeenCalledTimes(1);
  });

  it('calls runAgeBased with no arguments', async () => {
    startCleanupJob();

    await cronCallback();

    expect(mockRunAgeBased).toHaveBeenCalledWith();
  });

  it('calls runCountBased with no arguments', async () => {
    startCleanupJob();

    await cronCallback();

    expect(mockRunCountBased).toHaveBeenCalledWith();
  });

  // ---- Error handling: job does not crash ----

  it('does not throw when cleanupService.runAgeBased() rejects', async () => {
    startCleanupJob();

    mockRunAgeBased.mockRejectedValue(new Error('Age-based cleanup failed'));

    await expect(cronCallback()).resolves.not.toThrow();
  });

  it('does not throw when cleanupService.runCountBased() rejects', async () => {
    startCleanupJob();

    mockRunCountBased.mockRejectedValue(new Error('Count-based cleanup failed'));

    await expect(cronCallback()).resolves.not.toThrow();
  });

  it('cron callback resolves (does not reject) when runAgeBased throws', async () => {
    startCleanupJob();

    mockRunAgeBased.mockRejectedValue(new Error('DB connection error'));

    await expect(cronCallback()).resolves.toBeUndefined();
  });

  it('cron callback resolves (does not reject) when runCountBased throws', async () => {
    startCleanupJob();

    mockRunCountBased.mockRejectedValue(new Error('Count threshold error'));

    await expect(cronCallback()).resolves.toBeUndefined();
  });

  it('catches errors and does not propagate them to the cron scheduler', async () => {
    startCleanupJob();

    const error = new Error('Unexpected failure');
    mockRunAgeBased.mockRejectedValue(error);

    // The cron callback must not reject even if cleanup service throws
    let threw = false;
    try {
      await cronCallback();
    } catch (_) {
      threw = true;
    }

    expect(threw).toBe(false);
  });

  it('continues gracefully even when both services fail', async () => {
    startCleanupJob();

    mockRunAgeBased.mockRejectedValue(new Error('Age-based error'));
    mockRunCountBased.mockRejectedValue(new Error('Count-based error'));

    await expect(cronCallback()).resolves.toBeUndefined();
  });
});
