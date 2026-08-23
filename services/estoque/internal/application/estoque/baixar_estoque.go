package estoque

import (
	"context"

	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
	"github.com/google/uuid"
)

type BaixarEstoqueItemInput struct {
	ProdutoID  uuid.UUID
	Quantidade int
}

type BaixarEstoqueInput struct {
	EventID      uuid.UUID
	NotaFiscalID uuid.UUID
	Itens        []BaixarEstoqueItemInput
}

type BaixarEstoqueOutput struct {
	Duplicada bool
}

type baixaRepository interface {
	BaixarEstoque(context.Context, domain.BaixaEstoque) (bool, error)
}

type BaixarEstoqueUseCase struct {
	repository baixaRepository
}

func NewBaixarEstoqueUseCase(repository baixaRepository) *BaixarEstoqueUseCase {
	return &BaixarEstoqueUseCase{repository: repository}
}

func (useCase *BaixarEstoqueUseCase) Execute(
	ctx context.Context,
	input BaixarEstoqueInput,
) (BaixarEstoqueOutput, error) {
	itens := make([]domain.BaixaItem, 0, len(input.Itens))
	for _, item := range input.Itens {
		baixaItem, err := domain.NewBaixaItem(item.ProdutoID, item.Quantidade)
		if err != nil {
			return BaixarEstoqueOutput{}, err
		}
		itens = append(itens, *baixaItem)
	}
	baixa, err := domain.NewBaixaEstoque(input.EventID, input.NotaFiscalID, itens)
	if err != nil {
		return BaixarEstoqueOutput{}, err
	}

	processada, err := useCase.repository.BaixarEstoque(ctx, *baixa)
	if err != nil {
		return BaixarEstoqueOutput{}, err
	}
	return BaixarEstoqueOutput{Duplicada: !processada}, nil
}
