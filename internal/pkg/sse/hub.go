package sse

import (
	"context"
	"sync"
)

type Hub struct {
	mu          sync.RWMutex
	subscribers map[string]map[string]chan string
	counter     int
}

func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[string]map[string]chan string),
	}
}

func (h *Hub) Subscribe(ctx context.Context, topic string) (<-chan string, error) {
	h.mu.Lock()
	h.counter++
	id := topic + "-" + itoa(h.counter)
	ch := make(chan string, 64)

	if h.subscribers[topic] == nil {
		h.subscribers[topic] = make(map[string]chan string)
	}
	h.subscribers[topic][id] = ch
	h.mu.Unlock()

	go func() {
		<-ctx.Done()
		h.mu.Lock()
		delete(h.subscribers[topic], id)
		if len(h.subscribers[topic]) == 0 {
			delete(h.subscribers, topic)
		}
		h.mu.Unlock()
	}()

	return ch, nil
}

func (h *Hub) Publish(topic string, data string) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, ch := range h.subscribers[topic] {
		select {
		case ch <- data:
		default:
		}
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [10]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
