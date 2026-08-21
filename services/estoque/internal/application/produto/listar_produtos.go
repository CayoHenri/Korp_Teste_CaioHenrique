package produto

import (
	"context"

	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
)

type ListarProdutosUseCase struct {
	repository domain.Repository
}

func NewListarProdutosUseCase(repository domain.Repository) *ListarProdutosUseCase {
	return &ListarProdutosUseCase{repository: repository}
}

func (useCase *ListarProdutosUseCase) Execute(ctx context.Context) ([]domain.Produto, error) {
	return useCase.repository.Listar(ctx)
}
