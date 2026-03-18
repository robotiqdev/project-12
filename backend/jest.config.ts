import type { Config } from "jest";

const config: Config = {
  preset: "ts-jest",
  globalSetup: "./tests/setup.ts",
  testEnvironment: "node",
  transform: {
    "^.+\\.tsx?$": ["ts-jest", {}],
  },
  testMatch: ["**/*.test.ts"],
  moduleFileExtensions: ["ts", "tsx", "js", "jsx", "json"],
};

export default config;
