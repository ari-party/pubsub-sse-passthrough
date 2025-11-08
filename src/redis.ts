import Redis from 'ioredis';

import { env } from '@/env';
import { log } from '@/pino';

export const redis = new Redis(env.REDIS_URL);

redis.on('connecting', () => log.info({ name: 'Redis' }, 'Connecting'));
redis.on('connect', () => log.info({ name: 'Redis' }, 'Connected'));
redis.on('error', (error) => log.error({ ...error, name: 'Redis' }, 'Error'));

redis
  .psubscribe(env.CHANNELS)
  .then(() => log.info({ name: 'Redis' }, `Subscribed to \`${env.CHANNELS}\``))
  .catch((err) => {
    log.error(
      { ...err, name: 'Redis' },
      `Failed to subscribe to \`${env.CHANNELS}\``,
    );
    process.exit(1);
  });
