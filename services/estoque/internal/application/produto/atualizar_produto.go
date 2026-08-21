package produto

import (
	"context"

	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
	"github.com/google/uuid"
)

type AtualizarProdutoInput struct {
	ID        uuid.UUID
	Descricao string
	Saldo     int
}

type AtualizarProdutoUseCase struct {
	repository domain.Repository
}

func NewAtualizarProdutoUseCase(repository domain.Repository) *AtualizarProdutoUseCase {
	return &AtualizarProdutoUseCase{repository: repository}
}

func (useCase *AtualizarProdutoUseCase) Execute(ctx context.Context, input AtualizarProdutoInput) (*domain.Produto, error) {
	produto, err := useCase.repository.BuscarPorID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if err := produto.Atualizar(input.Descricao, input.Saldo); err != nil {
		return nil, err
	}
	if err := useCase.repository.Atualizar(ctx, produto); err != nil {
		return nil, err
	}
	return produto, nil
}
