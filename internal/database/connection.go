package database

import (
	"context"
	"fmt"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/tabular/stag/internal/config"
)

type DB struct {
	client   driver.Client
	database driver.Database
	config   config.DatabaseConfig
}

func NewConnection(ctx context.Context, cfg config.DatabaseConfig) (*DB, error) {
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{cfg.URL},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	client, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication(cfg.Username, cfg.Password),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	dbExists, err := client.DatabaseExists(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to check database existence: %w", err)
	}

	var database driver.Database
	if !dbExists {
		database, err = client.CreateDatabase(ctx, cfg.Database, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create database: %w", err)
		}
	} else {
		database, err = client.Database(ctx, cfg.Database)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}
	}

	return &DB{
		client:   client,
		database: database,
		config:   cfg,
	}, nil
}

func (db *DB) Close() error {
	return nil
}

func (db *DB) Database() driver.Database {
	return db.database
}

func (db *DB) Client() driver.Client {
	return db.client
}