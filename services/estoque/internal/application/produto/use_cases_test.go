package produto

import (
	"context"
	"testing"

	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/movimentacao"
	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
	"github.com/google/uuid"
)

type repositoryStub struct {
	criado     *domain.Produto
	produto    *domain.Produto
	produtos   []domain.Produto
	id         uuid.UUID
	codigo     string
	atualizado *domain.Produto
}

func (repository *repositoryStub) Criar(_ context.Context, produto *domain.Produto) error {
	repository.criado = produto
	return nil
}
func (repository *repositoryStub) Atualizar(_ context.Context, produto *domain.Produto) error {
	repository.atualizado = produto
	return nil
}
func (*repositoryStub) BaixarEstoque(context.Context, domain.BaixaEstoque) (bool, error) {
	return true, nil
}
func (*repositoryStub) ListarMovimentacoes(context.Context, uuid.UUID) ([]movimentacao.Movimentacao, error) {
	return nil, nil
}
func (repository *repositoryStub) BuscarPorID(_ context.Context, id uuid.UUID) (*domain.Produto, error) {
	repository.id = id
	return repository.produto, nil
}
func (repository *repositoryStub) BuscarPorCodigo(_ context.Context, codigo string) (*domain.Produto, error) {
	repository.codigo = codigo
	return repository.produto, nil
}
func (repository *repositoryStub) Listar(context.Context) ([]domain.Produto, error) {
	return repository.produtos, nil
}

func TestCriarProdutoUseCaseValidaEPersiste(t *testing.T) {
	repository := &repositoryStub{}
	useCase := NewCriarProdutoUseCase(repository)

	produto, err := useCase.Execute(context.Background(), CriarProdutoInput{
		Codigo:    "SKU-001",
		Descricao: "Teclado",
		Saldo:     5,
	})
	if err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}
	if repository.criado != produto {
		t.Fatal("esperava que o produto criado fosse persistido")
	}
}

func TestBuscarProdutoPorIDUseCaseDelegaAoRepository(t *testing.T) {
	id := uuid.New()
	repository := &repositoryStub{}
	_, _ = NewBuscarProdutoPorIDUseCase(repository).Execute(context.Background(), id)
	if repository.id != id {
		t.Fatalf("esperava id %s, recebeu %s", id, repository.id)
	}
}

func TestBuscarProdutoPorCodigoUseCaseDelegaAoRepository(t *testing.T) {
	repository := &repositoryStub{}
	_, _ = NewBuscarProdutoPorCodigoUseCase(repository).Execute(context.Background(), "SKU-001")
	if repository.codigo != "SKU-001" {
		t.Fatalf("esperava codigo SKU-001, recebeu %s", repository.codigo)
	}
}

func TestListarProdutosUseCaseRetornaProdutos(t *testing.T) {
	produto, err := domain.NewProduto("SKU-001", "Teclado", 5)
	if err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}
	repository := &repositoryStub{produtos: []domain.Produto{*produto}}

	produtos, err := NewListarProdutosUseCase(repository).Execute(context.Background())
	if err != nil || len(produtos) != 1 {
		t.Fatalf("resultado inesperado: produtos=%d erro=%v", len(produtos), err)
	}
}

func TestInativarEAtivarProdutoUseCases(t *testing.T) {
	produto, err := domain.NewProduto("SKU-001", "Teclado", 5)
	if err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}
	repository := &repositoryStub{produto: produto}

	if _, err := NewInativarProdutoUseCase(repository).Execute(context.Background(), produto.ID()); err != nil {
		t.Fatalf("nao esperava erro ao inativar: %v", err)
	}
	if produto.Ativo() || repository.atualizado != produto {
		t.Fatal("produto inativo deveria ser persistido")
	}

	if _, err := NewAtivarProdutoUseCase(repository).Execute(context.Background(), produto.ID()); err != nil {
		t.Fatalf("nao esperava erro ao ativar: %v", err)
	}
	if !produto.Ativo() {
		t.Fatal("produto deveria voltar a ficar ativo")
	}
}

func TestAtualizarProdutoUseCaseAtualizaCamposPermitidos(t *testing.T) {
	produto, err := domain.NewProduto("SKU-001", "Teclado", 5)
	if err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}
	repository := &repositoryStub{produto: produto}

	atualizado, err := NewAtualizarProdutoUseCase(repository).Execute(
		context.Background(),
		AtualizarProdutoInput{
			ID:        produto.ID(),
			Descricao: "Mouse",
			Saldo:     12,
		},
	)
	if err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}
	if atualizado.Descricao() != "MOUSE" || atualizado.Saldo() != 12 {
		t.Fatalf("produto nao foi atualizado: descricao=%s saldo=%d", atualizado.Descricao(), atualizado.Saldo())
	}
	if repository.atualizado != produto {
		t.Fatal("produto atualizado deveria ser persistido")
	}
}
