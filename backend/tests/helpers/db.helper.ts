import { PrismaClient } from "@prisma/client";

const prisma = new PrismaClient();

export interface SeedRunsOptions {
  branches?: string[];
  statuses?: string[];
  userId?: string;
}

export async function seedRuns(
  count: number,
  options: SeedRunsOptions = {}
): Promise<void> {
  const {
    branches = ["main"],
    statuses = ["PENDING", "RUNNING", "SUCCESS", "FAILED", "CANCELLED"],
    userId = "test_user",
  } = options;

  const data = Array.from({ length: count }, (_, i) => ({
    id: `seed-${Date.now()}-${i}`,
    branchName: branches[i % branches.length],
    commitSha: `sha${i.toString().padStart(8, "0")}`,
    status: statuses[i % statuses.length] as never,
    config: {},
    userId: `${userId}_${i % 10}`,
    createdAt: new Date(Date.now() - i * 60_000),
    updatedAt: new Date(Date.now() - i * 60_000),
  }));

  await prisma.validationRun.createMany({ data });
}

export default prisma;
