package spatial

import (
	"context"
	"fmt"
	"unsafe"

	"github.com/arangodb/go-driver"
	"github.com/tabular/stag/internal/compression"
	"github.com/tabular/stag/internal/database"
	"github.com/tabular/stag/pkg/api"
	"github.com/tabular/stag/pkg/errors"
)

type Repository struct {
	db            *database.DB
	dracoService  *compression.DracoService
	meshCache     map[string]*api.Mesh
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{
		db:           db,
		dracoService: compression.NewDracoService(),
		meshCache:    make(map[string]*api.Mesh),
	}
}

func (r *Repository) IngestEvent(ctx context.Context, event *api.SpatialEvent) error {
	for _, anchor := range event.Anchors {
		if err := r.upsertAnchor(ctx, &anchor); err != nil {
			return err
		}
	}

	for _, mesh := range event.Meshes {
		processedMesh, err := r.processMeshForStorage(ctx, &mesh)
		if err != nil {
			return fmt.Errorf("failed to process mesh %s: %w", mesh.ID, err)
		}
		
		if err := r.insertMesh(ctx, processedMesh); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) upsertAnchor(ctx context.Context, anchor *api.Anchor) error {
	db := r.db.Database()

	query := `
		UPSERT { _key: @key }
		INSERT @anchor
		UPDATE @anchor
		IN @@collection
	`

	bindVars := map[string]interface{}{
		"@collection": database.AnchorsCollection,
		"key":         anchor.ID,
		"anchor":      anchor,
	}

	_, err := db.Query(ctx, query, bindVars)
	if err != nil {
		return errors.DatabaseError(fmt.Sprintf("failed to upsert anchor: %v", err))
	}

	return nil
}

func (r *Repository) insertMesh(ctx context.Context, mesh *api.Mesh) error {
	db := r.db.Database()
	col, err := db.Collection(ctx, database.MeshesCollection)
	if err != nil {
		return errors.DatabaseError(fmt.Sprintf("failed to get meshes collection: %v", err))
	}

	meshDoc := map[string]interface{}{
		"_key":              mesh.ID,
		"anchor_id":         mesh.AnchorID,
		"vertices":          mesh.Vertices,
		"faces":             mesh.Faces,
		"is_delta":          mesh.IsDelta,
		"base_mesh_id":      mesh.BaseMeshID,
		"compression_level": mesh.CompressionLevel,
		"timestamp":         mesh.Timestamp,
	}

	_, err = col.CreateDocument(ctx, meshDoc)
	if err != nil {
		return errors.DatabaseError(fmt.Sprintf("failed to insert mesh: %v", err))
	}

	return nil
}

func (r *Repository) Query(ctx context.Context, params *api.QueryParams) (*api.QueryResponse, error) {
	db := r.db.Database()
	
	query := r.buildQuery(params)
	bindVars := r.buildBindVars(params)

	cursor, err := db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, errors.DatabaseError(fmt.Sprintf("failed to execute query: %v", err))
	}
	defer cursor.Close()

	var anchors []api.Anchor
	var meshes []api.Mesh

	for cursor.HasMore() {
		var result struct {
			Anchor api.Anchor `json:"anchor"`
			Mesh   *api.Mesh  `json:"mesh,omitempty"`
		}

		_, err := cursor.ReadDocument(ctx, &result)
		if err != nil {
			return nil, errors.DatabaseError(fmt.Sprintf("failed to read document: %v", err))
		}

		anchors = append(anchors, result.Anchor)
		if result.Mesh != nil && params.IncludeMeshes {
			meshes = append(meshes, *result.Mesh)
		}
	}

	response := &api.QueryResponse{
		Anchors: anchors,
		Total:   len(anchors),
		HasMore: len(anchors) == params.Limit,
	}

	if params.IncludeMeshes {
		response.Meshes = meshes
	}

	return response, nil
}

