import { prisma } from '../db/client';
import {
  CreateValidationRunInput,
  ValidationRunRecord,
  RunListFilters,
} from '../types/validation-run.types';

export async function create(input: CreateValidationRunInput): Promise<ValidationRunRecord> {
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

export async function findById(id: string): Promise<ValidationRunRecord | null> {
  return prisma.validationRun.findUnique({
    where: { id },
  }) as unknown as ValidationRunRecord | null;
}

export async function findByIdWithStages(id: string) {
  return prisma.validationRun.findUnique({
    where: { id },
    include: { stages: true },
  });
}

export async function findActiveOnBranch(branchName: string) {
  return prisma.validationRun.findFirst({
    where: {
      branchName,
      status: { in: ['PENDING', 'RUNNING'] },
    },
  });
}

export async function updateStatus(
  id: string,
  status: string,
  extra?: { completedAt?: Date; cancelledAt?: Date },
) {
  return prisma.validationRun.update({
    where: { id },
    data: {
      status: status as never,
      ...(extra?.completedAt && { completedAt: extra.completedAt }),
      ...(extra?.cancelledAt && { cancelledAt: extra.cancelledAt }),
    },
  });
}

export async function list(filters: RunListFilters): Promise<{ data: ValidationRunRecord[]; total: number }> {
  const { branchName, userId, status, page = 1, pageSize = 20 } = filters;
  const where = {
    ...(branchName && { branchName }),
    ...(userId && { userId }),
    ...(status && { status: status as never }),
  };
  const [data, total] = await Promise.all([
    prisma.validationRun.findMany({
      where,
      skip: (page - 1) * pageSize,
      take: pageSize,
      orderBy: { createdAt: 'desc' },
    }),
    prisma.validationRun.count({ where }),
  ]);
  return { data: data as unknown as ValidationRunRecord[], total };
}
