// SSE controller — TASK-4596
// Stub: validates runId exists, then calls sseService.subscribe(runId, res).
// Does NOT return (keeps connection open).

import { Request, Response, NextFunction } from 'express';

export async function subscribe(req: Request, res: Response, next: NextFunction): Promise<void> {
  try {
    void req;
    void res;
    void next;
  } catch (err) {
    next(err);
  }
}
