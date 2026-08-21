package produto

import (
	"context"

	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
	"github.com/google/uuid"
)

type BuscarProdutoPorIDUseCase struct {
	repository domain.Repository
}

func NewBuscarProdutoPorIDUseCase(repository domain.Repository) *BuscarProdutoPorIDUseCase {
	return &BuscarProdutoPorIDUseCase{repository: repository}
}

func (useCase *BuscarProdutoPorIDUseCase) Execute(ctx context.Context, id uuid.UUID) (*domain.Produto, error) {
	return useCase.repository.BuscarPorID(ctx, id)
}
