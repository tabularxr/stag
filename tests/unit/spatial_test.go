package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tabular/stag/internal/spatial"
	"github.com/tabular/stag/pkg/api"
	"github.com/google/uuid"
)

func TestSpatialRepository_IngestEvent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := spatial.NewRepository(db)
	ctx := context.Background()

	event := &api.SpatialEvent{
		SessionID: "test-session",
		EventID:   uuid.New().String(),
		Timestamp: time.Now().UnixMilli(),
		Anchors: []api.Anchor{
			{
				ID:        "anchor-1",
				SessionID: "test-session",
				Pose: api.Pose{
					X:        1.0,
					Y:        2.0,
					Z:        3.0,
					Rotation: [4]float64{0, 0, 0, 1},
				},
				Timestamp: time.Now().UnixMilli(),
			},
		},
		Meshes: []api.Mesh{
			{
				ID:               "mesh-1",
				AnchorID:         "anchor-1",
				Vertices:         []byte("compressed-vertices"),
				Faces:            []byte("compressed-faces"),
				CompressionLevel: 5,
				Timestamp:        time.Now().UnixMilli(),
			},
		},
	}

	err := repo.IngestEvent(ctx, event)
	require.NoError(t, err)

	anchor, err := repo.GetAnchor(ctx, "anchor-1")
	require.NoError(t, err)
	assert.Equal(t, "anchor-1", anchor.ID)
	assert.Equal(t, "test-session", anchor.SessionID)
	assert.Equal(t, 1.0, anchor.Pose.X)
}

func TestSpatialRepository_Query(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := spatial.NewRepository(db)
	ctx := context.Background()

	event := &api.SpatialEvent{
		SessionID: "test-session",
		EventID:   uuid.New().String(),
		Timestamp: time.Now().UnixMilli(),
		Anchors: []api.Anchor{
			{
				ID:        "anchor-1",
				SessionID: "test-session",
				Pose: api.Pose{
					X:        1.0,
					Y:        2.0,
					Z:        3.0,
					Rotation: [4]float64{0, 0, 0, 1},
				},
				Timestamp: time.Now().UnixMilli(),
			},
			{
				ID:        "anchor-2",
				SessionID: "test-session",
				Pose: api.Pose{
					X:        4.0,
					Y:        5.0,
					Z:        6.0,
					Rotation: [4]float64{0, 0, 0, 1},
				},
				Timestamp: time.Now().UnixMilli(),
			},
		},
	}

	err := repo.IngestEvent(ctx, event)
	require.NoError(t, err)

	params := &api.QueryParams{
		SessionID:     "test-session",
		IncludeMeshes: false,
		Limit:         10,
	}

	response, err := repo.Query(ctx, params)
	require.NoError(t, err)
	assert.Len(t, response.Anchors, 2)
	assert.Equal(t, 2, response.Total)
}

func TestSpatialRepository_QueryWithAnchorID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := spatial.NewRepository(db)
	ctx := context.Background()

	event := &api.SpatialEvent{
		SessionID: "test-session",
		EventID:   uuid.New().String(),
		Timestamp: time.Now().UnixMilli(),
		Anchors: []api.Anchor{
			{
				ID:        "anchor-1",
				SessionID: "test-session",
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

	err := repo.IngestEvent(ctx, event)
	require.NoError(t, err)

	params := &api.QueryParams{
		AnchorID: "anchor-1",
		Limit:    10,
	}

	response, err := repo.Query(ctx, params)
	require.NoError(t, err)
	assert.Len(t, response.Anchors, 1)
	assert.Equal(t, "anchor-1", response.Anchors[0].ID)
}