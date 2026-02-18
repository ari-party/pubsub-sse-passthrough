# pubsub-sse-passthrough

Basic Go app for forwarding Redis pub/sub messages to SSE.

## Usage

### Running locally

```bash
go run .
```

### Environment variables

| Variable                | Type           | Default                  | Description                                                                                                                                                                                                                                                         |
| ----------------------- | -------------- | ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| NODE_ENV                | `string`       | `development`            | The application environment.                                                                                                                                                                                                                                        |
| PORT                    | `number`       | `3000`                   | The port the server listens on. Parsed with JavaScript-like `parseInt` semantics (`123abc` becomes `123`).                                                                                                                                                          |
| REDIS_URL               | `string` (URL) | `redis://localhost:6379` | The URL for connecting to Redis.                                                                                                                                                                                                                                    |
| CHANNELS                | `string`       | `*`                      | Redis pub/sub channel pattern(s) to subscribe to. ([docs](https://redis.io/docs/latest/develop/pubsub/#pattern-matching-subscriptions))                                                                                                                             |
| HEARTBEAT_INTERVAL      | `number`       | `30`                     | Heartbeat interval in seconds for SSE connections. Parsed with JavaScript-like `parseInt` semantics.                                                                                                                                                                |
| SEND_RAW_REDIS_MESSAGES | `boolean`      | `true`                   | If true, forwards only the raw Redis message and uses the Redis channel as event name. If false, sends a JSON object with `message` and `channel`, using `message` as the event name. Only the exact string `true` maps to `true`; all other values map to `false`. |

Notes:

- Empty string values are treated as unset and fall back to defaults.
- Invalid numeric or URL values fail fast during startup.

### `/events`

SSE event stream. Redis messages are JSON-encoded and forwarded as events:

With SEND_RAW_REDIS_MESSAGES:

```text
id: int
event: channel
data: "message"

```

Without:

```text
id: int
event: message
data: {"message":"message","channel":"channel"}

```

Heartbeats are sent as comments:

```text
: heartbeat

```

### `/health`

Returns `OK` (200) or `NOT OK` (503) depending on the Redis connection status.
