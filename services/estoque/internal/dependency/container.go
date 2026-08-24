package dependency

import (
	"fmt"
	"log/slog"
	"net/http"

	estoqueApplication "github.com/caiog/korp-notas-fiscais/services/estoque/internal/application/estoque"
	produtoApplication "github.com/caiog/korp-notas-fiscais/services/estoque/internal/application/produto"
	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/infrastructure/config"
	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/infrastructure/database"
	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/infrastructure/messaging"
	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/infrastructure/repository"
	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/infrastructure/worker"
	httpapi "github.com/caiog/korp-notas-fiscais/services/estoque/internal/presentation/http"
)

type Container struct {
	HTTPHandler        http.Handler
	BaixaEstoqueWorker *worker.BaixaEstoqueWorker
	rabbitMQ           *messaging.RabbitMQ
}

func NewContainer(
	connection *database.Connection,
	cfg config.Config,
	logger *slog.Logger,
) (*Container, error) {
	produtoRepository := repository.NewProdutoRepository(connection.Gorm)
	rabbitMQ, err := messaging.NewRabbitMQ(
		cfg.RabbitMQURL,
		cfg.RabbitMQRecoveryMaxRetries,
		cfg.RabbitMQRecoveryInterval,
		cfg.RabbitMQMessageTimeout,
		cfg.RabbitMQMessageMaxRetries,
		cfg.RabbitMQMessageRetryDelay,
	)
	if err != nil {
		return nil, fmt.Errorf("inicializar RabbitMQ: %w", err)
	}
	baixarEstoque := estoqueApplication.NewBaixarEstoqueUseCase(produtoRepository)
	processarBaixa := estoqueApplication.NewProcessarBaixaSolicitadaUseCase(
		baixarEstoque,
		rabbitMQ,
	)
	produtoHandler := httpapi.NewProdutoHandler(
		produtoApplication.NewCriarProdutoUseCase(produtoRepository),
		produtoApplication.NewListarProdutosUseCase(produtoRepository),
		produtoApplication.NewBuscarProdutoPorIDUseCase(produtoRepository),
		produtoApplication.NewBuscarProdutoPorCodigoUseCase(produtoRepository),
		produtoApplication.NewAtivarProdutoUseCase(produtoRepository),
		produtoApplication.NewInativarProdutoUseCase(produtoRepository),
		produtoApplication.NewAtualizarProdutoUseCase(produtoRepository),
		estoqueApplication.NewListarMovimentacoesUseCase(produtoRepository),
	)

	return &Container{
		HTTPHandler:        httpapi.NewRouter(connection.SQL, produtoHandler, cfg.CORSAllowedOrigins...),
		BaixaEstoqueWorker: worker.NewBaixaEstoqueWorker(rabbitMQ, processarBaixa, logger),
		rabbitMQ:           rabbitMQ,
	}, nil
}

func (container *Container) Close() error {
	return container.rabbitMQ.Close()
}
