package notafiscal

import (
	"context"

	domain "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/domain/nota_fiscal"
	"github.com/google/uuid"
)

type AtualizarNotaFiscalInput struct {
	ID              uuid.UUID
	NomeCliente     string
	EnderecoCliente string
	Itens           []CriarNotaFiscalItemInput
}

type AtualizarNotaFiscalUseCase struct {
	repository domain.Repository
	produtos   produtoCatalogo
}

func NewAtualizarNotaFiscalUseCase(
	repository domain.Repository,
	produtos produtoCatalogo,
) *AtualizarNotaFiscalUseCase {
	return &AtualizarNotaFiscalUseCase{repository: repository, produtos: produtos}
}

func (useCase *AtualizarNotaFiscalUseCase) Execute(
	ctx context.Context,
	input AtualizarNotaFiscalInput,
) (*domain.NotaFiscal, error) {
	nota, err := useCase.repository.BuscarPorID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if err := nota.ValidarEdicao(); err != nil {
		return nil, err
	}
	itens, err := montarItens(ctx, useCase.produtos, input.Itens)
	if err != nil {
		return nil, err
	}
	if err := nota.Atualizar(input.NomeCliente, input.EnderecoCliente, itens); err != nil {
		return nil, err
	}
	if err := useCase.repository.Atualizar(ctx, nota); err != nil {
		return nil, err
	}
	return nota, nil
}
