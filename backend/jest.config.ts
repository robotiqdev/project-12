import type { Config } from "jest";

const config: Config = {
  globalSetup: "./tests/setup.ts",
  testEnvironment: "node",
  transform: {
    "^.+\\.tsx?$": "ts-jest",
  },
  testMatch: ["**/*.test.ts"],
};

export default config;
