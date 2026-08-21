package produto

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrEventoInvalido = errors.New("eventId e obrigatorio")
	ErrNotaInvalida   = errors.New("notaFiscalId e obrigatorio")
	ErrItensVazios    = errors.New("a baixa deve possuir ao menos um item")
)

type BaixaItem struct {
	produtoID  uuid.UUID
	quantidade int
}

type BaixaEstoque struct {
	eventID      uuid.UUID
	notaFiscalID uuid.UUID
	itens        []BaixaItem
}

func NewBaixaItem(produtoID uuid.UUID, quantidade int) (*BaixaItem, error) {
	if produtoID == uuid.Nil || quantidade <= 0 {
		return nil, ErrQuantidadeInvalida
	}
	return &BaixaItem{produtoID: produtoID, quantidade: quantidade}, nil
}

func NewBaixaEstoque(eventID, notaFiscalID uuid.UUID, itens []BaixaItem) (*BaixaEstoque, error) {
	if eventID == uuid.Nil {
		return nil, ErrEventoInvalido
	}
	if notaFiscalID == uuid.Nil {
		return nil, ErrNotaInvalida
	}
	if len(itens) == 0 {
		return nil, ErrItensVazios
	}

	itensAgrupados := make(map[uuid.UUID]int, len(itens))
	for _, item := range itens {
		itensAgrupados[item.produtoID] += item.quantidade
	}

	baixa := &BaixaEstoque{
		eventID:      eventID,
		notaFiscalID: notaFiscalID,
		itens:        make([]BaixaItem, 0, len(itensAgrupados)),
	}
	for produtoID, quantidade := range itensAgrupados {
		baixa.itens = append(baixa.itens, BaixaItem{
			produtoID:  produtoID,
			quantidade: quantidade,
		})
	}
	return baixa, nil
}

func (item BaixaItem) ProdutoID() uuid.UUID {
	return item.produtoID
}

func (item BaixaItem) Quantidade() int {
	return item.quantidade
}

func (baixa *BaixaEstoque) EventID() uuid.UUID {
	return baixa.eventID
}

func (baixa *BaixaEstoque) NotaFiscalID() uuid.UUID {
	return baixa.notaFiscalID
}

func (baixa *BaixaEstoque) Itens() []BaixaItem {
	return append([]BaixaItem(nil), baixa.itens...)
}
