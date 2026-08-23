package produto

import (
	"errors"
	"testing"
)

func TestNovoProdutoNormalizaDados(t *testing.T) {
	produto, err := NewProduto("  sku-001  ", "  Teclado Mecânico  ", 10, 159.90)
	if err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}

	if produto.Codigo() != "SKU-001" || produto.Descricao() != "TECLADO MECÂNICO" {
		t.Fatalf("dados nao foram normalizados: %+v", produto)
	}
	if produto.ID().String() == "00000000-0000-0000-0000-000000000000" {
		t.Fatal("esperava um UUID gerado")
	}
	if !produto.Ativo() {
		t.Fatal("produto novo deve iniciar ativo")
	}
	if produto.Valor() != 159.90 {
		t.Fatalf("esperava valor 159.90, recebeu %.2f", produto.Valor())
	}
}

func TestAtivarEInativarProduto(t *testing.T) {
	produto, err := NewProduto("SKU", "Teclado", 1, 159.90)
	if err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}

	produto.Inativar()
	if produto.Ativo() {
		t.Fatal("produto deveria estar inativo")
	}

	produto.Ativar()
	if !produto.Ativo() {
		t.Fatal("produto deveria estar ativo")
	}
}

func TestProdutoValidarBaixa(t *testing.T) {
	produto, err := NewProduto("SKU", "Teclado", 2, 159.90)
	if err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}
	if err := produto.ValidarBaixa(2); err != nil {
		t.Fatalf("baixa valida retornou erro: %v", err)
	}
	if err := produto.ValidarBaixa(3); !errors.Is(err, ErrEstoqueInsuficiente) {
		t.Fatalf("esperava estoque insuficiente, recebeu %v", err)
	}
	produto.Inativar()
	if err := produto.ValidarBaixa(1); !errors.Is(err, ErrProdutoInativo) {
		t.Fatalf("esperava produto inativo, recebeu %v", err)
	}
}

func TestAtualizarProdutoPreservaInvariantes(t *testing.T) {
	produto, err := NewProduto("SKU", "Teclado", 1, 159.90)
	if err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}

	if err := produto.AtualizarDescricao("  mouse gamer  "); err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}
	if err := produto.AtualizarSaldo(8); err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}
	if err := produto.AtualizarValor(249.90); err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}
	if produto.Descricao() != "MOUSE GAMER" {
		t.Fatalf("descricao nao foi normalizada: %q", produto.Descricao())
	}
	if produto.Saldo() != 8 {
		t.Fatalf("esperava saldo 8, recebeu %d", produto.Saldo())
	}
	if produto.Valor() != 249.90 {
		t.Fatalf("esperava valor 249.90, recebeu %.2f", produto.Valor())
	}
	if err := produto.AtualizarDescricao("  "); !errors.Is(err, ErrDescricaoObrigatoria) {
		t.Fatalf("esperava erro de descricao, recebeu %v", err)
	}
	if produto.Descricao() != "MOUSE GAMER" || produto.Saldo() != 8 {
		t.Fatal("atualizacao invalida nao deve alterar parcialmente o produto")
	}
	if err := produto.AtualizarSaldo(-1); !errors.Is(err, ErrSaldoInvalido) {
		t.Fatalf("esperava erro de saldo, recebeu %v", err)
	}
	if err := produto.AtualizarValor(0); !errors.Is(err, ErrValorInvalido) {
		t.Fatalf("esperava erro de valor, recebeu %v", err)
	}
}

func TestNovoProdutoValidaInvariantes(t *testing.T) {
	tests := []struct {
		name      string
		codigo    string
		descricao string
		saldo     int
		valor     float64
		expected  error
	}{
		{
			name:      "codigo vazio",
			descricao: "Produto",
			saldo:     0,
			expected:  ErrCodigoObrigatorio,
			valor:     10,
		},
		{
			name:     "descricao vazia",
			codigo:   "SKU",
			saldo:    0,
			expected: ErrDescricaoObrigatoria,
			valor:    10,
		},
		{
			name:      "saldo negativo",
			codigo:    "SKU",
			descricao: "Produto",
			saldo:     -1,
			expected:  ErrSaldoInvalido,
			valor:     10,
		},
		{
			name:      "valor zerado",
			codigo:    "SKU",
			descricao: "Produto",
			saldo:     1,
			valor:     0,
			expected:  ErrValorInvalido,
		},
		{
			name:      "valor negativo",
			codigo:    "SKU",
			descricao: "Produto",
			saldo:     1,
			valor:     -0.01,
			expected:  ErrValorInvalido,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewProduto(test.codigo, test.descricao, test.saldo, test.valor)
			if !errors.Is(err, test.expected) {
				t.Fatalf("esperava %v, recebeu %v", test.expected, err)
			}
		})
	}
}
