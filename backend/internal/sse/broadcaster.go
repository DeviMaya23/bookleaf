// Package sse provides an in-process event broadcaster for Server-Sent Events.
// Single-server assumption: events are not propagated across multiple instances.
// To support horizontal scaling, replace this with a Redis pub/sub or Postgres LISTEN/NOTIFY.
package sse

import (
	"sync"

	"github.com/devi/bookleaf/internal/domain"
)

type EventBroadcaster struct {
	mu      sync.RWMutex
	clients map[string][]chan domain.Event
}

func NewEventBroadcaster() *EventBroadcaster {
	return &EventBroadcaster{
		clients: make(map[string][]chan domain.Event),
	}
}

func (b *EventBroadcaster) Subscribe(userID string) chan domain.Event {
	ch := make(chan domain.Event, 16)
	b.mu.Lock()
	b.clients[userID] = append(b.clients[userID], ch)
	b.mu.Unlock()
	return ch
}

func (b *EventBroadcaster) Unsubscribe(userID string, ch chan domain.Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	channels := b.clients[userID]
	for i, c := range channels {
		if c == ch {
			b.clients[userID] = append(channels[:i], channels[i+1:]...)
			break
		}
	}
	if len(b.clients[userID]) == 0 {
		delete(b.clients, userID)
	}
}

func (b *EventBroadcaster) Publish(userID string, event domain.Event) {
	b.mu.RLock()
	channels := make([]chan domain.Event, len(b.clients[userID]))
	copy(channels, b.clients[userID])
	b.mu.RUnlock()

	for _, ch := range channels {
		select {
		case ch <- event:
		default:
			// buffer full — drop event rather than blocking the caller
		}
	}
}
