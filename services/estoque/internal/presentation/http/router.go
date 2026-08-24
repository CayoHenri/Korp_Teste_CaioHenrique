package http

import (
	"context"
	"net/http"
	"time"

	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/presentation/http/response"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type databasePinger interface {
	PingContext(context.Context) error
}

func NewRouter(
	database databasePinger,
	produtoHandler *ProdutoHandler,
	allowedOrigins ...string,
) http.Handler {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), corsMiddleware(allowedOrigins))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/health", healthHandler(database))
	if produtoHandler != nil {
		produtoHandler.RegisterRoutes(router)
	}

	return router
}

// healthHandler godoc
// @Summary Verifica a saude do servico
// @Description Verifica se a API e a conexao com PostgreSQL estao disponiveis.
// @Tags Health
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Failure 503 {object} response.ErrorResponse
// @Router /health [get]
func healthHandler(database databasePinger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := database.PingContext(ctx); err != nil {
			response.Error(c, http.StatusServiceUnavailable, "SERVICO_INDISPONIVEL", "banco de dados indisponivel")
			return
		}

		response.OK(c, gin.H{
			"status":   "healthy",
			"service":  "estoque-service",
			"database": "available",
		})
	}
}
