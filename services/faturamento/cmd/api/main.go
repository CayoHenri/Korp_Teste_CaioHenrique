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

	_ "github.com/caiog/korp-notas-fiscais/services/faturamento/docs"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/dependency"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/infrastructure/config"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/infrastructure/database"
)

// @title Faturamento Service API
// @version 1.0
// @description API do contexto de Faturamento do Sistema de Emissao de Notas Fiscais.
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
	container, err := dependency.NewContainer(connection, cfg, logger)
	if err != nil {
		logger.Error("nao foi possivel montar as dependencias", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := container.Close(); err != nil {
			logger.Error("falha ao encerrar RabbitMQ", "error", err)
		}
	}()
	shutdownSignal, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()
	go container.OutboxWorker.Run(shutdownSignal)
	go container.ResultadoBaixaWorker.Run(shutdownSignal)
	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           container.HTTPHandler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("faturamento-service iniciado", "port", cfg.HTTPPort)
		serverErrors <- server.ListenAndServe()
	}()
	select {
	case <-shutdownSignal.Done():
		logger.Info("encerrando faturamento-service")
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("servidor HTTP encerrado inesperadamente", "error", err)
			os.Exit(1)
		}
	}
	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		logger.Error("falha ao encerrar servidor HTTP", "error", err)
		os.Exit(1)
	}
}
