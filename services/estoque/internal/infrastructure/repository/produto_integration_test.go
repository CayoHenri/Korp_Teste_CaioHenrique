//go:build integration

package repository

import (
	"context"
	"errors"
	"sync"
	"testing"

	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/infrastructure/config"
	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/infrastructure/database"
	"github.com/google/uuid"
)

func integrationRepository(t *testing.T) (*ProdutoRepository, func([]uuid.UUID, []uuid.UUID)) {
	t.Helper()
	databaseURL, err := config.DatabaseURL()
	if err != nil {
		t.Fatal(err)
	}
	connection, err := database.Open(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })

	cleanup := func(productIDs, eventIDs []uuid.UUID) {
		if len(productIDs) > 0 {
			connection.Gorm.Exec("DELETE FROM estoque.movimentacoes_estoque WHERE produto_id IN ?", productIDs)
			connection.Gorm.Exec("DELETE FROM estoque.produtos WHERE id IN ?", productIDs)
		}
		if len(eventIDs) > 0 {
			connection.Gorm.Exec("DELETE FROM estoque.mensagens_processadas WHERE event_id IN ?", eventIDs)
		}
	}
	return NewProdutoRepository(connection.Gorm), cleanup
}

func createProduct(t *testing.T, repository *ProdutoRepository, code string, balance int) *domain.Produto {
	t.Helper()
	produto, err := domain.NewProduto(code, "Produto de integracao", balance, 159.90)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Criar(context.Background(), produto); err != nil {
		t.Fatal(err)
	}
	return produto
}

func createBaixa(
	t *testing.T,
	eventID uuid.UUID,
	notaFiscalID uuid.UUID,
	quantidades map[uuid.UUID]int,
) domain.BaixaEstoque {
	t.Helper()
	itens := make([]domain.BaixaItem, 0, len(quantidades))
	for produtoID, quantidade := range quantidades {
		item, err := domain.NewBaixaItem(produtoID, quantidade)
		if err != nil {
			t.Fatal(err)
		}
		itens = append(itens, *item)
	}
	baixa, err := domain.NewBaixaEstoque(eventID, notaFiscalID, itens)
	if err != nil {
		t.Fatal(err)
	}
	return *baixa
}

func TestIntegrationBaixaIsIdempotentAndUpdatesMultipleProducts(t *testing.T) {
	repository, cleanup := integrationRepository(t)
	first := createProduct(t, repository, "INT-"+uuid.NewString(), 10)
	second := createProduct(t, repository, "INT-"+uuid.NewString(), 8)
	eventID := uuid.New()
	t.Cleanup(func() { cleanup([]uuid.UUID{first.ID(), second.ID()}, []uuid.UUID{eventID}) })
	request := createBaixa(t, eventID, uuid.New(), map[uuid.UUID]int{
		first.ID():  3,
		second.ID(): 2,
	})
	processed, err := repository.BaixarEstoque(context.Background(), request)
	if err != nil || !processed {
		t.Fatalf("baixa falhou: processada=%t erro=%v", processed, err)
	}
	processed, err = repository.BaixarEstoque(context.Background(), request)
	if err != nil || processed {
		t.Fatalf("evento duplicado deveria ser ignorado: processada=%t erro=%v", processed, err)
	}

	firstAfter, _ := repository.BuscarPorID(context.Background(), first.ID())
	secondAfter, _ := repository.BuscarPorID(context.Background(), second.ID())
	if firstAfter.Saldo() != 7 || secondAfter.Saldo() != 6 {
		t.Fatalf("saldos inesperados: primeiro=%d segundo=%d", firstAfter.Saldo(), secondAfter.Saldo())
	}
}

func TestIntegrationBaixaRollsBackEveryItemOnFailure(t *testing.T) {
	repository, cleanup := integrationRepository(t)
	first := createProduct(t, repository, "ROLLBACK-"+uuid.NewString(), 10)
	second := createProduct(t, repository, "ROLLBACK-"+uuid.NewString(), 1)
	eventID := uuid.New()
	t.Cleanup(func() { cleanup([]uuid.UUID{first.ID(), second.ID()}, []uuid.UUID{eventID}) })

	baixa := createBaixa(t, eventID, uuid.New(), map[uuid.UUID]int{
		first.ID():  3,
		second.ID(): 2,
	})
	_, err := repository.BaixarEstoque(context.Background(), baixa)
	if !errors.Is(err, domain.ErrEstoqueInsuficiente) {
		t.Fatalf("esperava estoque insuficiente, recebeu %v", err)
	}
	firstAfter, _ := repository.BuscarPorID(context.Background(), first.ID())
	if firstAfter.Saldo() != 10 {
		t.Fatalf("primeiro item nao sofreu rollback: saldo=%d", firstAfter.Saldo())
	}
}

func TestIntegrationConcurrentBaixaAllowsOnlyOneSuccess(t *testing.T) {
	repository, cleanup := integrationRepository(t)
	produto := createProduct(t, repository, "CONCORRENCIA-"+uuid.NewString(), 1)
	eventIDs := []uuid.UUID{uuid.New(), uuid.New()}
	t.Cleanup(func() { cleanup([]uuid.UUID{produto.ID()}, eventIDs) })

	results := make(chan error, 2)
	var waitGroup sync.WaitGroup
	for index := range 2 {
		waitGroup.Add(1)
		go func(eventID uuid.UUID) {
			defer waitGroup.Done()
			baixa := createBaixa(t, eventID, uuid.New(), map[uuid.UUID]int{produto.ID(): 1})
			_, err := repository.BaixarEstoque(context.Background(), baixa)
			results <- err
		}(eventIDs[index])
	}
	waitGroup.Wait()
	close(results)

	successes, insufficient := 0, 0
	for err := range results {
		if err == nil {
			successes++
		} else if errors.Is(err, domain.ErrEstoqueInsuficiente) {
			insufficient++
		}
	}
	if successes != 1 || insufficient != 1 {
		t.Fatalf("resultado concorrente inesperado: sucessos=%d insuficientes=%d", successes, insufficient)
	}
}

func TestIntegrationCadastroSaldoCreatesAuditMovement(t *testing.T) {
	repository, cleanup := integrationRepository(t)
	produto := createProduct(t, repository, "AJUSTE-"+uuid.NewString(), 5)
	t.Cleanup(func() { cleanup([]uuid.UUID{produto.ID()}, nil) })

	if err := produto.AtualizarDescricao("Produto ajustado"); err != nil {
		t.Fatal(err)
	}
	if err := produto.AtualizarSaldo(9); err != nil {
		t.Fatal(err)
	}
	if err := produto.AtualizarValor(249.90); err != nil {
		t.Fatal(err)
	}
	if err := repository.Atualizar(context.Background(), produto); err != nil {
		t.Fatal(err)
	}

	movimentacoes, err := repository.ListarMovimentacoes(context.Background(), produto.ID())
	if err != nil {
		t.Fatal(err)
	}
	if len(movimentacoes) != 1 || string(movimentacoes[0].Tipo()) != "ENTRADA" || movimentacoes[0].Quantidade() != 4 {
		t.Fatalf("movimentacao de ajuste inesperada: %+v", movimentacoes)
	}
	persistido, err := repository.BuscarPorID(context.Background(), produto.ID())
	if err != nil {
		t.Fatal(err)
	}
	if persistido.Valor() != 249.90 {
		t.Fatalf("esperava valor persistido 249.90, recebeu %.2f", persistido.Valor())
	}
}
