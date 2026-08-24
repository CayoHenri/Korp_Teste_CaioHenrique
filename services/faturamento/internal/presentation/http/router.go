package http

import (
	"context"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/presentation/http/response"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	nethttp "net/http"
	"time"
)

type databasePinger interface {
	PingContext(context.Context) error
}

func NewRouter(
	database databasePinger,
	handler *NotaFiscalHandler,
	allowedOrigins ...string,
) nethttp.Handler {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), corsMiddleware(allowedOrigins))
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/health", healthHandler(database))
	handler.RegisterRoutes(router)
	return router
}
func healthHandler(database databasePinger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := database.PingContext(ctx); err != nil {
			response.Error(c, nethttp.StatusServiceUnavailable, "SERVICO_INDISPONIVEL", "banco de dados indisponivel")
			return
		}
		response.OK(c, gin.H{
			"status":   "healthy",
			"service":  "faturamento-service",
			"database": "available",
		})
	}
}
