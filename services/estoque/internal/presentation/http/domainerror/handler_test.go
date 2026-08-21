package domainerror

import (
	"errors"
	"net/http"
	"testing"

	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
)

func TestMapHandlesWrappedDomainError(t *testing.T) {
	mapping := Map(errors.Join(errors.New("repository context"), domain.ErrProdutoNaoEncontrado))

	if mapping.Status != http.StatusNotFound || mapping.Code != "PRODUTO_NAO_ENCONTRADO" {
		t.Fatalf("mapeamento inesperado: %+v", mapping)
	}
}

func TestMapHidesUnexpectedError(t *testing.T) {
	mapping := Map(errors.New("database password leaked"))

	if mapping.Status != http.StatusInternalServerError || mapping.Message != "erro interno do servidor" {
		t.Fatalf("erro interno nao foi protegido: %+v", mapping)
	}
}
