package estoque

import (
	"context"

	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/movimentacao"
	"github.com/google/uuid"
)

type movimentacaoRepository interface {
	ListarMovimentacoes(context.Context, uuid.UUID) ([]movimentacao.Movimentacao, error)
}

type ListarMovimentacoesUseCase struct {
	repository movimentacaoRepository
}

func NewListarMovimentacoesUseCase(repository movimentacaoRepository) *ListarMovimentacoesUseCase {
	return &ListarMovimentacoesUseCase{repository: repository}
}

func (useCase *ListarMovimentacoesUseCase) Execute(
	ctx context.Context,
	produtoID uuid.UUID,
) ([]movimentacao.Movimentacao, error) {
	return useCase.repository.ListarMovimentacoes(ctx, produtoID)
}
