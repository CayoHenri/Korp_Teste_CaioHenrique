package produto

import (
	"context"

	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
	"github.com/google/uuid"
)

type InativarProdutoUseCase struct {
	repository domain.Repository
}

func NewInativarProdutoUseCase(repository domain.Repository) *InativarProdutoUseCase {
	return &InativarProdutoUseCase{repository: repository}
}

func (useCase *InativarProdutoUseCase) Execute(ctx context.Context, id uuid.UUID) (*domain.Produto, error) {
	produto, err := useCase.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, err
	}

	produto.Inativar()
	if err := useCase.repository.Atualizar(ctx, produto); err != nil {
		return nil, err
	}
	return produto, nil
}
