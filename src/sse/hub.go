package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type payload struct {
	Data  any
	Event string
}

type client struct {
	id int64
	ch chan payload
}

type Hub struct {
	heartbeatInterval time.Duration

	mu      sync.RWMutex
	clients map[*client]struct{}
}

func NewHub(heartbeatInterval time.Duration) *Hub {
	return &Hub{
		heartbeatInterval: heartbeatInterval,
		clients:           make(map[*client]struct{}),
	}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	if r.ProtoMajor < 2 {
		w.Header().Set("Connection", "keep-alive")
	}
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(": connected\n\n")); err != nil {
		return
	}
	flusher.Flush()

	c := &client{
		ch: make(chan payload, 64),
	}

	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, c)
		h.mu.Unlock()
	}()

	heartbeatTicker := time.NewTicker(h.heartbeatInterval)
	defer heartbeatTicker.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeatTicker.C:
			if _, err := w.Write([]byte(": heartbeat\n\n")); err != nil {
				return
			}
			flusher.Flush()
		case msg, ok := <-c.ch:
			if !ok {
				return
			}

			encoded, err := json.Marshal(msg.Data)
			if err != nil {
				continue
			}

			if _, err := fmt.Fprintf(w, "id: %d\n", c.id); err != nil {
				return
			}
			c.id++

			if msg.Event != "" {
				if _, err := fmt.Fprintf(w, "event: %s\n", msg.Event); err != nil {
					return
				}
			}

			if _, err := fmt.Fprintf(w, "data: %s\n\n", encoded); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (h *Hub) Publish(data any, event string) {
	msg := payload{
		Data:  data,
		Event: event,
	}

	h.mu.RLock()
	clients := make([]*client, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	for _, c := range clients {
		select {
		case c.ch <- msg:
		default:
		}
	}
}
