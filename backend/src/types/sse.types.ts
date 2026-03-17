// SSE (Server-Sent Events) types for real-time pipeline event streaming — TASK-4596

export type SSEEventType =
  | 'connected'
  | 'stage_started'
  | 'stage_completed'
  | 'log_line'
  | 'run_completed';

export interface SSEEvent {
  id?: string;
  event: SSEEventType;
  data: unknown;
}
