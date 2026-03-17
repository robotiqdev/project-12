export enum ValidationRunStatus {
  PENDING = 'PENDING',
  RUNNING = 'RUNNING',
  SUCCESS = 'SUCCESS',
  FAILED = 'FAILED',
  CANCELLED = 'CANCELLED',
}

export interface ValidationRunRecord {
  id: string;
  branchName: string;
  commitSha: string;
  status: ValidationRunStatus;
  config: unknown;
  userId: string;
  createdAt: Date;
  updatedAt: Date;
  completedAt: Date | null;
  cancelledAt: Date | null;
}
