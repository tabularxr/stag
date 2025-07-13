package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tabular/stag/internal/metrics"
)

func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		c.Next()
		
		duration := time.Since(start).Seconds()
		method := c.Request.Method
		path := c.FullPath()
		status := strconv.Itoa(c.Writer.Status())
		
		metrics.RequestDuration.WithLabelValues(method, path, status).Observe(duration)
		metrics.RequestsTotal.WithLabelValues(method, path, status).Inc()
	}
}