func (r *Repository) buildQuery(params *api.QueryParams) string {
	query := `
		FOR anchor IN anchors
	`

	if params.SessionID != "" {
		query += ` FILTER anchor.session_id == @session_id`
	}

	if params.AnchorID != "" {
		query += ` FILTER anchor._key == @anchor_id`
	}

	if params.Since > 0 {
		query += ` FILTER anchor.timestamp >= @since`
	}

	if params.Radius > 0 {
		query += ` FILTER GEO_DISTANCE([anchor.pose.x, anchor.pose.y], [@center_x, @center_y]) <= @radius`
	}

	query += ` SORT anchor.timestamp DESC`

	if params.Limit > 0 {
		query += ` LIMIT @offset, @limit`
	}

	if params.IncludeMeshes {
		query += `
			LET mesh = (
				FOR m IN meshes
				FILTER m.anchor_id == anchor._key
				SORT m.timestamp DESC
				LIMIT 1
				RETURN m
			)[0]
			RETURN { anchor: anchor, mesh: mesh }
		`
	} else {
		query += ` RETURN { anchor: anchor }`
	}

	return query
}

func (r *Repository) buildBindVars(params *api.QueryParams) map[string]interface{} {
	bindVars := make(map[string]interface{})

	if params.SessionID != "" {
		bindVars["session_id"] = params.SessionID
	}

	if params.AnchorID != "" {
		bindVars["anchor_id"] = params.AnchorID
	}

	if params.Since > 0 {
		bindVars["since"] = params.Since
	}

	if params.Radius > 0 {
		bindVars["radius"] = params.Radius
		bindVars["center_x"] = 0.0
		bindVars["center_y"] = 0.0
	}

	if params.Limit > 0 {
		bindVars["limit"] = params.Limit
		bindVars["offset"] = params.Offset
	} else {
		bindVars["limit"] = 100
		bindVars["offset"] = 0
	}

	return bindVars
}

func (r *Repository) GetAnchor(ctx context.Context, anchorID string) (*api.Anchor, error) {
	db := r.db.Database()
	col, err := db.Collection(ctx, database.AnchorsCollection)
	if err != nil {
		return nil, errors.DatabaseError(fmt.Sprintf("failed to get collection: %v", err))
	}

	var anchor api.Anchor
	_, err = col.ReadDocument(ctx, anchorID, &anchor)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, errors.NotFound("anchor not found")
		}
		return nil, errors.DatabaseError(fmt.Sprintf("failed to read anchor: %v", err))
	}

	return &anchor, nil
}

func (r *Repository) processMeshForStorage(ctx context.Context, mesh *api.Mesh) (*api.Mesh, error) {
	if len(mesh.Vertices) == 0 {
		return mesh, nil
	}

	if r.isMeshAlreadyCompressed(mesh.Vertices) {
		return mesh, nil
	}

	result, err := r.dracoService.CompressMeshFromAPI(mesh)
	if err != nil {
		return nil, fmt.Errorf("failed to compress mesh: %w", err)
	}

	baseMesh := r.meshCache[mesh.AnchorID]
	if baseMesh != nil && !mesh.IsDelta {
		deltaMesh, err := r.dracoService.ComputeMeshDiff(baseMesh, mesh)
		if err == nil && len(deltaMesh.Vertices) < len(result.CompressedData) {
			r.meshCache[mesh.AnchorID] = mesh
			return deltaMesh, nil
		}
	}

	r.meshCache[mesh.AnchorID] = mesh

	compressedMesh := &api.Mesh{
		ID:               mesh.ID,
		AnchorID:         mesh.AnchorID,
		Vertices:         result.CompressedData,
		Faces:            []byte{},
		IsDelta:          false,
		CompressionLevel: mesh.CompressionLevel,
		Timestamp:        mesh.Timestamp,
	}

	return compressedMesh, nil
}

func (r *Repository) isMeshAlreadyCompressed(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	
	return data[0] == 'S' && data[1] == 'T' && data[2] == 'A' && data[3] == 'G'
}

