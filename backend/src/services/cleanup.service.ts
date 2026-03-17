import { prisma } from '../db/client';
import * as cleanupJobModel from '../models/cleanup-job.model';
import { CleanupJobType, CleanupResult } from '../types/cleanup.types';

export async function runAgeBased(): Promise<CleanupResult> {
  const job = await cleanupJobModel.create(CleanupJobType.AGE_BASED);
  await cleanupJobModel.start(job.id);

  const startTime = Date.now();
  const cutoff = new Date(Date.now() - 90 * 24 * 60 * 60 * 1000);

  try {
    const result = await prisma.validationRun.deleteMany({
      where: {
        createdAt: { lt: cutoff },
        status: { not: 'RUNNING' },
      },
    });

    const deletedCount = result.count;
    await cleanupJobModel.complete(job.id, deletedCount);

    return {
      jobId: job.id,
      type: CleanupJobType.AGE_BASED,
      deletedCount,
      duration: Date.now() - startTime,
    };
  } catch (error) {
    await cleanupJobModel.fail(job.id, (error as Error).message);
    throw error;
  }
}
