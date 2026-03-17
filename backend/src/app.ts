import express from 'express';
import { registerRoutes } from './routes/index';
import { errorMiddleware } from './middleware/error.middleware';

export const app = express();

app.use(express.json());

registerRoutes(app);

app.use(errorMiddleware);
