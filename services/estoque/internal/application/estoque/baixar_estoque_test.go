package estoque

import (
	"context"
	"errors"
	"testing"

	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
	"github.com/google/uuid"
)

type repositoryStub struct {
	baixa      domain.BaixaEstoque
	processada bool
	err        error
}

func (repository *repositoryStub) BaixarEstoque(_ context.Context, baixa domain.BaixaEstoque) (bool, error) {
	repository.baixa = baixa
	return repository.processada, repository.err
}

type resultadoPublisherStub struct {
	resultado ResultadoBaixa
	err       error
}

func (publisher *resultadoPublisherStub) PublicarResultado(
	_ context.Context,
	resultado ResultadoBaixa,
) error {
	publisher.resultado = resultado
	return publisher.err
}

func TestProcessarBaixaSolicitadaPublicaSucesso(t *testing.T) {
	repository := &repositoryStub{processada: true}
	publisher := &resultadoPublisherStub{}
	input := BaixarEstoqueInput{
		EventID:      uuid.New(),
		NotaFiscalID: uuid.New(),
		Itens: []BaixarEstoqueItemInput{
			{ProdutoID: uuid.New(), Quantidade: 1},
		},
	}

	err := NewProcessarBaixaSolicitadaUseCase(
		NewBaixarEstoqueUseCase(repository),
		publisher,
	).Execute(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if publisher.resultado.Type != EventTypeBaixaRealizada ||
		publisher.resultado.CorrelationID != input.EventID ||
		publisher.resultado.NotaFiscalID != input.NotaFiscalID {
		t.Fatalf("resultado inesperado: %+v", publisher.resultado)
	}
}

func TestProcessarBaixaSolicitadaPublicaRejeicaoDeNegocio(t *testing.T) {
	repository := &repositoryStub{err: domain.ErrEstoqueInsuficiente}
	publisher := &resultadoPublisherStub{}
	input := BaixarEstoqueInput{
		EventID:      uuid.New(),
		NotaFiscalID: uuid.New(),
		Itens: []BaixarEstoqueItemInput{
			{ProdutoID: uuid.New(), Quantidade: 1},
		},
	}

	err := NewProcessarBaixaSolicitadaUseCase(
		NewBaixarEstoqueUseCase(repository),
		publisher,
	).Execute(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if publisher.resultado.Type != EventTypeBaixaRejeitada ||
		publisher.resultado.Motivo != "ESTOQUE_INSUFICIENTE" {
		t.Fatalf("rejeicao inesperada: %+v", publisher.resultado)
	}
}

func TestProcessarBaixaSolicitadaRetornaErroDeInfraestrutura(t *testing.T) {
	infrastructureErr := errors.New("banco indisponivel")
	repository := &repositoryStub{err: infrastructureErr}
	publisher := &resultadoPublisherStub{}
	input := BaixarEstoqueInput{
		EventID:      uuid.New(),
		NotaFiscalID: uuid.New(),
		Itens: []BaixarEstoqueItemInput{
			{ProdutoID: uuid.New(), Quantidade: 1},
		},
	}

	err := NewProcessarBaixaSolicitadaUseCase(
		NewBaixarEstoqueUseCase(repository),
		publisher,
	).Execute(context.Background(), input)
	if !errors.Is(err, infrastructureErr) {
		t.Fatalf("esperava erro de infraestrutura, recebeu %v", err)
	}
	if publisher.resultado.EventID != uuid.Nil {
		t.Fatal("erro de infraestrutura nao deveria publicar resultado")
	}
}

func TestBaixarEstoqueAgrupaItensDoMesmoProduto(t *testing.T) {
	produtoID := uuid.New()
	repository := &repositoryStub{processada: true}
	output, err := NewBaixarEstoqueUseCase(repository).Execute(context.Background(), BaixarEstoqueInput{
		EventID:      uuid.New(),
		NotaFiscalID: uuid.New(),
		Itens: []BaixarEstoqueItemInput{
			{
				ProdutoID:  produtoID,
				Quantidade: 2,
			},
			{
				ProdutoID:  produtoID,
				Quantidade: 3,
			},
		},
	})
	if err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}
	if output.Duplicada || len(repository.baixa.Itens()) != 1 || repository.baixa.Itens()[0].Quantidade() != 5 {
		t.Fatalf("agrupamento inesperado: %+v", repository.baixa.Itens())
	}
}

func TestBaixarEstoqueValidaQuantidade(t *testing.T) {
	_, err := NewBaixarEstoqueUseCase(&repositoryStub{}).Execute(context.Background(), BaixarEstoqueInput{
		EventID:      uuid.New(),
		NotaFiscalID: uuid.New(),
		Itens: []BaixarEstoqueItemInput{
			{
				ProdutoID:  uuid.New(),
				Quantidade: 0,
			},
		},
	})
	if !errors.Is(err, domain.ErrQuantidadeInvalida) {
		t.Fatalf("esperava quantidade invalida, recebeu %v", err)
	}
}
