package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tabular/stag/internal/spatial"
	"github.com/tabular/stag/pkg/api"
	"github.com/tabular/stag/pkg/errors"
	"github.com/tabular/stag/pkg/logger"
)

func QueryHandler(repo *spatial.Repository, logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var params api.QueryParams
		if err := c.ShouldBindQuery(&params); err != nil {
			handleError(c, errors.ValidationError(err.Error()), logger)
			return
		}

		if params.Limit <= 0 {
			params.Limit = 100
		}
		if params.Limit > 1000 {
			params.Limit = 1000
		}

		response, err := repo.QueryWithDecompression(c.Request.Context(), &params, params.Decompress)
		if err != nil {
			handleError(c, err, logger)
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

func GetAnchorHandler(repo *spatial.Repository, logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		anchorID := c.Param("id")
		if anchorID == "" {
			handleError(c, errors.BadRequest("anchor ID is required"), logger)
			return
		}

		anchor, err := repo.GetAnchor(c.Request.Context(), anchorID)
		if err != nil {
			handleError(c, err, logger)
			return
		}

		c.JSON(http.StatusOK, anchor)
	}
}