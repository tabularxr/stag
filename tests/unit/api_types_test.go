package unit

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tabular/stag/pkg/api"
	"github.com/google/uuid"
)

func TestSpatialEvent_JSON(t *testing.T) {
	event := api.SpatialEvent{
		SessionID: "test-session",
		EventID:   uuid.New().String(),
		Timestamp: time.Now().UnixMilli(),
		Anchors: []api.Anchor{
			{
				ID:        "anchor-1",
				SessionID: "test-session",
				Pose: api.Pose{
					X:        1.5,
					Y:        2.5,
					Z:        3.5,
					Rotation: [4]float64{0.1, 0.2, 0.3, 0.9},
				},
				Timestamp: time.Now().UnixMilli(),
			},
		},
		Meshes: []api.Mesh{
			{
				ID:               "mesh-1",
				AnchorID:         "anchor-1",
				Vertices:         []byte("test-vertices"),
				Faces:            []byte("test-faces"),
				CompressionLevel: 5,
				Timestamp:        time.Now().UnixMilli(),
			},
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(event)
	require.NoError(t, err)
	assert.Contains(t, string(data), "test-session")

	// Test JSON unmarshaling
	var decoded api.SpatialEvent
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	
	assert.Equal(t, event.SessionID, decoded.SessionID)
	assert.Len(t, decoded.Anchors, 1)
	assert.Equal(t, "anchor-1", decoded.Anchors[0].ID)
	assert.Equal(t, 1.5, decoded.Anchors[0].Pose.X)
	assert.Len(t, decoded.Meshes, 1)
	assert.Equal(t, "mesh-1", decoded.Meshes[0].ID)
}

func TestQueryParams_Validation(t *testing.T) {
	params := api.QueryParams{
		SessionID:     "test-session",
		Radius:        10.5,
		IncludeMeshes: true,
		Limit:         100,
		Offset:        0,
	}

	assert.Equal(t, "test-session", params.SessionID)
	assert.Equal(t, 10.5, params.Radius)
	assert.True(t, params.IncludeMeshes)
	assert.Equal(t, 100, params.Limit)
}

func TestWSMessage_Types(t *testing.T) {
	tests := []struct {
		msgType  string
		expected string
	}{
		{api.WSMessageTypeAnchorUpdate, "anchor_update"},
		{api.WSMessageTypeMeshUpdate, "mesh_update"},
		{api.WSMessageTypeError, "error"},
		{api.WSMessageTypePing, "ping"},
		{api.WSMessageTypePong, "pong"},
	}

	for _, tt := range tests {
		t.Run(tt.msgType, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.msgType)
		})
	}
}