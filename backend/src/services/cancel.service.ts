import { prisma } from '../db/client';
import { ValidationRunRecord, ValidationRunStatus } from '../models/validation-run.types';
import { ValidationRunModel } from '../models/validation-run.model';

// ---------------------------------------------------------------------------
// Errors
// ---------------------------------------------------------------------------

export class RunNotCancellableError extends Error {
  constructor() {
    super('RUN_NOT_CANCELLABLE');
    this.name = 'RunNotCancellableError';
  }
}

export class RunNotFoundError extends Error {
  constructor(id: string) {
    super(`Validation run not found: ${id}`);
    this.name = 'RunNotFoundError';
  }
}

// ---------------------------------------------------------------------------
// Pipeline service stub — provides cancellation token for RUNNING runs.
// The token may not exist if the pipeline hasn't started yet (async scheduling).
// ---------------------------------------------------------------------------

interface CancellationToken {
  cancel(): void;
}

export const pipelineService = {
  _tokens: new Map<string, CancellationToken>(),

  getToken(id: string): CancellationToken | undefined {
    return this._tokens.get(id);
  },

  registerToken(id: string, token: CancellationToken): void {
    this._tokens.set(id, token);
  },
};

// ---------------------------------------------------------------------------
// Cancel service
// ---------------------------------------------------------------------------

const validationRunModel = new ValidationRunModel();

export const cancelService = {
  async cancelRun(id: string): Promise<ValidationRunRecord> {
    // Fetch run to validate its current status
    const run = await prisma.validationRun.findUnique({ where: { id } });

    if (!run) {
      throw new RunNotFoundError(id);
    }

    // Only PENDING and RUNNING runs are cancellable
    if (
      run.status !== ValidationRunStatus.PENDING &&
      run.status !== ValidationRunStatus.RUNNING
    ) {
      throw new RunNotCancellableError();
    }

    // For RUNNING runs, signal the pipeline via cancellation token if it exists.
    // The token may not exist yet due to async scheduling — handle gracefully.
    if (run.status === ValidationRunStatus.RUNNING) {
      const token = pipelineService.getToken(id);
      if (token !== null && token !== undefined) {
        token.cancel();
      }
    }

    // Update the DB: set status=CANCELLED and cancelledAt=now
    return validationRunModel.cancel(id);
  },
};
