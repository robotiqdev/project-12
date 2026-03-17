import { prisma } from '../db/client';
import { CreateValidationRunInput, ValidationRunRecord } from '../types/validation-run.types';

interface LockResult {
  allowed: boolean;
  conflictingRunId?: string;
}

interface ActiveRunRow {
  id: string;
  status: string;
  branch_name: string;
}

export async function checkAndLockBranch(branchName: string): Promise<LockResult> {
  return prisma.$transaction(async (tx) => {
    const rows = await (tx as typeof prisma).$queryRaw<ActiveRunRow[]>`
      SELECT id, status, branch_name
      FROM validation_runs
      WHERE branch_name = ${branchName}
        AND status IN ('PENDING', 'RUNNING')
      FOR UPDATE SKIP LOCKED
    `;

    if (rows.length === 0) {
      return { allowed: true };
    }

    const conflicting = rows[0];
    return {
      allowed: false,
      conflictingRunId: conflicting.id,
    };
  });
}

export async function createRun(input: CreateValidationRunInput): Promise<ValidationRunRecord> {
  return prisma.validationRun.create({
    data: {
      branchName: input.branchName,
      commitSha: input.commitSha,
      userId: input.userId,
      config: input.config as object,
      status: 'PENDING',
    },
  }) as unknown as ValidationRunRecord;
}

export async function getRunWithStages(id: string) {
  const run = await prisma.validationRun.findUnique({
    where: { id },
    include: { stages: true },
  });

  if (!run) return null;

  return {
    ...run,
    stages: [...run.stages].sort((a, b) => a.stageIndex - b.stageIndex),
  };
}
