package dependency

import (
	"net/http"
	"time"

	notafiscalApplication "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/application/nota_fiscal"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/infrastructure/client"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/infrastructure/config"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/infrastructure/database"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/infrastructure/repository"
	httpPresentation "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/presentation/http"
)

type Container struct {
	HTTPHandler http.Handler
}

func NewContainer(connection *database.Connection, cfg config.Config) *Container {
	notaRepository := repository.NewNotaFiscalRepository(connection.Gorm)
	estoqueClient := client.NewEstoqueClient(
		cfg.EstoqueBaseURL,
		&http.Client{Timeout: 5 * time.Second},
	)
	handler := httpPresentation.NewNotaFiscalHandler(
		notafiscalApplication.NewCriarNotaFiscalUseCase(notaRepository, estoqueClient),
		notafiscalApplication.NewBuscarNotaFiscalUseCase(notaRepository),
		notafiscalApplication.NewListarNotasFiscaisUseCase(notaRepository),
		notafiscalApplication.NewIniciarFechamentoUseCase(notaRepository),
	)
	return &Container{
		HTTPHandler: httpPresentation.NewRouter(connection.SQL, handler),
	}
}
