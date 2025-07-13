package database

import (
	"context"

	"github.com/arangodb/go-driver"
)

const (
	AnchorsCollection = "anchors"
	MeshesCollection  = "meshes"
	TopologyGraph     = "topology"
	TopologyEdges     = "topology_edges"
)

func Migrate(ctx context.Context, db *DB) error {
	database := db.Database()

	if err := createCollection(ctx, database, AnchorsCollection, false); err != nil {
		return err
	}

	if err := createCollection(ctx, database, MeshesCollection, false); err != nil {
		return err
	}

	if err := createCollection(ctx, database, TopologyEdges, true); err != nil {
		return err
	}

	if err := createGraph(ctx, database); err != nil {
		return err
	}

	if err := createIndexes(ctx, database); err != nil {
		return err
	}

	return nil
}

func createCollection(ctx context.Context, db driver.Database, name string, isEdge bool) error {
	exists, err := db.CollectionExists(ctx, name)
	if err != nil {
		return err
	}

	if !exists {
		options := &driver.CreateCollectionOptions{}
		if isEdge {
			options.Type = driver.CollectionTypeEdge
		}

		_, err = db.CreateCollection(ctx, name, options)
		if err != nil {
			return err
		}
	}

	return nil
}

func createGraph(ctx context.Context, db driver.Database) error {
	exists, err := db.GraphExists(ctx, TopologyGraph)
	if err != nil {
		return err
	}

	if !exists {
		edgeDefinitions := []driver.EdgeDefinition{
			{
				Collection: TopologyEdges,
				From:       []string{AnchorsCollection},
				To:         []string{AnchorsCollection},
			},
		}

		_, err = db.CreateGraph(ctx, TopologyGraph, &driver.CreateGraphOptions{
			EdgeDefinitions: edgeDefinitions,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func createIndexes(ctx context.Context, db driver.Database) error {
	anchorsCol, err := db.Collection(ctx, AnchorsCollection)
	if err != nil {
		return err
	}

	_, _, err = anchorsCol.EnsurePersistentIndex(ctx, []string{"session_id"}, &driver.EnsurePersistentIndexOptions{
		Name: "session_id_idx",
	})
	if err != nil {
		return err
	}

	_, _, err = anchorsCol.EnsurePersistentIndex(ctx, []string{"timestamp"}, &driver.EnsurePersistentIndexOptions{
		Name: "timestamp_idx",
	})
	if err != nil {
		return err
	}

	meshesCol, err := db.Collection(ctx, MeshesCollection)
	if err != nil {
		return err
	}

	_, _, err = meshesCol.EnsurePersistentIndex(ctx, []string{"anchor_id"}, &driver.EnsurePersistentIndexOptions{
		Name: "anchor_id_idx",
	})
	if err != nil {
		return err
	}

	_, _, err = meshesCol.EnsurePersistentIndex(ctx, []string{"timestamp"}, &driver.EnsurePersistentIndexOptions{
		Name: "mesh_timestamp_idx",
	})
	if err != nil {
		return err
	}

	return nil
}