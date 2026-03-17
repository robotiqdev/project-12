import cron from 'node-cron';
import * as cleanupService from '../services/cleanup.service';

export function startCleanupJob(): void {
  if (process.env.NODE_ENV === 'test') {
    return;
  }

  cron.schedule('0 2 * * *', async () => {
    try {
      const ageResult = await cleanupService.runAgeBased();
      const countResult = await cleanupService.runCountBased();
      console.info('Cleanup complete', { ageResult, countResult });
    } catch (error) {
      console.error('Cleanup job failed', error);
    }
  });
}
