package unit

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tabular/stag/internal/compression"
	"github.com/tabular/stag/pkg/api"
)

func TestSimpleCompressionService_CompressMesh(t *testing.T) {
	service := compression.NewSimpleCompressionService()

	vertices := []float32{
		0.0, 0.0, 0.0,  // vertex 0
		1.0, 0.0, 0.0,  // vertex 1  
		0.0, 1.0, 0.0,  // vertex 2
		0.0, 0.0, 1.0,  // vertex 3
	}

	faces := []uint32{
		0, 1, 2,  // triangle 1
		0, 2, 3,  // triangle 2
	}

	result, err := service.CompressMesh(vertices, faces, 7)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, result.OriginalSize, 0)
	assert.Greater(t, result.CompressedSize, 0)
	assert.Less(t, result.CompressedSize, result.OriginalSize)
	assert.Greater(t, result.CompressionRatio, 0.0)
	assert.Less(t, result.CompressionRatio, 1.0)

	t.Logf("Original size: %d bytes", result.OriginalSize)
	t.Logf("Compressed size: %d bytes", result.CompressedSize)
	t.Logf("Compression ratio: %.2f%% reduction", result.CompressionRatio*100)

	assert.Greater(t, result.CompressionRatio, 0.1, "Should achieve at least 10% compression")
}

func TestSimpleCompressionService_CompressDecompressRoundTrip(t *testing.T) {
	service := compression.NewSimpleCompressionService()

	originalVertices := []float32{
		0.0, 0.0, 0.0,
		1.0, 0.0, 0.0,
		0.0, 1.0, 0.0,
		0.0, 0.0, 1.0,
		2.0, 2.0, 2.0,
		-1.0, -1.0, -1.0,
	}

	originalFaces := []uint32{
		0, 1, 2,
		0, 2, 3,
		1, 4, 5,
	}

	compressResult, err := service.CompressMesh(originalVertices, originalFaces, 7)
	require.NoError(t, err)

	decompressResult, err := service.DecompressMesh(compressResult.CompressedData)
	require.NoError(t, err)
	assert.True(t, decompressResult.Success)

	assert.Len(t, decompressResult.Vertices, len(originalVertices))
	assert.Len(t, decompressResult.Faces, len(originalFaces))

	tolerance := float32(0.001)
	for i := range originalVertices {
		assert.InDelta(t, originalVertices[i], decompressResult.Vertices[i], float64(tolerance),
			"Vertex %d should be within tolerance", i)
	}

	for i := range originalFaces {
		assert.Equal(t, originalFaces[i], decompressResult.Faces[i], "Face %d should match", i)
	}
}

func TestSimpleCompressionService_CompressMeshFromAPI(t *testing.T) {
	service := compression.NewSimpleCompressionService()

	vertices := []float32{1.0, 2.0, 3.0, 4.0, 5.0, 6.0}
	faces := []uint32{0, 1, 2}

	verticesBytes := make([]byte, len(vertices)*4)
	facesBytes := make([]byte, len(faces)*4)

	for i, v := range vertices {
		bits := *(*uint32)(unsafe.Pointer(&v))
		verticesBytes[i*4] = byte(bits)
		verticesBytes[i*4+1] = byte(bits >> 8)
		verticesBytes[i*4+2] = byte(bits >> 16)
		verticesBytes[i*4+3] = byte(bits >> 24)
	}

	for i, f := range faces {
		facesBytes[i*4] = byte(f)
		facesBytes[i*4+1] = byte(f >> 8)
		facesBytes[i*4+2] = byte(f >> 16)
		facesBytes[i*4+3] = byte(f >> 24)
	}

	mesh := &api.Mesh{
		ID:               "test-mesh",
		AnchorID:         "test-anchor",
		Vertices:         verticesBytes,
		Faces:            facesBytes,
		CompressionLevel: 5,
	}

	result, err := service.CompressMeshFromAPI(mesh)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, result.OriginalSize, 0)
	assert.Greater(t, result.CompressedSize, 0)
}

