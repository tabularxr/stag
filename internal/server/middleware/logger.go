package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tabular/stag/pkg/logger"
)

func Logger(logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		logger.Infof("%s %s - %d - %v", method, path, statusCode, duration)
	}
}