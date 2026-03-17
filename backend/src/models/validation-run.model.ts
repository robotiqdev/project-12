import { ValidationRunRecord, ValidationRunStatus } from './validation-run.types';

interface UpdateStatusOptions {
  completedAt?: Date | null;
  cancelledAt?: Date | null;
}

export class ValidationRunModel {
  async updateStatus(
    id: string,
    status: ValidationRunStatus,
    options: UpdateStatusOptions = {},
  ): Promise<ValidationRunRecord> {
    const { completedAt, cancelledAt } = options;
    // Implementation delegated to Prisma client
    throw new Error('Not implemented');
  }

  async cancel(id: string): Promise<ValidationRunRecord> {
    return this.updateStatus(id, ValidationRunStatus.CANCELLED, {
      cancelledAt: new Date(),
    });
  }
}
