package produto

import (
	"context"

	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
)

type CriarProdutoInput struct {
	Codigo    string
	Descricao string
	Saldo     int
}

type CriarProdutoUseCase struct {
	repository domain.Repository
}

func NewCriarProdutoUseCase(repository domain.Repository) *CriarProdutoUseCase {
	return &CriarProdutoUseCase{repository: repository}
}

func (useCase *CriarProdutoUseCase) Execute(ctx context.Context, input CriarProdutoInput) (*domain.Produto, error) {
	produto, err := domain.NewProduto(input.Codigo, input.Descricao, input.Saldo)
	if err != nil {
		return nil, err
	}

	if err := useCase.repository.Criar(ctx, produto); err != nil {
		return nil, err
	}

	return produto, nil
}
