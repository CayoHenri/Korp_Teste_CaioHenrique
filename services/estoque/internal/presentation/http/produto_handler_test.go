package http

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	application "github.com/caiog/korp-notas-fiscais/services/estoque/internal/application/produto"
	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
	"github.com/google/uuid"
)

type produtoRepositoryStub struct {
	produtos []domain.Produto
}

func (repository *produtoRepositoryStub) Criar(_ context.Context, produto *domain.Produto) error {
	repository.produtos = append(repository.produtos, *produto)
	return nil
}
func (*produtoRepositoryStub) BuscarPorID(context.Context, uuid.UUID) (*domain.Produto, error) {
	return nil, domain.ErrProdutoNaoEncontrado
}
func (*produtoRepositoryStub) BuscarPorCodigo(context.Context, string) (*domain.Produto, error) {
	return nil, domain.ErrProdutoNaoEncontrado
}
func (repository *produtoRepositoryStub) Listar(context.Context) ([]domain.Produto, error) {
	return repository.produtos, nil
}

func TestCriarProdutoReturnsCreated(t *testing.T) {
	repository := &produtoRepositoryStub{}
	handler := NewProdutoHandler(application.NewService(repository))
	router := NewRouter(databaseStub{}, handler)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/produtos",
		bytes.NewBufferString(`{"codigo":"SKU-001","descricao":"Teclado","saldo":5}`),
	)
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("esperava status %d, recebeu %d: %s", http.StatusCreated, recorder.Code, recorder.Body.String())
	}
	if len(repository.produtos) != 1 {
		t.Fatal("esperava um produto persistido")
	}
}

func TestBuscarProdutoRejectsInvalidID(t *testing.T) {
	handler := NewProdutoHandler(application.NewService(&produtoRepositoryStub{}))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/produtos/id-invalido", nil)

	NewRouter(databaseStub{}, handler).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("esperava status %d, recebeu %d", http.StatusBadRequest, recorder.Code)
	}
}
