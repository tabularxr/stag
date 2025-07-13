package compression

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"

	"github.com/tabular/stag/internal/metrics"
	"github.com/tabular/stag/pkg/api"
)

type CompressionResult struct {
	CompressedData   []byte
	OriginalSize     int
	CompressedSize   int
	CompressionRatio float64
}

type DecompressionResult struct {
	Vertices []float32
	Faces    []uint32
	Success  bool
}

type SimpleCompressionService struct {
	compressionLevel int
}

func NewSimpleCompressionService() *SimpleCompressionService {
	return &SimpleCompressionService{
		compressionLevel: gzip.BestCompression,
	}
}

func (s *SimpleCompressionService) CompressMesh(vertices []float32, faces []uint32, compressionLevel int) (*CompressionResult, error) {
	if compressionLevel <= 0 {
		compressionLevel = s.compressionLevel
	}

	if len(vertices)%3 != 0 {
		return nil, fmt.Errorf("vertices must be divisible by 3 (x,y,z coordinates)")
	}

	if len(faces)%3 != 0 {
		return nil, fmt.Errorf("faces must be divisible by 3 (triangle indices)")
	}


	var buf bytes.Buffer
	buf.Write([]byte("STAG"))
	
	binary.Write(&buf, binary.LittleEndian, uint32(len(vertices)))
	binary.Write(&buf, binary.LittleEndian, uint32(len(faces)))
	
	for _, v := range vertices {
		binary.Write(&buf, binary.LittleEndian, v)
	}
	
	for _, f := range faces {
		binary.Write(&buf, binary.LittleEndian, f)
	}

	var compressed bytes.Buffer
	writer, err := gzip.NewWriterLevel(&compressed, compressionLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip writer: %w", err)
	}

	if _, err := writer.Write(buf.Bytes()); err != nil {
		return nil, fmt.Errorf("failed to compress data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	originalSize := len(vertices)*4 + len(faces)*4
	compressedSize := compressed.Len()
	ratio := 1.0 - (float64(compressedSize) / float64(originalSize))

	metrics.MeshProcessingDuration.Observe(0.001)
	metrics.StorageUsage.WithLabelValues("compressed").Set(float64(compressedSize))
	metrics.StorageUsage.WithLabelValues("original").Set(float64(originalSize))
	metrics.CompressionRatio.WithLabelValues(fmt.Sprintf("%d", compressionLevel)).Observe(ratio)
	metrics.CompressionOperations.WithLabelValues("compress", "success").Inc()
	metrics.MeshSizeBytes.WithLabelValues("original").Observe(float64(originalSize))
	metrics.MeshSizeBytes.WithLabelValues("compressed").Observe(float64(compressedSize))

	return &CompressionResult{
		CompressedData:   compressed.Bytes(),
		OriginalSize:     originalSize,
		CompressedSize:   compressedSize,
		CompressionRatio: ratio,
	}, nil
}

func (s *SimpleCompressionService) DecompressMesh(compressedData []byte) (*DecompressionResult, error) {
	reader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		metrics.CompressionOperations.WithLabelValues("decompress", "error").Inc()
		return &DecompressionResult{Success: false}, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		metrics.CompressionOperations.WithLabelValues("decompress", "error").Inc()
		return &DecompressionResult{Success: false}, fmt.Errorf("failed to decompress data: %w", err)
	}

	if len(decompressed) < 16 {
		metrics.CompressionOperations.WithLabelValues("decompress", "error").Inc()
		return &DecompressionResult{Success: false}, fmt.Errorf("invalid compressed data: too short")
	}

	header := string(decompressed[:4])
	if header != "STAG" {
		metrics.CompressionOperations.WithLabelValues("decompress", "error").Inc()
		return &DecompressionResult{Success: false}, fmt.Errorf("invalid header: %s", header)
	}

	buf := bytes.NewReader(decompressed[4:])
	
	var vertexCount, faceCount uint32
	binary.Read(buf, binary.LittleEndian, &vertexCount)
	binary.Read(buf, binary.LittleEndian, &faceCount)

	vertices := make([]float32, vertexCount)
	for i := uint32(0); i < vertexCount; i++ {
		binary.Read(buf, binary.LittleEndian, &vertices[i])
	}

	faces := make([]uint32, faceCount)
	for i := uint32(0); i < faceCount; i++ {
		binary.Read(buf, binary.LittleEndian, &faces[i])
	}

	metrics.CompressionOperations.WithLabelValues("decompress", "success").Inc()

	return &DecompressionResult{
		Vertices: vertices,
		Faces:    faces,
		Success:  true,
	}, nil
}

func (s *SimpleCompressionService) CompressMeshFromAPI(mesh *api.Mesh) (*CompressionResult, error) {
	if len(mesh.Vertices) == 0 || len(mesh.Faces) == 0 {
		return nil, fmt.Errorf("mesh vertices and faces cannot be empty")
	}

	vertices, err := bytesToFloat32Array(mesh.Vertices)
	if err != nil {
		return nil, fmt.Errorf("invalid vertices data: %w", err)
	}

	faces, err := bytesToUint32Array(mesh.Faces)
	if err != nil {
		return nil, fmt.Errorf("invalid faces data: %w", err)
	}

	return s.CompressMesh(vertices, faces, mesh.CompressionLevel)
}

func (s *SimpleCompressionService) ComputeMeshDiff(baseMesh, newMesh *api.Mesh) (*api.Mesh, error) {
	baseResult, err := s.CompressMeshFromAPI(baseMesh)
	if err != nil {
		return nil, fmt.Errorf("failed to compress base mesh: %w", err)
	}

	newResult, err := s.CompressMeshFromAPI(newMesh)
	if err != nil {
		return nil, fmt.Errorf("failed to compress new mesh: %w", err)
	}

	delta := computeBinaryDiff(baseResult.CompressedData, newResult.CompressedData)

	if len(delta) > len(newResult.CompressedData) {
		return &api.Mesh{
			ID:               newMesh.ID,
			AnchorID:         newMesh.AnchorID,
			Vertices:         newResult.CompressedData,
			Faces:            []byte{},
			IsDelta:          false,
			CompressionLevel: newMesh.CompressionLevel,
			Timestamp:        newMesh.Timestamp,
		}, nil
	}

	return &api.Mesh{
		ID:               newMesh.ID,
		AnchorID:         newMesh.AnchorID,
		Vertices:         delta,
		Faces:            []byte{},
		IsDelta:          true,
		BaseMeshID:       baseMesh.ID,
		CompressionLevel: newMesh.CompressionLevel,
		Timestamp:        newMesh.Timestamp,
	}, nil
}

func (s *SimpleCompressionService) ApplyMeshDiff(baseMeshData []byte, deltaData []byte) ([]byte, error) {
	if len(baseMeshData) == 0 {
		return deltaData, nil
	}

	result := make([]byte, len(deltaData))
	for i := 0; i < len(deltaData); i++ {
		var baseB byte
		if i < len(baseMeshData) {
			baseB = baseMeshData[i]
		}
		result[i] = deltaData[i] ^ baseB
	}

	return result, nil
}

func quantizeVertices(vertices []float32, bits int) []uint16 {
	maxValInt := (1 << uint(bits)) - 1
	maxVal := float32(maxValInt)
	
	var minX, maxX, minY, maxY, minZ, maxZ float32
	if len(vertices) >= 3 {
		minX, maxX = vertices[0], vertices[0]
		minY, maxY = vertices[1], vertices[1]
		minZ, maxZ = vertices[2], vertices[2]
	}

	for i := 0; i < len(vertices); i += 3 {
		if vertices[i] < minX {
			minX = vertices[i]
		}
		if vertices[i] > maxX {
			maxX = vertices[i]
		}
		if vertices[i+1] < minY {
			minY = vertices[i+1]
		}
		if vertices[i+1] > maxY {
			maxY = vertices[i+1]
		}
		if vertices[i+2] < minZ {
			minZ = vertices[i+2]
		}
		if vertices[i+2] > maxZ {
			maxZ = vertices[i+2]
		}
	}

	rangeX := maxX - minX
	rangeY := maxY - minY
	rangeZ := maxZ - minZ

	quantized := make([]uint16, len(vertices))
	for i := 0; i < len(vertices); i += 3 {
		if rangeX > 0 {
			quantized[i] = uint16((vertices[i] - minX) / rangeX * maxVal)
		}
		if rangeY > 0 {
			quantized[i+1] = uint16((vertices[i+1] - minY) / rangeY * maxVal)
		}
		if rangeZ > 0 {
			quantized[i+2] = uint16((vertices[i+2] - minZ) / rangeZ * maxVal)
		}
	}

	return quantized
}

func dequantizeVertices(quantized []uint16, bits int) []float32 {
	maxValInt := (1 << uint(bits)) - 1
	maxVal := float32(maxValInt)
	
	dequantized := make([]float32, len(quantized))
	for i := 0; i < len(quantized); i++ {
		dequantized[i] = float32(quantized[i]) / maxVal
	}

	return dequantized
}

func bytesToFloat32Array(data []byte) ([]float32, error) {
	if len(data)%4 != 0 {
		return nil, fmt.Errorf("data length must be divisible by 4 for float32")
	}

	result := make([]float32, len(data)/4)
	for i := 0; i < len(result); i++ {
		bits := uint32(data[i*4]) | uint32(data[i*4+1])<<8 | uint32(data[i*4+2])<<16 | uint32(data[i*4+3])<<24
		result[i] = *(*float32)(unsafe.Pointer(&bits))
	}
	return result, nil
}

func bytesToUint32Array(data []byte) ([]uint32, error) {
	if len(data)%4 != 0 {
		return nil, fmt.Errorf("data length must be divisible by 4 for uint32")
	}

	result := make([]uint32, len(data)/4)
	for i := 0; i < len(result); i++ {
		result[i] = uint32(data[i*4]) | uint32(data[i*4+1])<<8 | uint32(data[i*4+2])<<16 | uint32(data[i*4+3])<<24
	}
	return result, nil
}

func computeBinaryDiff(base, new []byte) []byte {
	if len(base) == 0 {
		return new
	}

	var diff []byte
	maxLen := len(base)
	if len(new) > maxLen {
		maxLen = len(new)
	}

	for i := 0; i < maxLen; i++ {
		var baseB, newB byte
		if i < len(base) {
			baseB = base[i]
		}
		if i < len(new) {
			newB = new[i]
		}
		diff = append(diff, newB^baseB)
	}

	return diff
}