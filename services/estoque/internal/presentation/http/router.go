package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type databasePinger interface {
	PingContext(context.Context) error
}

func NewRouter(database databasePinger) http.Handler {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/health", healthHandler(database))

	return router
}

// healthHandler godoc
// @Summary Verifica a saude do servico
// @Description Verifica se a API e a conexao com PostgreSQL estao disponiveis.
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /health [get]
func healthHandler(database databasePinger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := database.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "unhealthy",
				"database": "unavailable",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "healthy",
			"service":  "estoque-service",
			"database": "available",
		})
	}
}
