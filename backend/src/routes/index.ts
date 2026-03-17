import { Express } from 'express';
import validationRunsRouter from './validation-runs.routes';

export function registerRoutes(app: Express): void {
  app.use('/api/validation-runs', validationRunsRouter);
}
