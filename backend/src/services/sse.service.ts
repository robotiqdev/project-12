// SSE (Server-Sent Events) service for emitting real-time pipeline events
// to connected clients.

export async function emit(
  runId: string,
  event: string,
  data?: unknown,
): Promise<void> {
  // Stub: in production this would push to connected SSE clients
  void runId;
  void event;
  void data;
}
