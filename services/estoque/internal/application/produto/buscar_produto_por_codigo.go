package produto

import (
	"context"

	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
)

type BuscarProdutoPorCodigoUseCase struct {
	repository domain.Repository
}

func NewBuscarProdutoPorCodigoUseCase(repository domain.Repository) *BuscarProdutoPorCodigoUseCase {
	return &BuscarProdutoPorCodigoUseCase{repository: repository}
}

func (useCase *BuscarProdutoPorCodigoUseCase) Execute(ctx context.Context, codigo string) (*domain.Produto, error) {
	return useCase.repository.BuscarPorCodigo(ctx, codigo)
}
