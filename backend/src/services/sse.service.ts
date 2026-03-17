// SSE (Server-Sent Events) service for emitting real-time pipeline events
// to connected clients — TASK-4596
//
// Stub implementation: subscribe, emit, unsubscribe are not yet implemented.
// Full implementation should:
//   - subscribe: set headers, send retry, send connected event, set up heartbeat, register close handler
//   - emit: serialize SSEEvent to wire format and write to all subscribers
//   - unsubscribe: remove client, delete key if empty, clear heartbeat interval

import { Response } from 'express';

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const clients = new Map<string, Set<Response>>();

export function subscribe(runId: string, res: Response): void {
  void runId;
  void res;
}

export async function emit(
  runId: string,
  event: string,
  data?: unknown,
): Promise<void> {
  void runId;
  void event;
  void data;
}

export function unsubscribe(runId: string, res: Response): void {
  void runId;
  void res;
}
