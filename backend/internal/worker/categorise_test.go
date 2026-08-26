package worker

import (
	"context"
	"errors"
	"testing"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCategorisationUsecase struct {
	err error
}

func (m *mockCategorisationUsecase) CategoriseImage(_ context.Context, _ string, _ uuid.UUID) error {
	return m.err
}

type mockCategorisationBroadcaster struct {
	published []domain.Event
	userIDs   []string
}

func (m *mockCategorisationBroadcaster) Publish(userID string, event domain.Event) {
	m.userIDs = append(m.userIDs, userID)
	m.published = append(m.published, event)
}

func TestCategorisationWorker_Work_OnSuccessPublishesBroadcastEvent(t *testing.T) {
	broadcaster := &mockCategorisationBroadcaster{}
	w := NewCategorisationWorker(&mockCategorisationUsecase{}, broadcaster)
	imageID := uuid.New()
	job := &river.Job[CategoriseImageArgs]{
		Args: CategoriseImageArgs{UserID: "user1", ImageID: imageID},
	}

	err := w.Work(context.Background(), job)

	require.NoError(t, err)
	require.Len(t, broadcaster.published, 1)
	assert.Equal(t, "user1", broadcaster.userIDs[0])
	assert.Equal(t, "categorisation_complete", broadcaster.published[0].Type)
	assert.Contains(t, string(broadcaster.published[0].Payload), imageID.String())
}

func TestCategorisationWorker_Work_OnUsecaseErrorBroadcasterNotCalled(t *testing.T) {
	broadcaster := &mockCategorisationBroadcaster{}
	w := NewCategorisationWorker(&mockCategorisationUsecase{err: errors.New("categorisation failed")}, broadcaster)
	job := &river.Job[CategoriseImageArgs]{
		Args: CategoriseImageArgs{UserID: "user1", ImageID: uuid.New()},
	}

	err := w.Work(context.Background(), job)

	assert.Error(t, err)
	assert.Empty(t, broadcaster.published)
}
