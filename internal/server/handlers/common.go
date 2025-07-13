package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/tabular/stag/pkg/errors"
	"github.com/tabular/stag/pkg/logger"
)

func handleError(c *gin.Context, err error, logger logger.Logger) {
	if apiErr, ok := err.(*errors.APIError); ok {
		logger.Errorf("API error: %v", apiErr)
		c.JSON(apiErr.Code, gin.H{
			"error":   apiErr.Message,
			"details": apiErr.Details,
		})
		return
	}

	logger.Errorf("Internal error: %v", err)
	c.JSON(500, gin.H{
		"error": "Internal server error",
	})
}