func (r *Repository) QueryWithDecompression(ctx context.Context, params *api.QueryParams, decompress bool) (*api.QueryResponse, error) {
	response, err := r.Query(ctx, params)
	if err != nil {
		return nil, err
	}

	if !decompress || len(response.Meshes) == 0 {
		return response, nil
	}

	decompressedMeshes := make([]api.Mesh, 0, len(response.Meshes))
	
	for _, mesh := range response.Meshes {
		if mesh.IsDelta {
			resolvedMesh, err := r.resolveDeltaMesh(ctx, &mesh)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve delta mesh %s: %w", mesh.ID, err)
			}
			mesh = *resolvedMesh
		}

		if len(mesh.Vertices) > 0 {
			decompResult, err := r.dracoService.DecompressMesh(mesh.Vertices)
			if err != nil {
				return nil, fmt.Errorf("failed to decompress mesh %s: %w", mesh.ID, err)
			}

			if decompResult.Success {
				mesh.Vertices = float32ArrayToBytes(decompResult.Vertices)
				mesh.Faces = uint32ArrayToBytes(decompResult.Faces)
			}
		}

		decompressedMeshes = append(decompressedMeshes, mesh)
	}

	response.Meshes = decompressedMeshes
	return response, nil
}

func (r *Repository) resolveDeltaMesh(ctx context.Context, deltaMesh *api.Mesh) (*api.Mesh, error) {
	if !deltaMesh.IsDelta || deltaMesh.BaseMeshID == "" {
		return deltaMesh, nil
	}

	baseMeshData, err := r.getMeshData(ctx, deltaMesh.BaseMeshID)
	if err != nil {
		return nil, fmt.Errorf("failed to get base mesh %s: %w", deltaMesh.BaseMeshID, err)
	}

	reconstructedData, err := r.dracoService.ApplyMeshDiff(baseMeshData, deltaMesh.Vertices)
	if err != nil {
		return nil, fmt.Errorf("failed to apply mesh diff: %w", err)
	}

	resolvedMesh := &api.Mesh{
		ID:               deltaMesh.ID,
		AnchorID:         deltaMesh.AnchorID,
		Vertices:         reconstructedData,
		Faces:            []byte{},
		IsDelta:          false,
		CompressionLevel: deltaMesh.CompressionLevel,
		Timestamp:        deltaMesh.Timestamp,
	}

	return resolvedMesh, nil
}

func (r *Repository) getMeshData(ctx context.Context, meshID string) ([]byte, error) {
	db := r.db.Database()
	col, err := db.Collection(ctx, database.MeshesCollection)
	if err != nil {
		return nil, errors.DatabaseError(fmt.Sprintf("failed to get collection: %v", err))
	}

	var meshDoc struct {
		Vertices []byte `json:"vertices"`
	}
	
	_, err = col.ReadDocument(ctx, meshID, &meshDoc)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, errors.NotFound("mesh not found")
		}
		return nil, errors.DatabaseError(fmt.Sprintf("failed to read mesh: %v", err))
	}

	return meshDoc.Vertices, nil
}

func float32ArrayToBytes(data []float32) []byte {
	result := make([]byte, len(data)*4)
	for i, f := range data {
		bits := *(*uint32)(unsafe.Pointer(&f))
		result[i*4] = byte(bits)
		result[i*4+1] = byte(bits >> 8)
		result[i*4+2] = byte(bits >> 16)
		result[i*4+3] = byte(bits >> 24)
	}
	return result
}

func uint32ArrayToBytes(data []uint32) []byte {
	result := make([]byte, len(data)*4)
	for i, u := range data {
		result[i*4] = byte(u)
		result[i*4+1] = byte(u >> 8)
		result[i*4+2] = byte(u >> 16)
		result[i*4+3] = byte(u >> 24)
	}
	return result
}