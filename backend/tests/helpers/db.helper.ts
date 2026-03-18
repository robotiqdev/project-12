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
  const branches = options.branches ?? ["main"];
  const statuses = options.statuses ?? ["PENDING", "SUCCESS", "FAILED"];
  const userId = options.userId ?? "test_user";

  const data = Array.from({ length: count }, (_, i) => ({
    id: `seed-${Date.now()}-${i}`,
    branchName: branches[i % branches.length],
    commitSha: `sha${i.toString().padStart(6, "0")}`,
    status: statuses[i % statuses.length] as never,
    config: {},
    userId,
  }));

  await prisma.validationRun.createMany({ data });
}
