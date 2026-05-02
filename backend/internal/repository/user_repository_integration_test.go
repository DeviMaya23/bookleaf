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

func TestUserRepository_GetOrCreate_DBError(t *testing.T) {
	db, err := testutil.NewTestDB(testContainer)
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.Close()

	repo := NewUserRepository(db)

	_, err = repo.GetOrCreate(context.Background(), "kp_abc123")
	require.Error(t, err)
}
