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
}

func (repository *repositoryStub) BaixarEstoque(_ context.Context, baixa domain.BaixaEstoque) (bool, error) {
	repository.baixa = baixa
	return repository.processada, nil
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
