package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tabular/stag/internal/spatial"
	"github.com/tabular/stag/pkg/api"
	"github.com/tabular/stag/pkg/errors"
	"github.com/tabular/stag/pkg/logger"
)

func IngestHandler(repo *spatial.Repository, wsHub *WSHub, logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var event api.SpatialEvent
		if err := c.ShouldBindJSON(&event); err != nil {
			handleError(c, errors.ValidationError(err.Error()), logger)
			return
		}

		if event.EventID == "" {
			event.EventID = uuid.New().String()
		}

		if event.Timestamp == 0 {
			event.Timestamp = time.Now().UnixMilli()
		}

		for i := range event.Anchors {
			if event.Anchors[i].ID == "" {
				event.Anchors[i].ID = uuid.New().String()
			}
			if event.Anchors[i].SessionID == "" {
				event.Anchors[i].SessionID = event.SessionID
			}
			if event.Anchors[i].Timestamp == 0 {
				event.Anchors[i].Timestamp = event.Timestamp
			}
		}

		for i := range event.Meshes {
			if event.Meshes[i].ID == "" {
				event.Meshes[i].ID = uuid.New().String()
			}
			if event.Meshes[i].Timestamp == 0 {
				event.Meshes[i].Timestamp = event.Timestamp
			}
		}

		if err := repo.IngestEvent(c.Request.Context(), &event); err != nil {
			handleError(c, err, logger)
			return
		}

		wsHub.BroadcastToSession(event.SessionID, api.WSMessage{
			Type:      api.WSMessageTypeAnchorUpdate,
			SessionID: event.SessionID,
			Data:      event,
			Timestamp: time.Now().UnixMilli(),
		})

		response := api.IngestResponse{
			Success:   true,
			EventID:   event.EventID,
			Processed: len(event.Anchors) + len(event.Meshes),
		}

		c.JSON(http.StatusOK, response)
	}
}