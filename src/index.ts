import express from 'express';

import { env } from '@/env';
import { log } from '@/pino';
import { redis } from '@/redis';
import { SSE } from '@/sse';

const app = express();
const sse = new SSE(env.HEARTBEAT_INTERVAL);

app.disable('x-powered-by');

app.get('/health', (_req, res) => {
  if (redis.status === 'ready') res.status(200).write('OK');
  else res.status(503).write('NOT OK');

  res.end();
});

app.get('/events', sse.init);
redis.on('pmessage', (channel, message) =>
  sse.send({ message, channel }, 'message'),
);

app.listen(env.PORT, '0.0.0.0', () =>
  log.info(`Server listening on 0.0.0.0:${env.PORT}`),
);
