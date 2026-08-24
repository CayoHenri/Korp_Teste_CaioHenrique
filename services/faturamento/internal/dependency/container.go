package dependency

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	notafiscalApplication "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/application/nota_fiscal"
	outboxApplication "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/application/outbox"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/infrastructure/client"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/infrastructure/config"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/infrastructure/database"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/infrastructure/messaging"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/infrastructure/repository"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/infrastructure/worker"
	httpPresentation "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/presentation/http"
)

type Container struct {
	HTTPHandler          http.Handler
	OutboxWorker         *worker.OutboxWorker
	ResultadoBaixaWorker *worker.ResultadoBaixaWorker
	publisher            *messaging.RabbitMQPublisher
}

func NewContainer(
	connection *database.Connection,
	cfg config.Config,
	logger *slog.Logger,
) (*Container, error) {
	notaRepository := repository.NewNotaFiscalRepository(connection.Gorm)
	outboxRepository := repository.NewOutboxRepository(connection.Gorm)
	publisher, err := messaging.NewRabbitMQPublisher(
		cfg.RabbitMQURL,
		cfg.RabbitMQRecoveryMaxRetries,
		cfg.RabbitMQRecoveryInterval,
		cfg.RabbitMQMessageTimeout,
		cfg.RabbitMQMessageMaxRetries,
		cfg.RabbitMQMessageRetryDelay,
	)
	if err != nil {
		return nil, fmt.Errorf("inicializar publisher RabbitMQ: %w", err)
	}
	estoqueClient := client.NewEstoqueClient(
		cfg.EstoqueBaseURL,
		&http.Client{Timeout: 5 * time.Second},
	)
	handler := httpPresentation.NewNotaFiscalHandler(
		notafiscalApplication.NewCriarNotaFiscalUseCase(notaRepository, estoqueClient),
		notafiscalApplication.NewAtualizarNotaFiscalUseCase(notaRepository, estoqueClient),
		notafiscalApplication.NewBuscarNotaFiscalUseCase(notaRepository),
		notafiscalApplication.NewListarNotasFiscaisUseCase(notaRepository),
		notafiscalApplication.NewIniciarFechamentoUseCase(notaRepository),
	)
	publicarEventos := outboxApplication.NewPublicarEventosUseCase(
		outboxRepository,
		publisher,
		100,
	)
	processarResultado := notafiscalApplication.NewProcessarResultadoBaixaUseCase(
		notaRepository,
	)
	return &Container{
		HTTPHandler:          httpPresentation.NewRouter(connection.SQL, handler, cfg.CORSAllowedOrigins...),
		OutboxWorker:         worker.NewOutboxWorker(publicarEventos, cfg.OutboxInterval, logger),
		ResultadoBaixaWorker: worker.NewResultadoBaixaWorker(publisher, processarResultado, logger),
		publisher:            publisher,
	}, nil
}

func (container *Container) Close() error {
	return container.publisher.Close()
}
