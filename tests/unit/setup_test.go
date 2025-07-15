package unit

import (
	"context"
	"testing"

	"github.com/tabular/stag/internal/config"
	"github.com/tabular/stag/internal/database"
)

func setupTestDB(t *testing.T) *database.DB {
	cfg := config.DatabaseConfig{
		URL:      "http://localhost:8529",
		Database: "stag_test",
		Username: "root",
		Password: "stagpassword",
	}

	ctx := context.Background()
	db, err := database.NewConnection(ctx, cfg)
	if err != nil {
		t.Skipf("Skipping test: ArangoDB not available: %v", err)
	}

	if err := database.Migrate(ctx, db); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	t.Cleanup(func() {
		ctx := context.Background()
		client := db.Client()
		if exists, _ := client.DatabaseExists(ctx, cfg.Database); exists {
			if database, err := client.Database(ctx, cfg.Database); err == nil {
				database.Remove(ctx)
			}
		}
	})

	return db
}