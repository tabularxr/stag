package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tabular/stag/internal/config"
	"github.com/tabular/stag/internal/database"
	"github.com/tabular/stag/internal/server/handlers"
	"github.com/tabular/stag/internal/server/middleware"
	"github.com/tabular/stag/internal/spatial"
	"github.com/tabular/stag/pkg/logger"
)

type Server struct {
	config     *config.Config
	db         *database.DB
	logger     logger.Logger
	repository *spatial.Repository
	wsHub      *handlers.WSHub
}

func New(cfg *config.Config, db *database.DB, logger logger.Logger) *Server {
	repository := spatial.NewRepository(db)
	wsHub := handlers.NewWSHub(logger)
	
	return &Server{
		config:     cfg,
		db:         db,
		logger:     logger,
		repository: repository,
		wsHub:      wsHub,
	}
}

func (s *Server) Router() http.Handler {
	if s.config.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(s.logger))
	router.Use(middleware.CORS())
	router.Use(middleware.Metrics())

	api := router.Group("/api/v1")
	{
		api.POST("/ingest", handlers.IngestHandler(s.repository, s.wsHub, s.logger))
		api.GET("/query", handlers.QueryHandler(s.repository, s.logger))
		api.GET("/anchors/:id", handlers.GetAnchorHandler(s.repository, s.logger))
		api.GET("/health", handlers.HealthHandler(s.db, s.logger))
	}

	router.GET("/stream/:session_id", handlers.WSHandler(s.wsHub, s.logger))
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return router
}