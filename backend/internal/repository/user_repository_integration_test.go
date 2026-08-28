package repository

import (
	"context"
	"testing"

	"github.com/devi/bookleaf/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_GetOrCreate_Success(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	repo := NewUserRepository(tx)

	user, err := repo.GetOrCreate(context.Background(), "kp_abc123")

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.Equal(t, "kp_abc123", user.IDPSubject)
	assert.False(t, user.VisionEnabled)
}

func TestUserRepository_GetOrCreate_ExistingUser_ReturnsSame(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	repo := NewUserRepository(tx)

	first, err := repo.GetOrCreate(context.Background(), "kp_idempotent")
	require.NoError(t, err)

	second, err := repo.GetOrCreate(context.Background(), "kp_idempotent")
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID)
	assert.Equal(t, "kp_idempotent", second.IDPSubject)
}

