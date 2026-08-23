package notafiscal

import (
	"context"

	domain "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/domain/nota_fiscal"
)

func montarItens(
	ctx context.Context,
	produtos produtoCatalogo,
	inputs []CriarNotaFiscalItemInput,
) ([]domain.ItemNotaFiscal, error) {
	if len(inputs) == 0 {
		return nil, domain.ErrNotaSemItens
	}
	itens := make([]domain.ItemNotaFiscal, 0, len(inputs))
	for _, input := range inputs {
		produto, err := produtos.BuscarPorCodigo(ctx, input.CodigoProduto)
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
			input.Quantidade,
			produto.Valor,
		)
		if err != nil {
			return nil, err
		}
		itens = append(itens, *item)
	}
	return itens, nil
}
