package testutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetupPostgresContainer(t *testing.T) {
	ctx := context.Background()

	container, err := SetupPostgresContainer(ctx)
	require.NoError(t, err)
	defer container.Terminate(ctx)

	db, err := NewTestDB(container)
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Ping())

	tx := NewTestTx(t, db)
	require.NotNil(t, tx)
	require.NoError(t, tx.Error)
}
