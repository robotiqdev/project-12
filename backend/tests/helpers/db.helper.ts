import { PrismaClient } from "@prisma/client";

const prisma = new PrismaClient();

export interface SeedRunsOptions {
  branches?: string[];
  statuses?: string[];
  timestamps?: Date[];
}

export async function seedRuns(
  count: number,
  options: SeedRunsOptions = {}
): Promise<void> {
  const {
    branches = ["main"],
    statuses = ["PENDING", "RUNNING", "SUCCESS", "FAILED", "CANCELLED"],
  } = options;

  const data = Array.from({ length: count }, (_, i) => ({
    id: require("crypto").randomUUID(),
    branchName: branches[i % branches.length],
    commitSha: `sha${i.toString().padStart(6, "0")}`,
    status: statuses[i % statuses.length] as any,
    config: { timeout: 300 },
    userId: `seed_user_${i % 10}`,
    createdAt: new Date(Date.now() - i * 60_000),
    updatedAt: new Date(Date.now() - i * 60_000),
  }));

  await prisma.validationRun.createMany({ data });
}

export default prisma;
