package notafiscal

import (
	"context"
	"errors"
	domain "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/domain/nota_fiscal"
	"github.com/google/uuid"
	"testing"
)

type repositoryStub struct {
	numero     int64
	criada     *domain.NotaFiscal
	nota       *domain.NotaFiscal
	fechamento *domain.NotaFiscal
}

func TestCriarNotaFiscalRejeitaProdutoInativo(t *testing.T) {
	repository := &repositoryStub{numero: 1}
	catalogo := &catalogoStub{produto: &ProdutoCatalogo{ID: uuid.New(), Codigo: "SKU", Descricao: "Produto", Ativo: false, Valor: 2500}}
	_, err := NewCriarNotaFiscalUseCase(repository, catalogo).Execute(context.Background(), CriarNotaFiscalInput{NomeCliente: "Cliente", EnderecoCliente: "Rua A", Itens: []CriarNotaFiscalItemInput{{CodigoProduto: "SKU", Quantidade: 1}}})
	if !errors.Is(err, domain.ErrProdutoInativo) {
		t.Fatalf("esperava produto inativo, recebeu %v", err)
	}
}

type catalogoStub struct{ produto *ProdutoCatalogo }

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
func (repository *repositoryStub) BuscarPorID(context.Context, uuid.UUID) (*domain.NotaFiscal, error) {
	return repository.nota, nil
}
func (*repositoryStub) Listar(context.Context) ([]domain.NotaFiscal, error) { return nil, nil }
func (repository *repositoryStub) IniciarFechamento(_ context.Context, nota *domain.NotaFiscal) error {
	repository.fechamento = nota
	return nil
}

func TestCriarNotaFiscalUsaNumeroDoRepository(t *testing.T) {
	repository := &repositoryStub{numero: 42}
	catalogo := &catalogoStub{produto: &ProdutoCatalogo{ID: uuid.New(), Codigo: "SKU", Descricao: "Produto", Ativo: true, Valor: 2500}}
	nota, err := NewCriarNotaFiscalUseCase(repository, catalogo).Execute(context.Background(), CriarNotaFiscalInput{NomeCliente: "Cliente", EnderecoCliente: "Rua A", Itens: []CriarNotaFiscalItemInput{{CodigoProduto: "SKU", Quantidade: 1}}})
	if err != nil {
		t.Fatal(err)
	}
	if nota.Numero() != 42 || repository.criada != nota {
		t.Fatal("nota nao foi criada com o numero esperado")
	}
}

func TestIniciarFechamentoOrquestraDominioEPersistencia(t *testing.T) {
	item, _ := domain.NewItemNotaFiscal(uuid.New(), "SKU", "Produto", 1, 2500)
	nota, _ := domain.NewNotaFiscal(1, "Cliente", "Rua A", []domain.ItemNotaFiscal{*item})
	repository := &repositoryStub{nota: nota}
	result, err := NewIniciarFechamentoUseCase(repository).Execute(context.Background(), nota.ID())
	if err != nil {
		t.Fatal(err)
	}
	if result.Status() != domain.StatusProcessando || repository.fechamento != nota {
		t.Fatal("fechamento nao foi orquestrado")
	}
}
