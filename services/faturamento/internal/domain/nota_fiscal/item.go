package notafiscal

import (
	"errors"

	sharedtext "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/shared/text"
	"github.com/google/uuid"
)

var (
	ErrProdutoInvalido      = errors.New("produto do item e obrigatorio")
	ErrCodigoObrigatorio    = errors.New("codigo do produto e obrigatorio")
	ErrDescricaoObrigatoria = errors.New("descricao do produto e obrigatoria")
	ErrQuantidadeInvalida   = errors.New("quantidade deve ser positiva")
	ErrValorInvalido        = errors.New("valor deve ser maior que zero")
	ErrProdutoInativo       = errors.New("produto esta inativo")
)

type ItemNotaFiscal struct {
	id               uuid.UUID
	produtoID        uuid.UUID
	codigoProduto    string
	descricaoProduto string
	quantidade       int
	valor            float64
}

func NewItemNotaFiscal(
	produtoID uuid.UUID,
	codigo, descricao string,
	quantidade int,
	valor float64,
) (*ItemNotaFiscal, error) {
	return NewItemNotaFiscalWithState(uuid.New(), produtoID, codigo, descricao, quantidade, valor)
}

func NewItemNotaFiscalWithState(
	id, produtoID uuid.UUID,
	codigo, descricao string,
	quantidade int,
	valor float64,
) (*ItemNotaFiscal, error) {
	codigo = sharedtext.NormalizeUpper(codigo)
	descricao = sharedtext.NormalizeUpper(descricao)
	if id == uuid.Nil || produtoID == uuid.Nil {
		return nil, ErrProdutoInvalido
	}
	if codigo == "" {
		return nil, ErrCodigoObrigatorio
	}
	if descricao == "" {
		return nil, ErrDescricaoObrigatoria
	}
	if quantidade <= 0 {
		return nil, ErrQuantidadeInvalida
	}
	if valor <= 0 {
		return nil, ErrValorInvalido
	}
	return &ItemNotaFiscal{
		id:               id,
		produtoID:        produtoID,
		codigoProduto:    codigo,
		descricaoProduto: descricao,
		quantidade:       quantidade,
		valor:            valor,
	}, nil
}

func (item *ItemNotaFiscal) ID() uuid.UUID {
	return item.id
}

func (item *ItemNotaFiscal) ProdutoID() uuid.UUID {
	return item.produtoID
}

func (item *ItemNotaFiscal) CodigoProduto() string {
	return item.codigoProduto
}

func (item *ItemNotaFiscal) DescricaoProduto() string {
	return item.descricaoProduto
}

func (item *ItemNotaFiscal) Quantidade() int {
	return item.quantidade
}

func (item *ItemNotaFiscal) Valor() float64 {
	return item.valor
}

func (item *ItemNotaFiscal) ValorTotal() float64 {
	return float64(item.quantidade) * item.valor
}
