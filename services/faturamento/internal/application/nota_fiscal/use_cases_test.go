package notafiscal

import (
	"context"
	"errors"
	"testing"

	domain "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/domain/nota_fiscal"
	sharedquery "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/shared/query"
	"github.com/google/uuid"
)

type repositoryStub struct {
	numero     int64
	criada     *domain.NotaFiscal
	atualizada *domain.NotaFiscal
	nota       *domain.NotaFiscal
	fechamento *domain.NotaFiscal
}

func TestCriarNotaFiscalRejeitaProdutoInativo(t *testing.T) {
	repository := &repositoryStub{numero: 1}
	catalogo := &catalogoStub{
		produto: &ProdutoCatalogo{
			ID:        uuid.New(),
			Codigo:    "SKU",
			Descricao: "Produto",
			Ativo:     false,
			Valor:     25.00,
		},
	}
	input := CriarNotaFiscalInput{
		NomeCliente:     "Cliente",
		EnderecoCliente: "Rua A",
		Itens: []CriarNotaFiscalItemInput{
			{CodigoProduto: "SKU", Quantidade: 1},
		},
	}
	_, err := NewCriarNotaFiscalUseCase(repository, catalogo).
		Execute(context.Background(), input)
	if !errors.Is(err, domain.ErrProdutoInativo) {
		t.Fatalf("esperava produto inativo, recebeu %v", err)
	}
}

type catalogoStub struct {
	produto *ProdutoCatalogo
}

func (catalogo *catalogoStub) BuscarPorCodigo(context.Context, string) (*ProdutoCatalogo, error) {
	return catalogo.produto, nil
}

func (repository *repositoryStub) ProximoNumero(context.Context) (int64, error) {
	return repository.numero, nil
}
func (repository *repositoryStub) Criar(_ context.Context, nota *domain.NotaFiscal) error {
	repository.criada = nota
	return nil
}
func (repository *repositoryStub) Atualizar(_ context.Context, nota *domain.NotaFiscal) error {
	repository.atualizada = nota
	return nil
}
func (repository *repositoryStub) BuscarPorID(context.Context, uuid.UUID) (*domain.NotaFiscal, error) {
	return repository.nota, nil
}
func (*repositoryStub) Listar(
	context.Context,
	sharedquery.Criteria[domain.ListFilters],
) (sharedquery.Page[domain.NotaFiscal], error) {
	return sharedquery.NewPage([]domain.NotaFiscal{}, 0, sharedquery.NewPagination(1, 20)), nil
}
func (repository *repositoryStub) IniciarFechamento(_ context.Context, nota *domain.NotaFiscal) error {
	repository.fechamento = nota
	return nil
}

func TestCriarNotaFiscalUsaNumeroDoRepository(t *testing.T) {
	repository := &repositoryStub{numero: 42}
	catalogo := &catalogoStub{
		produto: &ProdutoCatalogo{
			ID:        uuid.New(),
			Codigo:    "SKU",
			Descricao: "Produto",
			Ativo:     true,
			Valor:     25.00,
		},
	}
	input := CriarNotaFiscalInput{
		NomeCliente:     "Cliente",
		EnderecoCliente: "Rua A",
		Itens: []CriarNotaFiscalItemInput{
			{CodigoProduto: "SKU", Quantidade: 1},
		},
	}
	nota, err := NewCriarNotaFiscalUseCase(repository, catalogo).
		Execute(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if nota.Numero() != 42 || repository.criada != nota {
		t.Fatal("nota nao foi criada com o numero esperado")
	}
}

func TestAtualizarNotaFiscalRenovaDadosEItens(t *testing.T) {
	itemOriginal, _ := domain.NewItemNotaFiscal(uuid.New(), "ANTIGO", "Produto antigo", 1, 10)
	nota, _ := domain.NewNotaFiscal(
		1,
		"Cliente antigo",
		"Rua antiga",
		[]domain.ItemNotaFiscal{*itemOriginal},
	)
	repository := &repositoryStub{nota: nota}
	catalogo := &catalogoStub{produto: &ProdutoCatalogo{
		ID:        uuid.New(),
		Codigo:    "NOVO",
		Descricao: "Produto novo",
		Ativo:     true,
		Valor:     12.50,
	}}

	result, err := NewAtualizarNotaFiscalUseCase(repository, catalogo).Execute(
		context.Background(),
		AtualizarNotaFiscalInput{
			ID:              nota.ID(),
			NomeCliente:     "Novo cliente",
			EnderecoCliente: "Nova rua",
			Itens: []CriarNotaFiscalItemInput{
				{CodigoProduto: "NOVO", Quantidade: 2},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if repository.atualizada != nota || result.NomeCliente() != "NOVO CLIENTE" {
		t.Fatal("nota atualizada nao foi persistida corretamente")
	}
	if result.QuantidadeTotal() != 2 || result.ValorTotal() != 25 {
		t.Fatalf("totais atualizados incorretos: %d e %.2f", result.QuantidadeTotal(), result.ValorTotal())
	}
}

func TestAtualizarNotaFiscalRejeitaNotaForaDoStatusAberta(t *testing.T) {
	item, _ := domain.NewItemNotaFiscal(uuid.New(), "SKU", "Produto", 1, 10)
	nota, _ := domain.NewNotaFiscal(1, "Cliente", "Rua", []domain.ItemNotaFiscal{*item})
	_ = nota.IniciarFechamento()
	repository := &repositoryStub{nota: nota}

	_, err := NewAtualizarNotaFiscalUseCase(repository, &catalogoStub{}).Execute(
		context.Background(),
		AtualizarNotaFiscalInput{ID: nota.ID()},
	)
	if !errors.Is(err, domain.ErrNotaNaoEstaAberta) {
		t.Fatalf("esperava nota nao aberta, recebeu %v", err)
	}
	if repository.atualizada != nil {
		t.Fatal("nota fora do status ABERTA nao deveria ser persistida")
	}
}

func TestIniciarFechamentoOrquestraDominioEPersistencia(t *testing.T) {
	item, _ := domain.NewItemNotaFiscal(uuid.New(), "SKU", "Produto", 1, 25.00)
	nota, _ := domain.NewNotaFiscal(
		1,
		"Cliente",
		"Rua A",
		[]domain.ItemNotaFiscal{*item},
	)
	repository := &repositoryStub{nota: nota}
	result, err := NewIniciarFechamentoUseCase(repository).
		Execute(context.Background(), nota.ID())
	if err != nil {
		t.Fatal(err)
	}
	if result.Status() != domain.StatusProcessando || repository.fechamento != nota {
		t.Fatal("fechamento nao foi orquestrado")
	}
}
