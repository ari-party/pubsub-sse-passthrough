# pubsub-sse-passthrough

Basic Express.js app for forwarding Redis pub/sub messages to SSE.

## Usage

### Environment variables

| Variable           | Type           | Default                  | Description                                                                                                                             |
| ------------------ | -------------- | ------------------------ | --------------------------------------------------------------------------------------------------------------------------------------- |
| NODE_ENV           | `string`       | `development`            | The application environment. Modifies logging when set to `development`.                                                                |
| PORT               | `number`       | `3000`                   | The port the Express server will listen on.                                                                                             |
| REDIS_URL          | `string` (URL) | `redis://localhost:6379` | The URL for connecting to the Redis server.                                                                                             |
| CHANNELS           | `string`       | `*`                      | Redis pub/sub channel(s) pattern to subscribe to. ([docs](https://redis.io/docs/latest/develop/pubsub/#pattern-matching-subscriptions)) |
| HEARTBEAT_INTERVAL | `number`       | `30`                     | Heartbeat interval (seconds) for SSE connections.                                                                                       |

### `/events`

SSE event stream. Redis messages will be forwarded as a message event:

```text
id: int
event: message
data: {"message":string,"channel":string}

```

Heartbeats are sent as comments:

```text
: heartbeat

```

### `/health`

Returns `OK` (200) or `NOT OK` (503) depending on the Redis connection status.
