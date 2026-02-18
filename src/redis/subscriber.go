package redis

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
)

type Handler func(channel string, message string)

type Subscriber struct {
	client   *goredis.Client
	pubsub   *goredis.PubSub
	channels string
	logger   *log.Logger
	ready    atomic.Bool
}

func NewSubscriber(redisURL string, channels string, logger *log.Logger) (*Subscriber, error) {
	opts, err := goredis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	opts.MaintNotificationsConfig = &maintnotifications.Config{
		Mode: maintnotifications.ModeDisabled,
	}

	client := goredis.NewClient(opts)

	return &Subscriber{
		client:   client,
		channels: channels,
		logger:   logger,
	}, nil
}

func (s *Subscriber) Start(ctx context.Context, onMessage Handler) error {
	if err := s.client.Ping(ctx).Err(); err != nil {
		return err
	}

	s.pubsub = s.client.PSubscribe(ctx, s.channels)
	if _, err := s.pubsub.Receive(ctx); err != nil {
		return err
	}

	s.ready.Store(true)
	s.logger.Printf("Redis connected and subscribed to %q", s.channels)

	go func() {
		for {
			msg, err := s.pubsub.ReceiveMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}

				s.ready.Store(false)
				s.logger.Printf("Redis receive error: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			s.ready.Store(true)
			onMessage(msg.Channel, msg.Payload)
		}
	}()

	return nil
}

func (s *Subscriber) Healthy(ctx context.Context) bool {
	if !s.ready.Load() {
		return false
	}

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return s.client.Ping(pingCtx).Err() == nil
}

func (s *Subscriber) Close() error {
	if s.pubsub != nil {
		_ = s.pubsub.Close()
	}

	return s.client.Close()
}
