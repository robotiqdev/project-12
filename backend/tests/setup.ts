import { execSync } from "child_process";

export default async function globalSetup(): Promise<void> {
  // Ensure DATABASE_URL is set for the test database connection
  if (!process.env.DATABASE_URL) {
    throw new Error("DATABASE_URL environment variable must be set for integration tests");
  }

  // Run database migrations before tests
  execSync("npx prisma migrate deploy", {
    env: { ...process.env, DATABASE_URL: process.env.DATABASE_URL },
    stdio: "inherit",
  });
}
