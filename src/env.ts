import { createEnv } from '@t3-oss/env-core';
import { z } from 'zod';

export const env = createEnv({
  server: {
    NODE_ENV: z.string().default('development'),

    PORT: z
      .string()
      .default('3000')
      .transform((v) => parseInt(v, 10))
      .pipe(z.number()),

    REDIS_URL: z.string().url().default('redis://localhost:6379'),

    CHANNELS: z.string().default('*'),

    HEARTBEAT_INTERVAL: z
      .string()
      .default('30')
      .transform((v) => parseInt(v, 10))
      .pipe(z.number()),

    SEND_RAW_REDIS_MESSAGES: z
      .string()
      .default('true')
      .transform((v) => v === 'true'),
  },
  runtimeEnv: process.env,
  emptyStringAsUndefined: true,
});
