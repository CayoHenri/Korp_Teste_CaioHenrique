package models

import (
	"testing"

	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
)

func TestProdutoConversionRoundTrip(t *testing.T) {
	produto, err := domain.NewProduto("sku-001", "teclado", 5)
	if err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}

	restored, err := NewProdutoModel(produto).ToDomain()
	if err != nil {
		t.Fatalf("nao esperava erro ao reconstituir: %v", err)
	}
	if restored.ID() != produto.ID() || restored.Codigo() != produto.Codigo() ||
		restored.Descricao() != produto.Descricao() || restored.Saldo() != produto.Saldo() {
		t.Fatalf("conversao alterou o produto: esperado %+v, recebido %+v", produto, restored)
	}
}
