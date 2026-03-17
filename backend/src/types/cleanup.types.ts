export enum CleanupJobType {
  AGE_BASED = 'AGE_BASED',
  COUNT_BASED = 'COUNT_BASED',
}

export enum CleanupJobStatus {
  PENDING = 'PENDING',
  RUNNING = 'RUNNING',
  DONE = 'DONE',
  FAILED = 'FAILED',
}

export interface CleanupJobRecord {
  id: string;
  type: CleanupJobType;
  status: CleanupJobStatus;
  deletedCount: number;
  startedAt: Date;
  completedAt: Date | null;
  errorMessage: string | null;
}

export interface CleanupResult {
  jobId: string;
  type: CleanupJobType;
  deletedCount: number;
  duration: number;
}
