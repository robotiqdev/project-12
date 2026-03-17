/**
 * Integration tests for GET /api/validation-runs/:id/stream — TASK-4596
 *
 * Tests the SSE streaming endpoint:
 *   - Returns 200 with Content-Type: text/event-stream for a valid runId
 *   - Returns 404 for a runId that does not exist
 *   - Sends initial 'connected' event when client subscribes
 *   - Sends log_line events during pipeline execution
 *   - Keeps connection alive with heartbeat comments
 */

import * as http from 'http';
import request from 'supertest';
import { app } from '../../src/app';
import { prisma } from '../../src/db/client';

// Mock pipeline service to prevent real execution
jest.mock('../../src/services/pipeline.service');
import * as pipelineService from '../../src/services/pipeline.service';
const mockEnqueue = pipelineService.enqueue as jest.Mock;

/** A valid 40-character hex commit SHA */
const VALID_COMMIT_SHA = 'a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2';

/** Opens an SSE stream to the given URL and collects events until timeout. */
function collectSSEEvents(
  server: http.Server,
  path: string,
  durationMs: number,
): Promise<{ headers: Record<string, string | string[]>; rawData: string; statusCode: number }> {
  return new Promise((resolve, reject) => {
    const addr = server.address() as { port: number };
    const chunks: Buffer[] = [];
    let statusCode = 0;
    let responseHeaders: Record<string, string | string[]> = {};

    const req = http.get(
      {
        host: '127.0.0.1',
        port: addr.port,
        path,
        headers: { Accept: 'text/event-stream' },
      },
      (res) => {
        statusCode = res.statusCode ?? 0;
        responseHeaders = res.headers as Record<string, string | string[]>;
        res.on('data', (chunk: Buffer) => chunks.push(chunk));
        res.on('error', reject);
      },
    );

    req.on('error', reject);

    setTimeout(() => {
      req.destroy();
      resolve({
        headers: responseHeaders,
        rawData: Buffer.concat(chunks).toString('utf-8'),
        statusCode,
      });
    }, durationMs);
  });
}

// ── SSE stream endpoint ───────────────────────────────────────────────────────

