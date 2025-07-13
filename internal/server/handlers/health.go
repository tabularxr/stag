package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tabular/stag/internal/database"
	"github.com/tabular/stag/pkg/api"
	"github.com/tabular/stag/pkg/logger"
)

func HealthHandler(db *database.DB, logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		dbStatus := "healthy"
		
		_, err := db.Database().Info(ctx)
		if err != nil {
			logger.Errorf("Database health check failed: %v", err)
			dbStatus = "unhealthy"
		}

		status := "healthy"
		if dbStatus != "healthy" {
			status = "unhealthy"
		}

		response := api.HealthResponse{
			Status:    status,
			Timestamp: time.Now(),
			Version:   "1.0.0",
			Database:  dbStatus,
		}

		httpStatus := http.StatusOK
		if status != "healthy" {
			httpStatus = http.StatusServiceUnavailable
		}

		c.JSON(httpStatus, response)
	}
}