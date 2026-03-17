/**
 * Unit tests for sse.service.ts — TASK-4596
 *
 * Tests the public interface of:
 *   - subscribe(runId, res): sets SSE headers, sends retry + connected event,
 *     sets up heartbeat interval, registers close handler
 *   - emit(runId, event, data): formats event as SSE wire format and writes
 *     to all subscribers for the given runId
 *   - unsubscribe(runId, res): removes client from map; subsequent emit does
 *     not reach that client; deletes runId key if last subscriber
 *
 * The SSE service is a process-local singleton, so module state is reset
 * between tests via jest.resetModules() to isolate each test.
 */

import { Response } from 'express';

// ─── Helpers ──────────────────────────────────────────────────────────────────

type MockResponse = Response & {
  _triggerClose: () => void;
};

function makeMockResponse(): MockResponse {
  const listeners: Record<string, Array<() => void>> = {};

  const res = {
    write: jest.fn(),
    setHeader: jest.fn(),
    on: jest.fn((event: string, cb: () => void) => {
      if (!listeners[event]) listeners[event] = [];
      listeners[event].push(cb);
    }),
    _triggerClose: () => {
      if (listeners['close']) {
        listeners['close'].forEach((cb) => cb());
      }
    },
  } as unknown as MockResponse;

  return res;
}

// ─── Module reset helpers ─────────────────────────────────────────────────────

// We use dynamic require to reset singleton state between tests.
// The type mirrors the public API described in TASK-4596.
type SSEServiceModule = {
  subscribe: (runId: string, res: Response) => void;
  emit: (runId: string, event: string, data?: unknown) => Promise<void>;
  unsubscribe: (runId: string, res: Response) => void;
};

// ─── Tests ───────────────────────────────────────────────────────────────────

