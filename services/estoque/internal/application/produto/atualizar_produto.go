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
	Valor     float64
}

type AtualizarProdutoUseCase struct {
	repository domain.Repository
}

func NewAtualizarProdutoUseCase(repository domain.Repository) *AtualizarProdutoUseCase {
	return &AtualizarProdutoUseCase{repository: repository}
}

func (useCase *AtualizarProdutoUseCase) Execute(
	ctx context.Context,
	input AtualizarProdutoInput,
) (*domain.Produto, error) {
	produto, err := useCase.repository.BuscarPorID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if err := produto.AtualizarDescricao(input.Descricao); err != nil {
		return nil, err
	}
	if err := produto.AtualizarSaldo(input.Saldo); err != nil {
		return nil, err
	}
	if input.Valor != 0 {
		if err := produto.AtualizarValor(input.Valor); err != nil {
			return nil, err
		}
	}
	if err := useCase.repository.Atualizar(ctx, produto); err != nil {
		return nil, err
	}
	return produto, nil
}
