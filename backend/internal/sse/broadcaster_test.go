package sse

import (
	"testing"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventBroadcaster_PublishDeliversToAllSubscribersForUser(t *testing.T) {
	b := NewEventBroadcaster()
	ch1 := b.Subscribe("user1")
	ch2 := b.Subscribe("user1")
	event := domain.Event{Type: "categorisation_complete"}

	b.Publish("user1", event)

	require.Len(t, ch1, 1)
	assert.Equal(t, event, <-ch1)
	require.Len(t, ch2, 1)
	assert.Equal(t, event, <-ch2)
}

func TestEventBroadcaster_PublishDoesNotDeliverToDifferentUser(t *testing.T) {
	b := NewEventBroadcaster()
	ch := b.Subscribe("user2")

	b.Publish("user1", domain.Event{Type: "categorisation_complete"})

	assert.Len(t, ch, 0)
}

func TestEventBroadcaster_UnsubscribedChannelNoLongerReceivesEvents(t *testing.T) {
	b := NewEventBroadcaster()
	ch := b.Subscribe("user1")
	b.Unsubscribe("user1", ch)

	b.Publish("user1", domain.Event{Type: "categorisation_complete"})

	assert.Len(t, ch, 0)
}
