import { prisma } from '../db/client';
import { RunStageRecord } from '../types/validation-run.types';

export async function createAll(
  runId: string,
  stages: Array<{ stageName: string; stageIndex: number }>,
): Promise<RunStageRecord[]> {
  await prisma.runStage.createMany({
    data: stages.map((s) => ({
      runId,
      stageName: s.stageName as never,
      stageIndex: s.stageIndex,
    })),
  });
  return findByRunId(runId);
}

export async function updateStage(
  id: string,
  update: Partial<{ status: string; logs: string; startedAt: Date; completedAt: Date }>,
) {
  return prisma.runStage.update({
    where: { id },
    data: update as never,
  });
}

export async function findByRunId(runId: string): Promise<RunStageRecord[]> {
  return prisma.runStage.findMany({
    where: { runId },
    orderBy: { stageIndex: 'asc' },
  }) as unknown as RunStageRecord[];
}
