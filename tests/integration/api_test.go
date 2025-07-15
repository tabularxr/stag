package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tabular/stag/internal/config"
	"github.com/tabular/stag/internal/database"
	"github.com/tabular/stag/internal/server"
	"github.com/tabular/stag/pkg/api"
	"github.com/tabular/stag/pkg/logger"
	"github.com/google/uuid"
)

func TestIngestAndQuery_Integration(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	sessionID := uuid.New().String()
	anchorID := uuid.New().String()

	event := api.SpatialEvent{
		SessionID: sessionID,
		EventID:   uuid.New().String(),
		Timestamp: time.Now().UnixMilli(),
		Anchors: []api.Anchor{
			{
				ID:        anchorID,
				SessionID: sessionID,
				Pose: api.Pose{
					X:        10.5,
					Y:        20.7,
					Z:        30.9,
					Rotation: [4]float64{0.1, 0.2, 0.3, 0.9},
				},
				Timestamp: time.Now().UnixMilli(),
			},
		},
		Meshes: []api.Mesh{
			{
				ID:               uuid.New().String(),
				AnchorID:         anchorID,
				Vertices:         []byte("test-vertices-data"),
				Faces:            []byte("test-faces-data"),
				CompressionLevel: 7,
				Timestamp:        time.Now().UnixMilli(),
			},
		},
	}

	eventJSON, err := json.Marshal(event)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/ingest", bytes.NewBuffer(eventJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var ingestResponse api.IngestResponse
	err = json.Unmarshal(w.Body.Bytes(), &ingestResponse)
	require.NoError(t, err)
	assert.True(t, ingestResponse.Success)
	assert.Equal(t, 2, ingestResponse.Processed)

	time.Sleep(100 * time.Millisecond)

	queryURL := fmt.Sprintf("/api/v1/query?session_id=%s&include_meshes=true", sessionID)
	req = httptest.NewRequest("GET", queryURL, nil)
	w = httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var queryResponse api.QueryResponse
	err = json.Unmarshal(w.Body.Bytes(), &queryResponse)
	require.NoError(t, err)
	assert.Len(t, queryResponse.Anchors, 1)
	assert.Equal(t, anchorID, queryResponse.Anchors[0].ID)
	assert.Equal(t, 10.5, queryResponse.Anchors[0].Pose.X)
	assert.Len(t, queryResponse.Meshes, 1)
	assert.Equal(t, anchorID, queryResponse.Meshes[0].AnchorID)
}

func TestHealthEndpoint_Integration(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var healthResponse api.HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &healthResponse)
	require.NoError(t, err)
	assert.Equal(t, "healthy", healthResponse.Status)
	assert.Equal(t, "healthy", healthResponse.Database)
}

func TestGetAnchor_Integration(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	sessionID := uuid.New().String()
	anchorID := uuid.New().String()

	event := api.SpatialEvent{
		SessionID: sessionID,
		Anchors: []api.Anchor{
			{
				ID:        anchorID,
				SessionID: sessionID,
				Pose: api.Pose{
					X:        1.0,
					Y:        2.0,
					Z:        3.0,
					Rotation: [4]float64{0, 0, 0, 1},
				},
				Timestamp: time.Now().UnixMilli(),
			},
		},
	}

	eventJSON, _ := json.Marshal(event)
	req := httptest.NewRequest("POST", "/api/v1/ingest", bytes.NewBuffer(eventJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	time.Sleep(100 * time.Millisecond)

	getURL := fmt.Sprintf("/api/v1/anchors/%s", anchorID)
	req = httptest.NewRequest("GET", getURL, nil)
	w = httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var anchor api.Anchor
	err := json.Unmarshal(w.Body.Bytes(), &anchor)
	require.NoError(t, err)
	assert.Equal(t, anchorID, anchor.ID)
	assert.Equal(t, 1.0, anchor.Pose.X)
}

func setupTestServer(t *testing.T) (http.Handler, func()) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
			Host: "localhost",
		},
		Database: config.DatabaseConfig{
			URL:      "http://localhost:8529",
			Database: "stag_integration_test",
			Username: "root",
			Password: "stagpassword",
		},
		LogLevel: "error",
	}

	logger := logger.New(cfg.LogLevel)
	
	ctx := context.Background()
	db, err := database.NewConnection(ctx, cfg.Database)
	if err != nil {
		t.Skipf("Skipping integration test: ArangoDB not available: %v", err)
	}

	if err := database.Migrate(ctx, db); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	srv := server.New(cfg, db, logger)

	cleanup := func() {
		ctx := context.Background()
		client := db.Client()
		if exists, _ := client.DatabaseExists(ctx, cfg.Database.Database); exists {
			if database, err := client.Database(ctx, cfg.Database.Database); err == nil {
				database.Remove(ctx)
			}
		}
		db.Close()
	}

	return srv.Router(), cleanup
}