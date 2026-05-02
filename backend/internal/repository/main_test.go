package repository

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/devi/bookleaf/internal/testutil"
	"gorm.io/gorm"
)

var (
	testDB        *gorm.DB
	testContainer *testutil.PostgresContainer
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := testutil.SetupPostgresContainer(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "setup postgres container: %v\n", err)
		os.Exit(1)
	}
	testContainer = container

	db, err := testutil.NewTestDB(container)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create test db: %v\n", err)
		container.Terminate(ctx)
		os.Exit(1)
	}
	testDB = db

	code := m.Run()

	container.Terminate(ctx)
	os.Exit(code)
}