describe('SSEService', () => {
  let sseService: SSEServiceModule;

  beforeEach(() => {
    jest.resetModules();
    jest.useFakeTimers();
    // eslint-disable-next-line @typescript-eslint/no-var-requires
    sseService = require('../../../src/services/sse.service') as SSEServiceModule;
  });

  afterEach(() => {
    jest.useRealTimers();
    jest.clearAllMocks();
  });

  // ─────────────────────────────────────────────────────────────────────────
  // subscribe — header setup
  // ─────────────────────────────────────────────────────────────────────────

  describe('subscribe — header setup', () => {
    it('sets Content-Type: text/event-stream header', () => {
      const res = makeMockResponse();
      sseService.subscribe('run-001', res);
      expect(res.setHeader).toHaveBeenCalledWith('Content-Type', 'text/event-stream');
    });

    it('sets Cache-Control: no-cache header', () => {
      const res = makeMockResponse();
      sseService.subscribe('run-001', res);
      expect(res.setHeader).toHaveBeenCalledWith('Cache-Control', 'no-cache');
    });

    it('sets Connection: keep-alive header', () => {
      const res = makeMockResponse();
      sseService.subscribe('run-001', res);
      expect(res.setHeader).toHaveBeenCalledWith('Connection', 'keep-alive');
    });
  });

  // ─────────────────────────────────────────────────────────────────────────
  // subscribe — initial messages
  // ─────────────────────────────────────────────────────────────────────────

  describe('subscribe — initial messages', () => {
    it('sends retry directive after subscribing', () => {
      const res = makeMockResponse();
      sseService.subscribe('run-002', res);
      const allWrites = (res.write as jest.Mock).mock.calls
        .map((c: [string]) => c[0])
        .join('');
      expect(allWrites).toContain('retry:');
    });

    it('sends initial connected event on subscribe', () => {
      const res = makeMockResponse();
      sseService.subscribe('run-003', res);
      const allWrites = (res.write as jest.Mock).mock.calls
        .map((c: [string]) => c[0])
        .join('');
      expect(allWrites).toContain('connected');
    });

    it('sends connected event with data field', () => {
      const res = makeMockResponse();
      sseService.subscribe('run-004', res);
      const allWrites = (res.write as jest.Mock).mock.calls
        .map((c: [string]) => c[0])
        .join('');
      // SSE connected event: event: connected\ndata: {}\n\n
      expect(allWrites).toContain('event:');
      expect(allWrites).toContain('data:');
    });
  });

  // ─────────────────────────────────────────────────────────────────────────
  // subscribe — close handler and heartbeat
  // ─────────────────────────────────────────────────────────────────────────

  describe('subscribe — close handler and heartbeat', () => {
    it('registers a close event listener on the response', () => {
      const res = makeMockResponse();
      sseService.subscribe('run-005', res);
      expect(res.on).toHaveBeenCalledWith('close', expect.any(Function));
    });

    it('sends heartbeat comment after 15 seconds', () => {
      const res = makeMockResponse();
      sseService.subscribe('run-heartbeat', res);
      (res.write as jest.Mock).mockClear();
      jest.advanceTimersByTime(15000);
      const allWrites = (res.write as jest.Mock).mock.calls
        .map((c: [string]) => c[0])
        .join('');
      expect(allWrites).toContain(':heartbeat');
    });

    it('sends heartbeat comment repeatedly every 15 seconds', () => {
      const res = makeMockResponse();
      sseService.subscribe('run-hb-repeat', res);
      (res.write as jest.Mock).mockClear();
      jest.advanceTimersByTime(45000);
      const heartbeatCalls = (res.write as jest.Mock).mock.calls.filter(
        (c: [string]) => c[0].includes(':heartbeat'),
      );
      expect(heartbeatCalls.length).toBeGreaterThanOrEqual(3);
    });

    it('calls unsubscribe automatically when connection closes', () => {
      const res = makeMockResponse();
      sseService.subscribe('run-close', res);
      res._triggerClose();
      // After close, emit should NOT reach this client
      (res.write as jest.Mock).mockClear();
      void sseService.emit('run-close', 'log_line', 'test');
      expect(res.write).not.toHaveBeenCalled();
    });
  });

  // ─────────────────────────────────────────────────────────────────────────
  // subscribe — clients map management
  // ─────────────────────────────────────────────────────────────────────────

  describe('subscribe — clients map management', () => {
    it('creates a new entry in the clients map for an unknown runId', () => {
      const res = makeMockResponse();
      sseService.subscribe('unknown-run-001', res);
      // Verify the subscriber receives subsequent emits
      (res.write as jest.Mock).mockClear();
      void sseService.emit('unknown-run-001', 'log_line', 'test payload');
      expect(res.write).toHaveBeenCalled();
    });

    it('adds multiple subscribers to the same runId entry', () => {
      const res1 = makeMockResponse();
      const res2 = makeMockResponse();
      sseService.subscribe('run-multi-sub', res1);
      sseService.subscribe('run-multi-sub', res2);
      (res1.write as jest.Mock).mockClear();
      (res2.write as jest.Mock).mockClear();
      void sseService.emit('run-multi-sub', 'log_line', 'broadcast');
      expect(res1.write).toHaveBeenCalled();
      expect(res2.write).toHaveBeenCalled();
    });
  });

  // ─────────────────────────────────────────────────────────────────────────
  // emit — SSE wire format
  // ─────────────────────────────────────────────────────────────────────────

  describe('emit — SSE wire format', () => {
    it('writes to all subscribers for the given runId', async () => {
      const res1 = makeMockResponse();
      const res2 = makeMockResponse();
      sseService.subscribe('run-emit-multi', res1);
      sseService.subscribe('run-emit-multi', res2);
      (res1.write as jest.Mock).mockClear();
      (res2.write as jest.Mock).mockClear();
      await sseService.emit('run-emit-multi', 'log_line', 'hello');
      expect(res1.write).toHaveBeenCalled();
      expect(res2.write).toHaveBeenCalled();
    });

    it('formats event with id: field in SSE wire format', async () => {
      const res = makeMockResponse();
      sseService.subscribe('run-format-id', res);
      (res.write as jest.Mock).mockClear();
      await sseService.emit('run-format-id', 'log_line', { message: 'hello' });
      const written = (res.write as jest.Mock).mock.calls[0][0] as string;
      expect(written).toContain('id:');
    });

    it('formats event with event: field in SSE wire format', async () => {
      const res = makeMockResponse();
      sseService.subscribe('run-format-event', res);
      (res.write as jest.Mock).mockClear();
      await sseService.emit('run-format-event', 'stage_started', {});
      const written = (res.write as jest.Mock).mock.calls[0][0] as string;
      expect(written).toContain('event:');
    });

    it('formats event with data: field in SSE wire format', async () => {
      const res = makeMockResponse();
      sseService.subscribe('run-format-data', res);
      (res.write as jest.Mock).mockClear();
      await sseService.emit('run-format-data', 'log_line', { line: 'test output' });
      const written = (res.write as jest.Mock).mock.calls[0][0] as string;
      expect(written).toContain('data:');
    });

    it('terminates each SSE event with a double newline', async () => {
      const res = makeMockResponse();
      sseService.subscribe('run-format-newline', res);
      (res.write as jest.Mock).mockClear();
      await sseService.emit('run-format-newline', 'log_line', 'test');
      const written = (res.write as jest.Mock).mock.calls[0][0] as string;
      expect(written).toMatch(/\n\n$/);
    });

    it('serializes data as JSON in the data: field', async () => {
      const res = makeMockResponse();
      sseService.subscribe('run-json-data', res);
      (res.write as jest.Mock).mockClear();
      const payload = { message: 'npm test passed', stage: 'TEST' };
      await sseService.emit('run-json-data', 'log_line', payload);
      const written = (res.write as jest.Mock).mock.calls[0][0] as string;
      expect(written).toContain(JSON.stringify(payload));
    });

    it('does not write to subscribers of a different runId', async () => {
      const resA = makeMockResponse();
      const resB = makeMockResponse();
      sseService.subscribe('run-A', resA);
      sseService.subscribe('run-B', resB);
      (resA.write as jest.Mock).mockClear();
      (resB.write as jest.Mock).mockClear();
      await sseService.emit('run-A', 'log_line', 'only for A');
      expect(resA.write).toHaveBeenCalled();
      expect(resB.write).not.toHaveBeenCalled();
    });

    it('does not throw when emitting to a runId with no subscribers', async () => {
      await expect(
        sseService.emit('no-subscribers-run', 'log_line', 'data'),
      ).resolves.not.toThrow();
    });

    it('writes the event name in the event: line', async () => {
      const res = makeMockResponse();
      sseService.subscribe('run-event-name', res);
      (res.write as jest.Mock).mockClear();
      await sseService.emit('run-event-name', 'stage_completed', { stage: 'LINT' });
      const written = (res.write as jest.Mock).mock.calls[0][0] as string;
      expect(written).toContain('stage_completed');
    });
  });

  // ─────────────────────────────────────────────────────────────────────────
  // unsubscribe
  // ─────────────────────────────────────────────────────────────────────────

  describe('unsubscribe', () => {
    it('removes the client so subsequent emit does not reach it', async () => {
      const res = makeMockResponse();
      sseService.subscribe('run-unsub-basic', res);
      sseService.unsubscribe('run-unsub-basic', res);
      (res.write as jest.Mock).mockClear();
      await sseService.emit('run-unsub-basic', 'log_line', 'should not arrive');
      expect(res.write).not.toHaveBeenCalled();
    });

    it('deletes the runId key when the last subscriber is removed', async () => {
      const res = makeMockResponse();
      sseService.subscribe('run-last-sub', res);
      sseService.unsubscribe('run-last-sub', res);
      // Emitting to a now-empty runId should not throw
      await expect(
        sseService.emit('run-last-sub', 'log_line', 'test'),
      ).resolves.not.toThrow();
    });

    it('does not remove other subscribers when one client unsubscribes', async () => {
      const res1 = makeMockResponse();
      const res2 = makeMockResponse();
      sseService.subscribe('run-partial-unsub', res1);
      sseService.subscribe('run-partial-unsub', res2);
      sseService.unsubscribe('run-partial-unsub', res1);
      (res1.write as jest.Mock).mockClear();
      (res2.write as jest.Mock).mockClear();
      await sseService.emit('run-partial-unsub', 'log_line', 'broadcast');
      expect(res1.write).not.toHaveBeenCalled();
      expect(res2.write).toHaveBeenCalled();
    });

    it('can be called for a runId with no subscribers without throwing', () => {
      expect(() => {
        sseService.unsubscribe('non-existent-run', {} as Response);
      }).not.toThrow();
    });

    it('can be called multiple times for the same client without throwing', () => {
      const res = makeMockResponse();
      sseService.subscribe('run-double-unsub', res);
      expect(() => {
        sseService.unsubscribe('run-double-unsub', res);
        sseService.unsubscribe('run-double-unsub', res);
      }).not.toThrow();
    });

    it('clears the heartbeat interval when unsubscribing', async () => {
      const res = makeMockResponse();
      sseService.subscribe('run-hb-cleanup', res);
      sseService.unsubscribe('run-hb-cleanup', res);
      (res.write as jest.Mock).mockClear();
      // Advance time past heartbeat interval — no more heartbeats should fire
      jest.advanceTimersByTime(30000);
      const heartbeatCalls = (res.write as jest.Mock).mock.calls.filter(
        (c: [string]) => c[0].includes(':heartbeat'),
      );
      expect(heartbeatCalls).toHaveLength(0);
    });
  });
});
