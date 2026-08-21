package dependency

import (
	"net/http"

	estoqueApplication "github.com/caiog/korp-notas-fiscais/services/estoque/internal/application/estoque"
	produtoApplication "github.com/caiog/korp-notas-fiscais/services/estoque/internal/application/produto"
	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/infrastructure/database"
	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/infrastructure/repository"
	httpapi "github.com/caiog/korp-notas-fiscais/services/estoque/internal/presentation/http"
)

type Container struct {
	HTTPHandler   http.Handler
	BaixarEstoque *estoqueApplication.BaixarEstoqueUseCase
}

func NewContainer(connection *database.Connection) *Container {
	produtoRepository := repository.NewProdutoRepository(connection.Gorm)
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
		HTTPHandler:   httpapi.NewRouter(connection.SQL, produtoHandler),
		BaixarEstoque: estoqueApplication.NewBaixarEstoqueUseCase(produtoRepository),
	}
}
