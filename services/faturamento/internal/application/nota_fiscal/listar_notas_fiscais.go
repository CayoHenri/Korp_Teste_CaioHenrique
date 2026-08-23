package notafiscal

import (
	"context"
	domain "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/domain/nota_fiscal"
)

type ListarNotasFiscaisUseCase struct{ repository domain.Repository }

func NewListarNotasFiscaisUseCase(repository domain.Repository) *ListarNotasFiscaisUseCase {
	return &ListarNotasFiscaisUseCase{repository: repository}
}
func (useCase *ListarNotasFiscaisUseCase) Execute(ctx context.Context) ([]domain.NotaFiscal, error) {
	return useCase.repository.Listar(ctx)
}
