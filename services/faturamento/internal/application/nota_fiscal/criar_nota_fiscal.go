package notafiscal

import (
	"context"
	"errors"

	domain "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/domain/nota_fiscal"
	"github.com/google/uuid"
)

var (
	ErrProdutoNaoEncontrado = errors.New("produto nao encontrado no estoque")
	ErrEstoqueIndisponivel  = errors.New("servico de estoque indisponivel")
)

type CriarNotaFiscalItemInput struct {
	CodigoProduto string
	Quantidade    int
}
type CriarNotaFiscalInput struct {
	NomeCliente     string
	EnderecoCliente string
	Itens           []CriarNotaFiscalItemInput
}
type ProdutoCatalogo struct {
	ID        uuid.UUID
	Codigo    string
	Descricao string
	Ativo     bool
	Valor     float64
}
type produtoCatalogo interface {
	BuscarPorCodigo(context.Context, string) (*ProdutoCatalogo, error)
}
type CriarNotaFiscalUseCase struct {
	repository domain.Repository
	produtos   produtoCatalogo
}

func NewCriarNotaFiscalUseCase(
	repository domain.Repository,
	produtos produtoCatalogo,
) *CriarNotaFiscalUseCase {
	return &CriarNotaFiscalUseCase{repository: repository, produtos: produtos}
}

func (useCase *CriarNotaFiscalUseCase) Execute(
	ctx context.Context,
	input CriarNotaFiscalInput,
) (*domain.NotaFiscal, error) {
	itens, err := montarItens(ctx, useCase.produtos, input.Itens)
	if err != nil {
		return nil, err
	}
	numero, err := useCase.repository.ProximoNumero(ctx)
	if err != nil {
		return nil, err
	}
	nota, err := domain.NewNotaFiscal(numero, input.NomeCliente, input.EnderecoCliente, itens)
	if err != nil {
		return nil, err
	}
	if err := useCase.repository.Criar(ctx, nota); err != nil {
		return nil, err
	}
	return nota, nil
}
