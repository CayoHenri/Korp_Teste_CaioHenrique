package dependency

import (
	"net/http"

	produtoApplication "github.com/caiog/korp-notas-fiscais/services/estoque/internal/application/produto"
	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/infrastructure/database"
	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/infrastructure/repository"
	httpapi "github.com/caiog/korp-notas-fiscais/services/estoque/internal/presentation/http"
)

type Container struct {
	HTTPHandler http.Handler
}

func NewContainer(connection *database.Connection) *Container {
	produtoRepository := repository.NewProdutoRepository(connection.Gorm)
	produtoHandler := httpapi.NewProdutoHandler(
		produtoApplication.NewCriarProdutoUseCase(produtoRepository),
		produtoApplication.NewListarProdutosUseCase(produtoRepository),
		produtoApplication.NewBuscarProdutoPorIDUseCase(produtoRepository),
		produtoApplication.NewBuscarProdutoPorCodigoUseCase(produtoRepository),
	)

	return &Container{
		HTTPHandler: httpapi.NewRouter(connection.SQL, produtoHandler),
	}
}
