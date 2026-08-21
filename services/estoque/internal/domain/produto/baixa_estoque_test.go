package produto

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestNewBaixaEstoqueAgrupaItensDoMesmoProduto(t *testing.T) {
	produtoID := uuid.New()
	primeiro, _ := NewBaixaItem(produtoID, 2)
	segundo, _ := NewBaixaItem(produtoID, 3)
	baixa, err := NewBaixaEstoque(uuid.New(), uuid.New(), []BaixaItem{*primeiro, *segundo})
	if err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}
	if len(baixa.Itens()) != 1 || baixa.Itens()[0].Quantidade() != 5 {
		t.Fatalf("esperava um item com quantidade 5, recebeu %+v", baixa.Itens())
	}
}

func TestNewBaixaEstoqueValidaDadosObrigatorios(t *testing.T) {
	item, _ := NewBaixaItem(uuid.New(), 1)
	itemValido := []BaixaItem{*item}
	tests := []struct {
		name         string
		eventID      uuid.UUID
		notaFiscalID uuid.UUID
		itens        []BaixaItem
		expected     error
	}{
		{name: "evento invalido", notaFiscalID: uuid.New(), itens: itemValido, expected: ErrEventoInvalido},
		{name: "nota invalida", eventID: uuid.New(), itens: itemValido, expected: ErrNotaInvalida},
		{name: "itens vazios", eventID: uuid.New(), notaFiscalID: uuid.New(), expected: ErrItensVazios},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewBaixaEstoque(test.eventID, test.notaFiscalID, test.itens)
			if !errors.Is(err, test.expected) {
				t.Fatalf("esperava %v, recebeu %v", test.expected, err)
			}
		})
	}
}

func TestNewBaixaItemValidaQuantidade(t *testing.T) {
	_, err := NewBaixaItem(uuid.New(), 0)
	if !errors.Is(err, ErrQuantidadeInvalida) {
		t.Fatalf("esperava quantidade invalida, recebeu %v", err)
	}
}
