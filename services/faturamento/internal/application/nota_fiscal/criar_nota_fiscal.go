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
	if len(input.Itens) == 0 {
		return nil, domain.ErrNotaSemItens
	}
	itens := make([]domain.ItemNotaFiscal, 0, len(input.Itens))
	for _, inputItem := range input.Itens {
		produto, err := useCase.produtos.BuscarPorCodigo(ctx, inputItem.CodigoProduto)
		if err != nil {
			return nil, err
		}
		if !produto.Ativo {
			return nil, domain.ErrProdutoInativo
		}
		item, err := domain.NewItemNotaFiscal(
			produto.ID,
			produto.Codigo,
			produto.Descricao,
			inputItem.Quantidade,
			produto.Valor,
		)
		if err != nil {
			return nil, err
		}
		itens = append(itens, *item)
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
