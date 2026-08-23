package worker

import (
	"context"
	"log/slog"
	"time"

	application "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/application/outbox"
)

type OutboxWorker struct {
	useCase  *application.PublicarEventosUseCase
	interval time.Duration
	logger   *slog.Logger
}

func NewOutboxWorker(
	useCase *application.PublicarEventosUseCase,
	interval time.Duration,
	logger *slog.Logger,
) *OutboxWorker {
	return &OutboxWorker{useCase: useCase, interval: interval, logger: logger}
}

func (worker *OutboxWorker) Run(ctx context.Context) {
	worker.publish(ctx)
	ticker := time.NewTicker(worker.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			worker.publish(ctx)
		}
	}
}

func (worker *OutboxWorker) publish(ctx context.Context) {
	published, err := worker.useCase.Execute(ctx)
	if err != nil {
		worker.logger.Error("falha ao publicar Outbox", "error", err, "published", published)
		return
	}
	if published > 0 {
		worker.logger.Info("eventos da Outbox publicados", "count", published)
	}
}
