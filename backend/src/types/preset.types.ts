export const PRESET_MAX_PER_USER = 5;

export interface CreatePresetInput {
  userId: string;
  name: string;
  branchName: string;
  commitSha: string | null;
  envVars: Record<string, string>;
  featureFlags: Record<string, boolean>;
}

export interface UpdatePresetInput {
  userId?: string;
  name?: string;
  branchName?: string;
  commitSha?: string | null;
  envVars?: Record<string, string>;
  featureFlags?: Record<string, boolean>;
}

export interface ValidationPresetRecord {
  id: string;
  userId: string;
  name: string;
  branchName: string;
  commitSha: string | null;
  envVars: Record<string, string>;
  featureFlags: Record<string, boolean>;
  createdAt: Date;
  updatedAt: Date;
}
