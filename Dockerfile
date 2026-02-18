# Build layer
FROM golang:1.23-alpine AS build

WORKDIR /build

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /pubsub-sse-passthrough .

# Package layer
FROM alpine:3.22 AS package

RUN apk --no-cache add ca-certificates curl

WORKDIR /app

COPY --from=build /pubsub-sse-passthrough /app/pubsub-sse-passthrough

CMD ["/app/pubsub-sse-passthrough"]

EXPOSE 3000