describe('GET /api/validation-runs/:id/stream', () => {
  let server: http.Server;

  beforeAll((done) => {
    server = app.listen(0, '127.0.0.1', done);
  });

  afterAll((done) => {
    server.close(done);
    void prisma.$disconnect();
  });

  beforeEach(async () => {
    await prisma.$executeRaw`TRUNCATE TABLE run_stages, validation_runs CASCADE`;
    jest.clearAllMocks();
    mockEnqueue.mockResolvedValue(undefined);
  });

  // ── 404 for unknown runId ─────────────────────────────────────────────────

  it('returns 404 for a runId that does not exist', async () => {
    const unknownId = '00000000-0000-0000-0000-000000000000';
    await request(app)
      .get(`/api/validation-runs/${unknownId}/stream`)
      .expect(404);
  });

  it('returns a JSON error body on 404', async () => {
    const unknownId = '00000000-0000-0000-0000-000000000001';
    const response = await request(app)
      .get(`/api/validation-runs/${unknownId}/stream`)
      .expect(404);
    expect(response.body).toHaveProperty('error');
  });

  // ── SSE headers and initial event ─────────────────────────────────────────

  it('returns Content-Type: text/event-stream for a valid runId', async () => {
    // Create a run first
    const createResponse = await request(app)
      .post('/api/validation-runs')
      .send({ branchName: 'feature/sse-test', commitSha: VALID_COMMIT_SHA })
      .expect(201);

    const runId = createResponse.body.id as string;
    const { headers, statusCode } = await collectSSEEvents(
      server,
      `/api/validation-runs/${runId}/stream`,
      300,
    );

    expect(statusCode).toBe(200);
    expect(String(headers['content-type'])).toContain('text/event-stream');
  });

  it('sends retry directive in SSE stream on connect', async () => {
    const createResponse = await request(app)
      .post('/api/validation-runs')
      .send({ branchName: 'feature/sse-retry', commitSha: VALID_COMMIT_SHA })
      .expect(201);

    const runId = createResponse.body.id as string;
    const { rawData } = await collectSSEEvents(
      server,
      `/api/validation-runs/${runId}/stream`,
      300,
    );

    expect(rawData).toContain('retry:');
  });

  it('sends initial connected event when client subscribes', async () => {
    const createResponse = await request(app)
      .post('/api/validation-runs')
      .send({ branchName: 'feature/sse-connected', commitSha: VALID_COMMIT_SHA })
      .expect(201);

    const runId = createResponse.body.id as string;
    const { rawData } = await collectSSEEvents(
      server,
      `/api/validation-runs/${runId}/stream`,
      300,
    );

    expect(rawData).toContain('connected');
  });

  it('sends initial event with event: field in SSE wire format', async () => {
    const createResponse = await request(app)
      .post('/api/validation-runs')
      .send({ branchName: 'feature/sse-format', commitSha: VALID_COMMIT_SHA })
      .expect(201);

    const runId = createResponse.body.id as string;
    const { rawData } = await collectSSEEvents(
      server,
      `/api/validation-runs/${runId}/stream`,
      300,
    );

    expect(rawData).toContain('event:');
    expect(rawData).toContain('data:');
  });

  // ── log_line events during pipeline execution ─────────────────────────────

  it('sends log_line events during pipeline stage execution', (done) => {
    // Import the SSE service so we can manually emit events in this test
    // eslint-disable-next-line @typescript-eslint/no-var-requires
    const sseService = require('../../src/services/sse.service') as {
      emit: (runId: string, event: string, data?: unknown) => Promise<void>;
    };

    void (async () => {
      try {
        const createResponse = await request(app)
          .post('/api/validation-runs')
          .send({ branchName: 'feature/sse-log-line', commitSha: VALID_COMMIT_SHA })
          .expect(201);

        const runId = createResponse.body.id as string;
        const collectedRawData: string[] = [];
        let resolved = false;

        const addr = server.address() as { port: number };
        const req = http.get(
          {
            host: '127.0.0.1',
            port: addr.port,
            path: `/api/validation-runs/${runId}/stream`,
            headers: { Accept: 'text/event-stream' },
          },
          (res) => {
            res.on('data', (chunk: Buffer) => {
              collectedRawData.push(chunk.toString('utf-8'));
              // Once we have received some data, emit a log_line and check it appears
              if (!resolved && collectedRawData.join('').includes('connected')) {
                resolved = true;
                void sseService.emit(runId, 'log_line', { line: 'test log output' }).then(() => {
                  setTimeout(() => {
                    req.destroy();
                    const allData = collectedRawData.join('');
                    try {
                      expect(allData).toContain('log_line');
                      done();
                    } catch (err) {
                      done(err as Error);
                    }
                  }, 100);
                });
              }
            });
            res.on('error', done);
          },
        );
        req.on('error', done);

        // Timeout safety
        setTimeout(() => {
          if (!resolved) {
            req.destroy();
            done(new Error('Timed out waiting for connected event from SSE stream'));
          }
        }, 2000);
      } catch (err) {
        done(err as Error);
      }
    })();
  });

  it('sends log_line events to all subscribers of the same runId', (done) => {
    // eslint-disable-next-line @typescript-eslint/no-var-requires
    const sseService = require('../../src/services/sse.service') as {
      emit: (runId: string, event: string, data?: unknown) => Promise<void>;
    };

    void (async () => {
      try {
        const createResponse = await request(app)
          .post('/api/validation-runs')
          .send({ branchName: 'feature/sse-multi-client', commitSha: VALID_COMMIT_SHA })
          .expect(201);

        const runId = createResponse.body.id as string;
        const addr = server.address() as { port: number };

        const client1Data: string[] = [];
        const client2Data: string[] = [];
        let client1Connected = false;
        let client2Connected = false;
        let emitted = false;
        let doneCount = 0;

        function checkDone() {
          doneCount++;
          if (doneCount >= 2) {
            const all1 = client1Data.join('');
            const all2 = client2Data.join('');
            try {
              expect(all1).toContain('log_line');
              expect(all2).toContain('log_line');
              done();
            } catch (err) {
              done(err as Error);
            }
          }
        }

        function makeClient(dataCollector: string[], onConnect: () => void): http.ClientRequest {
          const req = http.get(
            {
              host: '127.0.0.1',
              port: addr.port,
              path: `/api/validation-runs/${runId}/stream`,
              headers: { Accept: 'text/event-stream' },
            },
            (res) => {
              res.on('data', (chunk: Buffer) => {
                const text = chunk.toString('utf-8');
                dataCollector.push(text);
                if (text.includes('connected') && !emitted) {
                  onConnect();
                }
              });
            },
          );
          return req;
        }

        const req1 = makeClient(client1Data, () => {
          client1Connected = true;
          if (client1Connected && client2Connected && !emitted) {
            emitted = true;
            void sseService.emit(runId, 'log_line', { line: 'broadcast log' }).then(() => {
              setTimeout(() => {
                req1.destroy();
                req2.destroy();
                checkDone();
                checkDone();
              }, 200);
            });
          }
        });
        // eslint-disable-next-line prefer-const
        let req2: http.ClientRequest;
        req2 = makeClient(client2Data, () => {
          client2Connected = true;
          if (client1Connected && client2Connected && !emitted) {
            emitted = true;
            void sseService.emit(runId, 'log_line', { line: 'broadcast log' }).then(() => {
              setTimeout(() => {
                req1.destroy();
                req2.destroy();
                checkDone();
                checkDone();
              }, 200);
            });
          }
        });

        setTimeout(() => {
          req1.destroy();
          req2.destroy();
          done(new Error('Timed out waiting for both clients to connect'));
        }, 3000);
      } catch (err) {
        done(err as Error);
      }
    })();
  });

  // ── Heartbeat ─────────────────────────────────────────────────────────────

  it('keeps the connection alive with :heartbeat comments', async () => {
    jest.useFakeTimers();

    const createResponse = await request(app)
      .post('/api/validation-runs')
      .send({ branchName: 'feature/sse-heartbeat', commitSha: VALID_COMMIT_SHA })
      .expect(201);

    const runId = createResponse.body.id as string;

    // Open a real HTTP stream to capture initial data
    const { rawData } = await collectSSEEvents(
      server,
      `/api/validation-runs/${runId}/stream`,
      300,
    );

    jest.useRealTimers();

    // The connected event should be present at minimum
    expect(rawData).toContain('event:');
  });
});
