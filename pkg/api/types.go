package api

import "time"

type SpatialEvent struct {
	SessionID string    `json:"session_id"`
	EventID   string    `json:"event_id"`
	Timestamp int64     `json:"timestamp"`
	Anchors   []Anchor  `json:"anchors"`
	Meshes    []Mesh    `json:"meshes"`
}

type Anchor struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Pose      Pose      `json:"pose"`
	Timestamp int64     `json:"timestamp"`
	Metadata  Metadata  `json:"metadata,omitempty"`
}

type Pose struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Z        float64 `json:"z"`
	Rotation [4]float64 `json:"rotation"` // Quaternion [x, y, z, w]
}

type Mesh struct {
	ID             string `json:"id"`
	AnchorID       string `json:"anchor_id"`
	Vertices       []byte `json:"vertices"`
	Faces          []byte `json:"faces"`
	IsDelta        bool   `json:"is_delta"`
	BaseMeshID     string `json:"base_mesh_id,omitempty"`
	CompressionLevel int  `json:"compression_level"`
	Timestamp      int64  `json:"timestamp"`
}

type TopologyEdge struct {
	From       string  `json:"from"`
	To         string  `json:"to"`
	Type       string  `json:"type"`
	Distance   float64 `json:"distance"`
	Confidence float64 `json:"confidence"`
}

type QueryParams struct {
	SessionID     string  `form:"session_id"`
	AnchorID      string  `form:"anchor_id"`
	Radius        float64 `form:"radius"`
	Since         int64   `form:"since"`
	IncludeMeshes bool    `form:"include_meshes"`
	Decompress    bool    `form:"decompress"`
	Limit         int     `form:"limit"`
	Offset        int     `form:"offset"`
}

type QueryResponse struct {
	Anchors   []Anchor       `json:"anchors"`
	Meshes    []Mesh         `json:"meshes,omitempty"`
	Topology  []TopologyEdge `json:"topology,omitempty"`
	Total     int            `json:"total"`
	HasMore   bool           `json:"has_more"`
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Database  string    `json:"database"`
}

type IngestResponse struct {
	Success   bool   `json:"success"`
	EventID   string `json:"event_id"`
	Processed int    `json:"processed"`
	Message   string `json:"message,omitempty"`
}

type Metadata map[string]interface{}

type WSMessage struct {
	Type      string      `json:"type"`
	SessionID string      `json:"session_id,omitempty"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

const (
	WSMessageTypeAnchorUpdate = "anchor_update"
	WSMessageTypeMeshUpdate   = "mesh_update"
	WSMessageTypeError        = "error"
	WSMessageTypePing         = "ping"
	WSMessageTypePong         = "pong"
)