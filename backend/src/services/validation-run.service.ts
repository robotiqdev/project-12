import * as concurrencyService from './concurrency.service';
import * as validationRunModel from '../models/validation-run.model';
import { CreateValidationRunInput, ValidationRunRecord } from '../types/validation-run.types';

export async function createRun(input: CreateValidationRunInput): Promise<ValidationRunRecord> {
  const lockResult = await concurrencyService.checkAndLockBranch(input.branchName);

  if (!lockResult.allowed) {
    throw Object.assign(new Error('BRANCH_LOCKED'), {
      code: 'BRANCH_LOCKED',
      conflictingRunId: lockResult.conflictingRunId,
    });
  }

  return validationRunModel.create(input);
}

export async function getRun(id: string): Promise<ValidationRunRecord | null> {
  return validationRunModel.findById(id);
}
