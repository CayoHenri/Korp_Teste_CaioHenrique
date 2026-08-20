package produto

import (
	"context"
	"testing"

	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
	"github.com/google/uuid"
)

type repositoryStub struct {
	criado *domain.Produto
}

func (repository *repositoryStub) Criar(_ context.Context, produto *domain.Produto) error {
	repository.criado = produto
	return nil
}
func (*repositoryStub) BuscarPorID(context.Context, uuid.UUID) (*domain.Produto, error) {
	return nil, domain.ErrProdutoNaoEncontrado
}
func (*repositoryStub) BuscarPorCodigo(context.Context, string) (*domain.Produto, error) {
	return nil, domain.ErrProdutoNaoEncontrado
}
func (*repositoryStub) Listar(context.Context) ([]domain.Produto, error) { return nil, nil }

func TestCriarValidaEPersisteProduto(t *testing.T) {
	repository := &repositoryStub{}
	service := NewService(repository)

	produto, err := service.Criar(context.Background(), CriarInput{
		Codigo: "SKU-001", Descricao: "Teclado", Saldo: 5,
	})
	if err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}
	if repository.criado != produto {
		t.Fatal("esperava que o produto criado fosse persistido")
	}
}
