package worker

import (
	"context"
	"errors"
	"log/slog"

	application "github.com/caiog/korp-notas-fiscais/services/estoque/internal/application/estoque"
	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/infrastructure/messaging"
)

type BaixaEstoqueWorker struct {
	rabbitMQ *messaging.RabbitMQ
	useCase  *application.ProcessarBaixaSolicitadaUseCase
	logger   *slog.Logger
}

func NewBaixaEstoqueWorker(
	rabbitMQ *messaging.RabbitMQ,
	useCase *application.ProcessarBaixaSolicitadaUseCase,
	logger *slog.Logger,
) *BaixaEstoqueWorker {
	return &BaixaEstoqueWorker{rabbitMQ: rabbitMQ, useCase: useCase, logger: logger}
}

func (worker *BaixaEstoqueWorker) Run(ctx context.Context) {
	worker.logger.Info("consumidor de baixas iniciado")
	if err := worker.rabbitMQ.ConsumirBaixas(ctx, worker.useCase); err != nil &&
		!errors.Is(err, context.Canceled) {
		worker.logger.Error("consumidor de baixas encerrado", "error", err)
	}
}
