package movimentacao

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewAjusteSaldoEntrada(t *testing.T) {
	movimentacao, err := NewAjusteSaldo(uuid.New(), 5, 9)
	if err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}
	if movimentacao.Tipo() != TipoEntrada || movimentacao.Quantidade() != 4 {
		t.Fatalf(
			"esperava entrada de 4, recebeu tipo=%s quantidade=%d",
			movimentacao.Tipo(),
			movimentacao.Quantidade(),
		)
	}
}

func TestNewAjusteSaldoSaida(t *testing.T) {
	movimentacao, err := NewAjusteSaldo(uuid.New(), 9, 5)
	if err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}
	if movimentacao.Tipo() != TipoSaida || movimentacao.Quantidade() != 4 {
		t.Fatalf(
			"esperava saida de 4, recebeu tipo=%s quantidade=%d",
			movimentacao.Tipo(),
			movimentacao.Quantidade(),
		)
	}
}

func TestNewAjusteSaldoSemAlteracao(t *testing.T) {
	movimentacao, err := NewAjusteSaldo(uuid.New(), 5, 5)
	if err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}
	if movimentacao != nil {
		t.Fatal("nao deveria criar movimentacao sem alteracao de saldo")
	}
}
