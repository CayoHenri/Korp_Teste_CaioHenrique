package worker

import (
	"context"
	"errors"
	"log/slog"

	application "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/application/nota_fiscal"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/infrastructure/messaging"
)

type ResultadoBaixaWorker struct {
	rabbitMQ *messaging.RabbitMQPublisher
	useCase  *application.ProcessarResultadoBaixaUseCase
	logger   *slog.Logger
}

func NewResultadoBaixaWorker(
	rabbitMQ *messaging.RabbitMQPublisher,
	useCase *application.ProcessarResultadoBaixaUseCase,
	logger *slog.Logger,
) *ResultadoBaixaWorker {
	return &ResultadoBaixaWorker{rabbitMQ: rabbitMQ, useCase: useCase, logger: logger}
}

func (worker *ResultadoBaixaWorker) Run(ctx context.Context) {
	worker.logger.Info("consumidor de resultados de baixa iniciado")
	err := worker.rabbitMQ.ConsumirResultados(ctx, worker.useCase)
	if err != nil && !errors.Is(err, context.Canceled) {
		worker.logger.Error("consumidor de resultados encerrado", "error", err)
	}
}
