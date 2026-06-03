package repository

import (
	"context"
	"testing"

	"github.com/devi/bookleaf/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_GetOrCreate_Success(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	repo := NewUserRepository(tx)

	user, err := repo.GetOrCreate(context.Background(), "kp_abc123")

	require.NoError(t, err)
	assert.Equal(t, "kp_abc123", user.ID)
	assert.False(t, user.VisionEnabled)
}

