# pubsub-sse-passthrough

Basic Express.js app for forwarding Redis pub/sub messages to SSE.

## Usage

### Environment variables

| Variable                | Type           | Default                  | Description                                                                                                                                                                                  |
| ----------------------- | -------------- | ------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| NODE_ENV                | `string`       | `development`            | The application environment. Modifies logging when set to `development`.                                                                                                                     |
| PORT                    | `number`       | `3000`                   | The port the Express server will listen on.                                                                                                                                                  |
| REDIS_URL               | `string` (URL) | `redis://localhost:6379` | The URL for connecting to the Redis server.                                                                                                                                                  |
| CHANNELS                | `string`       | `*`                      | Redis pub/sub channel(s) pattern to subscribe to. ([docs](https://redis.io/docs/latest/develop/pubsub/#pattern-matching-subscriptions))                                                      |
| HEARTBEAT_INTERVAL      | `number`       | `30`                     | Heartbeat interval (seconds) for SSE connections.                                                                                                                                            |
| SEND_RAW_REDIS_MESSAGES | `boolean`      | `true`                   | If true, forwards only the raw Redis message with the event name set to the Redis channel. If false, sends a JSON object with `message` and `channel`, and uses `message` as the event name. |

### `/events`

SSE event stream. Redis messages will be [converted to a JSON string](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/JSON/stringify) and forwarded as a message event:

With SEND_RAW_REDIS_MESSAGES:

```text
id: int
event: channel
data: message

```

Without:

```text
id: int
event: message
data: {"message":message,"channel":channel}

```

Heartbeats are sent as comments:

```text
: heartbeat

```

### `/health`

Returns `OK` (200) or `NOT OK` (503) depending on the Redis connection status.
