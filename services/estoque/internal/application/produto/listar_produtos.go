package produto

import (
	"context"

	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
	sharedquery "github.com/caiog/korp-notas-fiscais/services/estoque/internal/shared/query"
)

type ListarProdutosUseCase struct {
	repository domain.Repository
}

func NewListarProdutosUseCase(repository domain.Repository) *ListarProdutosUseCase {
	return &ListarProdutosUseCase{repository: repository}
}

func (useCase *ListarProdutosUseCase) Execute(
	ctx context.Context,
	criteria sharedquery.Criteria[domain.ListFilters],
) (sharedquery.Page[domain.Produto], error) {
	criteria.Pagination = sharedquery.NewPagination(
		criteria.Pagination.Page,
		criteria.Pagination.PageSize,
	)
	return useCase.repository.Listar(ctx, criteria)
}
