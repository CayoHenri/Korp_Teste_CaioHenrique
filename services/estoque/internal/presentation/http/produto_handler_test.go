package http

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	estoqueApplication "github.com/caiog/korp-notas-fiscais/services/estoque/internal/application/estoque"
	application "github.com/caiog/korp-notas-fiscais/services/estoque/internal/application/produto"
	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/movimentacao"
	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
	"github.com/google/uuid"
)

type produtoRepositoryStub struct {
	produtos []domain.Produto
}

func newProdutoHandlerForTest(repository *produtoRepositoryStub) *ProdutoHandler {
	return NewProdutoHandler(
		application.NewCriarProdutoUseCase(repository),
		application.NewListarProdutosUseCase(repository),
		application.NewBuscarProdutoPorIDUseCase(repository),
		application.NewBuscarProdutoPorCodigoUseCase(repository),
		application.NewAtivarProdutoUseCase(repository),
		application.NewInativarProdutoUseCase(repository),
		application.NewAtualizarProdutoUseCase(repository),
		estoqueApplication.NewListarMovimentacoesUseCase(repository),
	)
}

func (repository *produtoRepositoryStub) Criar(_ context.Context, produto *domain.Produto) error {
	repository.produtos = append(repository.produtos, *produto)
	return nil
}
func (*produtoRepositoryStub) Atualizar(context.Context, *domain.Produto) error {
	return nil
}
func (*produtoRepositoryStub) BaixarEstoque(context.Context, domain.BaixaEstoque) (bool, error) {
	return true, nil
}
func (*produtoRepositoryStub) ListarMovimentacoes(context.Context, uuid.UUID) ([]movimentacao.Movimentacao, error) {
	return nil, nil
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
	handler := newProdutoHandlerForTest(repository)
	router := NewRouter(databaseStub{}, handler)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/produtos",
		bytes.NewBufferString(`{"codigo":"SKU-001","descricao":"Teclado","saldo":5,"valor":159.90}`),
	)
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("esperava status %d, recebeu %d: %s", http.StatusCreated, recorder.Code, recorder.Body.String())
	}
	if len(repository.produtos) != 1 {
		t.Fatal("esperava um produto persistido")
	}
	if repository.produtos[0].Descricao() != "TECLADO" {
		t.Fatalf("esperava descricao em uppercase, recebeu %q", repository.produtos[0].Descricao())
	}
	if repository.produtos[0].Valor() != 159.90 {
		t.Fatalf("esperava valor 159.90, recebeu %.2f", repository.produtos[0].Valor())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"valor":159.9`)) {
		t.Fatalf("resposta nao contem o valor do produto: %s", recorder.Body.String())
	}
}

func TestBuscarProdutoRejectsInvalidID(t *testing.T) {
	handler := newProdutoHandlerForTest(&produtoRepositoryStub{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/produtos/id-invalido", nil)

	NewRouter(databaseStub{}, handler).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("esperava status %d, recebeu %d", http.StatusBadRequest, recorder.Code)
	}
}
