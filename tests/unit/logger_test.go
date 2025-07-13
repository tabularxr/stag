package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tabular/stag/pkg/logger"
)

func TestLogger_New(t *testing.T) {
	tests := []struct {
		level    string
		expected string
	}{
		{"debug", "debug"},
		{"info", "info"},
		{"warn", "warn"},
		{"error", "error"},
		{"invalid", "info"}, // Should default to info for invalid levels
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			logger := logger.New(tt.level)
			assert.NotNil(t, logger)
			
			// Test that logger methods don't panic
			logger.Info("test message")
			logger.Debug("debug message")
			logger.Error("error message")
		})
	}
}