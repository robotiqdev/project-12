import { PrismaClient } from "@prisma/client";

const prisma = new PrismaClient();

export interface SeedRunsOptions {
  branches?: string[];
  statuses?: string[];
  timestamps?: Date[];
  opts?: Record<string, unknown>;
}

export async function seedRuns(
  count: number,
  options: SeedRunsOptions = {}
): Promise<void> {
  const branches = options.branches ?? ["main"];
  const statuses = options.statuses ?? ["PENDING", "RUNNING", "SUCCESS", "FAILED", "CANCELLED"];

  const data = Array.from({ length: count }, (_, i) => ({
    id: crypto.randomUUID(),
    branchName: branches[i % branches.length],
    commitSha: `sha${i.toString().padStart(6, "0")}`,
    status: statuses[i % statuses.length] as never,
    config: {},
    userId: `seed_user_${i % 10}`,
    createdAt: options.timestamps?.[i] ?? new Date(Date.now() - i * 60_000),
    updatedAt: options.timestamps?.[i] ?? new Date(Date.now() - i * 60_000),
  }));

  await (prisma.validationRun as { createMany: Function }).createMany({ data });
}

export { prisma };
