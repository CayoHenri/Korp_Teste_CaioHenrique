package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/caiog/korp-notas-fiscais/services/estoque/docs"
	produtoApplication "github.com/caiog/korp-notas-fiscais/services/estoque/internal/application/produto"
	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/infrastructure/config"
	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/infrastructure/database"
	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/infrastructure/repository"
	httpapi "github.com/caiog/korp-notas-fiscais/services/estoque/internal/presentation/http"
)

// @title Estoque Service API
// @version 1.0
// @description API do contexto de Estoque do Sistema de Emissao de Notas Fiscais.
// @BasePath /
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("configuracao invalida", "error", err)
		os.Exit(1)
	}

	connection, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		logger.Error("nao foi possivel conectar ao PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer connection.Close()
	produtoRepository := repository.NewGormProdutoRepository(connection.Gorm)
	produtoService := produtoApplication.NewService(produtoRepository)
	produtoHandler := httpapi.NewProdutoHandler(produtoService)

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           httpapi.NewRouter(connection.SQL, produtoHandler),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("estoque-service iniciado", "port", cfg.HTTPPort)
		serverErrors <- server.ListenAndServe()
	}()

	shutdownSignal, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	select {
	case <-shutdownSignal.Done():
		logger.Info("encerrando estoque-service")
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("servidor HTTP encerrado inesperadamente", "error", err)
			os.Exit(1)
		}
	}

	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownContext); err != nil {
		logger.Error("falha ao encerrar servidor HTTP", "error", err)
		os.Exit(1)
	}
}
