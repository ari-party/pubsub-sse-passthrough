package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ari-party/pubsub-sse-passthrough/src/config"
	redisstream "github.com/ari-party/pubsub-sse-passthrough/src/redis"
	"github.com/ari-party/pubsub-sse-passthrough/src/sse"
)

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("invalid environment: %v", err)
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	heartbeatInterval := time.Duration(cfg.HeartbeatIntervalSec) * time.Second
	hub := sse.NewHub(heartbeatInterval)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	subscriber, err := redisstream.NewSubscriber(cfg.RedisURL, cfg.Channels, logger)
	if err != nil {
		log.Fatalf("failed to configure redis client: %v", err)
	}
	defer func() {
		if closeErr := subscriber.Close(); closeErr != nil {
			logger.Printf("redis close error: %v", closeErr)
		}
	}()

	if err := subscriber.Start(ctx, func(channel string, message string) {
		if cfg.SendRawRedisMessages {
			hub.Publish(message, channel)
			return
		}

		hub.Publish(map[string]string{
			"message": message,
			"channel": channel,
		}, "message")
	}); err != nil {
		log.Fatalf("failed to subscribe to redis channel pattern %q: %v", cfg.Channels, err)
	}

	mux := http.NewServeMux()
	mux.Handle("/events", hub)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if subscriber.Healthy(r.Context()) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
			return
		}

		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("NOT OK"))
	})

	server := &http.Server{
		Addr:              "0.0.0.0:" + strconv.Itoa(cfg.Port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Printf("Server listening on 0.0.0.0:%d", cfg.Port)
		if serveErr := server.ListenAndServe(); serveErr != nil && serveErr != http.ErrServerClosed {
			log.Fatalf("http server error: %v", serveErr)
		}
	}()

	<-ctx.Done()
	logger.Print("Shutting down server")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Printf("graceful shutdown failed: %v", err)
		if closeErr := server.Close(); closeErr != nil {
			logger.Printf("forced close failed: %v", closeErr)
		}
	}

	fmt.Println("Server stopped")
}
