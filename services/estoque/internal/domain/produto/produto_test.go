package produto

import (
	"errors"
	"testing"
)

func TestNovoProdutoNormalizaDados(t *testing.T) {
	produto, err := Novo("  SKU-001  ", "  Teclado  ", 10)
	if err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}

	if produto.Codigo != "SKU-001" || produto.Descricao != "Teclado" {
		t.Fatalf("dados nao foram normalizados: %+v", produto)
	}
	if produto.ID.String() == "00000000-0000-0000-0000-000000000000" {
		t.Fatal("esperava um UUID gerado")
	}
}

func TestNovoProdutoValidaInvariantes(t *testing.T) {
	tests := []struct {
		name      string
		codigo    string
		descricao string
		saldo     int
		expected  error
	}{
		{name: "codigo vazio", descricao: "Produto", saldo: 0, expected: ErrCodigoObrigatorio},
		{name: "descricao vazia", codigo: "SKU", saldo: 0, expected: ErrDescricaoObrigatoria},
		{name: "saldo negativo", codigo: "SKU", descricao: "Produto", saldo: -1, expected: ErrSaldoInvalido},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Novo(test.codigo, test.descricao, test.saldo)
			if !errors.Is(err, test.expected) {
				t.Fatalf("esperava %v, recebeu %v", test.expected, err)
			}
		})
	}
}
