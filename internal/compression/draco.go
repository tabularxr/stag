package compression

import (
	"github.com/tabular/stag/pkg/api"
)

type DracoService struct {
	simpleService *SimpleCompressionService
}

func NewDracoService() *DracoService {
	return &DracoService{
		simpleService: NewSimpleCompressionService(),
	}
}

func (d *DracoService) CompressMesh(vertices []float32, faces []uint32, compressionLevel int) (*CompressionResult, error) {
	return d.simpleService.CompressMesh(vertices, faces, compressionLevel)
}

func (d *DracoService) DecompressMesh(compressedData []byte) (*DecompressionResult, error) {
	return d.simpleService.DecompressMesh(compressedData)
}

func (d *DracoService) CompressMeshFromAPI(mesh *api.Mesh) (*CompressionResult, error) {
	return d.simpleService.CompressMeshFromAPI(mesh)
}

func (d *DracoService) ComputeMeshDiff(baseMesh, newMesh *api.Mesh) (*api.Mesh, error) {
	return d.simpleService.ComputeMeshDiff(baseMesh, newMesh)
}

func (d *DracoService) ApplyMeshDiff(baseMeshData []byte, deltaData []byte) ([]byte, error) {
	return d.simpleService.ApplyMeshDiff(baseMeshData, deltaData)
}