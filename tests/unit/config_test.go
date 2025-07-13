package unit

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tabular/stag/internal/config"
)

func TestConfig_Load(t *testing.T) {
	// Set required environment variable
	os.Setenv("ARANGO_PASSWORD", "testpass")
	defer os.Unsetenv("ARANGO_PASSWORD")

	cfg, err := config.Load()
	require.NoError(t, err)
	
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, "http://localhost:8529", cfg.Database.URL)
	assert.Equal(t, "stag", cfg.Database.Database)
	assert.Equal(t, "root", cfg.Database.Username)
	assert.Equal(t, "testpass", cfg.Database.Password)
	assert.Equal(t, "info", cfg.LogLevel)
}

func TestConfig_LoadWithCustomValues(t *testing.T) {
	// Set environment variables
	os.Setenv("ARANGO_PASSWORD", "custompass")
	os.Setenv("STAG_PORT", "9090")
	os.Setenv("LOG_LEVEL", "debug")
	
	defer func() {
		os.Unsetenv("ARANGO_PASSWORD")
		os.Unsetenv("STAG_PORT")
		os.Unsetenv("LOG_LEVEL")
	}()

	cfg, err := config.Load()
	require.NoError(t, err)
	
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "custompass", cfg.Database.Password)
}

func TestConfig_LoadMissingPassword(t *testing.T) {
	// Ensure password is not set
	os.Unsetenv("ARANGO_PASSWORD")
	
	_, err := config.Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ARANGO_PASSWORD environment variable is required")
}