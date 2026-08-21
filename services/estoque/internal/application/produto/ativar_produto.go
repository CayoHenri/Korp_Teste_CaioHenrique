package produto

import (
	"context"

	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
	"github.com/google/uuid"
)

type AtivarProdutoUseCase struct {
	repository domain.Repository
}

func NewAtivarProdutoUseCase(repository domain.Repository) *AtivarProdutoUseCase {
	return &AtivarProdutoUseCase{repository: repository}
}

func (useCase *AtivarProdutoUseCase) Execute(ctx context.Context, id uuid.UUID) (*domain.Produto, error) {
	produto, err := useCase.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, err
	}

	produto.Ativar()
	if err := useCase.repository.Atualizar(ctx, produto); err != nil {
		return nil, err
	}
	return produto, nil
}
