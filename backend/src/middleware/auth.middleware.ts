import { Request, Response, NextFunction } from 'express';

export function adminAuthMiddleware(req: Request, res: Response, next: NextFunction): void {
  const adminToken = process.env.ADMIN_TOKEN;
  // Check X-Admin-Token header (Express lowercases header names)
  const providedToken = req.headers['x-admin-token'];

  if (!adminToken || !providedToken || providedToken !== adminToken) {
    res.status(403).json({ error: 'Forbidden' });
    return;
  }

  next();
}
