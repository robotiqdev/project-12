import { PrismaClient } from '@prisma/client';
import { CleanupJobType, CleanupJobStatus } from '../types/cleanup.types';

const prisma = new PrismaClient();

export async function create(type: CleanupJobType) {
  return prisma.cleanupJob.create({
    data: {
      type,
      status: CleanupJobStatus.PENDING,
    },
  });
}

export async function start(id: string) {
  return prisma.cleanupJob.update({
    where: { id },
    data: {
      status: CleanupJobStatus.RUNNING,
    },
  });
}

export async function complete(id: string, deletedCount: number) {
  return prisma.cleanupJob.update({
    where: { id },
    data: {
      status: CleanupJobStatus.DONE,
      deletedCount,
      completedAt: new Date(),
    },
  });
}

export async function fail(id: string, errorMessage: string) {
  return prisma.cleanupJob.update({
    where: { id },
    data: {
      status: CleanupJobStatus.FAILED,
      errorMessage,
      completedAt: new Date(),
    },
  });
}

export async function list(limit = 20) {
  return prisma.cleanupJob.findMany({
    take: limit,
    orderBy: {
      startedAt: 'desc',
    },
  });
}
