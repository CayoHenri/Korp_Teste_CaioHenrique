package notafiscal

import (
	"context"
	domain "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/domain/nota_fiscal"
	"github.com/google/uuid"
)

type BuscarNotaFiscalUseCase struct{ repository domain.Repository }

func NewBuscarNotaFiscalUseCase(repository domain.Repository) *BuscarNotaFiscalUseCase {
	return &BuscarNotaFiscalUseCase{repository: repository}
}
func (useCase *BuscarNotaFiscalUseCase) Execute(
	ctx context.Context,
	id uuid.UUID,
) (*domain.NotaFiscal, error) {
	return useCase.repository.BuscarPorID(ctx, id)
}
