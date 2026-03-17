export enum ValidationRunStatus {
  PENDING = 'PENDING',
  RUNNING = 'RUNNING',
  SUCCESS = 'SUCCESS',
  FAILED = 'FAILED',
  CANCELLED = 'CANCELLED',
}

export enum StageType {
  SETUP = 'SETUP',
  LINT = 'LINT',
  TEST = 'TEST',
  BUILD = 'BUILD',
  DEPLOY = 'DEPLOY',
}

export enum StageStatus {
  PENDING = 'PENDING',
  RUNNING = 'RUNNING',
  SUCCESS = 'SUCCESS',
  FAILED = 'FAILED',
  CANCELLED = 'CANCELLED',
}

export const STAGE_ORDER: StageType[] = [
  StageType.SETUP,
  StageType.LINT,
  StageType.TEST,
  StageType.BUILD,
  StageType.DEPLOY,
];

export interface RunConfig {
  stages?: string[];
  [key: string]: unknown;
}

export interface CreateValidationRunInput {
  branchName: string;
  commitSha: string;
  userId: string;
  config: RunConfig | Record<string, unknown>;
}

export interface RunListFilters {
  branchName?: string;
  userId?: string;
  status?: ValidationRunStatus;
  page?: number;
  pageSize?: number;
}

export interface ValidationRunRecord {
  id: string;
  branchName: string;
  commitSha: string;
  status: string;
  config: Record<string, unknown>;
  userId: string;
  createdAt: Date;
  updatedAt: Date;
  completedAt: Date | null;
  cancelledAt: Date | null;
}

export interface RunStageRecord {
  id: string;
  runId: string;
  stageName: string;
  stageIndex: number;
  status: string;
  logs: string;
  startedAt: Date | null;
  completedAt: Date | null;
}