func TestSimpleCompressionService_ComputeMeshDiff(t *testing.T) {
	service := compression.NewSimpleCompressionService()

	// Create base mesh
	baseVertices := []float32{0.0, 0.0, 0.0, 1.0, 0.0, 0.0}
	baseFaces := []uint32{0, 1, 2}
	baseVerticesBytes := float32SliceToBytes(baseVertices)
	baseFacesBytes := uint32SliceToBytes(baseFaces)

	baseMesh := &api.Mesh{
		ID:       "base-mesh",
		AnchorID: "test-anchor",
		Vertices: baseVerticesBytes,
		Faces:    baseFacesBytes,
	}

	// Create new mesh (slightly different)
	newVertices := []float32{0.1, 0.1, 0.1, 1.1, 0.1, 0.1}
	newFaces := []uint32{0, 1, 2}
	newVerticesBytes := float32SliceToBytes(newVertices)
	newFacesBytes := uint32SliceToBytes(newFaces)

	newMesh := &api.Mesh{
		ID:       "new-mesh",
		AnchorID: "test-anchor",
		Vertices: newVerticesBytes,
		Faces:    newFacesBytes,
	}

	diffMesh, err := service.ComputeMeshDiff(baseMesh, newMesh)
	require.NoError(t, err)
	assert.NotNil(t, diffMesh)
	assert.Equal(t, "new-mesh", diffMesh.ID)
	assert.Equal(t, "test-anchor", diffMesh.AnchorID)
}

func TestDracoService_Integration(t *testing.T) {
	service := compression.NewDracoService()

	vertices := []float32{
		0.0, 0.0, 0.0,
		1.0, 0.0, 0.0,
		0.0, 1.0, 0.0,
	}

	faces := []uint32{0, 1, 2}

	result, err := service.CompressMesh(vertices, faces, 7)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, result.CompressionRatio, 0.0)

	decompResult, err := service.DecompressMesh(result.CompressedData)
	require.NoError(t, err)
	assert.True(t, decompResult.Success)
	assert.Len(t, decompResult.Vertices, len(vertices))
	assert.Len(t, decompResult.Faces, len(faces))
}

func TestCompressionRatioHighEfficiency(t *testing.T) {
	service := compression.NewSimpleCompressionService()

	// Create a large mesh with repetitive data (should compress well)
	size := 300  // 300 triangles = 900 faces indices  
	vertices := make([]float32, size*3)
	faces := make([]uint32, size*3)  // 3 indices per triangle

	// Fill with pattern that should compress well
	for i := 0; i < size; i++ {
		vertices[i*3] = float32(i % 10)     // Repeating pattern
		vertices[i*3+1] = float32(i % 10)   // Repeating pattern
		vertices[i*3+2] = float32(i % 10)   // Repeating pattern
		faces[i*3] = uint32(i % 100)        // Repeating pattern
		faces[i*3+1] = uint32((i+1) % 100)  // Repeating pattern
		faces[i*3+2] = uint32((i+2) % 100)  // Repeating pattern
	}

	result, err := service.CompressMesh(vertices, faces, 9) // Max compression
	require.NoError(t, err)

	compressionRatio := result.CompressionRatio
	t.Logf("Large mesh compression ratio: %.2f%% reduction", compressionRatio*100)
	t.Logf("Original: %d bytes, Compressed: %d bytes", result.OriginalSize, result.CompressedSize)

	// Should achieve significant compression on repetitive data
	assert.Greater(t, compressionRatio, 0.80, "Should achieve >80% compression on repetitive data")
}

func float32SliceToBytes(data []float32) []byte {
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

func uint32SliceToBytes(data []uint32) []byte {
	result := make([]byte, len(data)*4)
	for i, u := range data {
		result[i*4] = byte(u)
		result[i*4+1] = byte(u >> 8)
		result[i*4+2] = byte(u >> 16)
		result[i*4+3] = byte(u >> 24)
	}
	return result
}