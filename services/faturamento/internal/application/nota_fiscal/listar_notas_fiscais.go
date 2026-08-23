package notafiscal

import (
	"context"

	domain "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/domain/nota_fiscal"
	sharedquery "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/shared/query"
)

type ListarNotasFiscaisUseCase struct{ repository domain.Repository }

func NewListarNotasFiscaisUseCase(repository domain.Repository) *ListarNotasFiscaisUseCase {
	return &ListarNotasFiscaisUseCase{repository: repository}
}
func (useCase *ListarNotasFiscaisUseCase) Execute(
	ctx context.Context,
	criteria sharedquery.Criteria[domain.ListFilters],
) (sharedquery.Page[domain.NotaFiscal], error) {
	criteria.Pagination = sharedquery.NewPagination(
		criteria.Pagination.Page,
		criteria.Pagination.PageSize,
	)
	return useCase.repository.Listar(ctx, criteria)
}
