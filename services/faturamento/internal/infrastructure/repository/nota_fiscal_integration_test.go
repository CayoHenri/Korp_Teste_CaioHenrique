//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	domain "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/domain/nota_fiscal"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/infrastructure/config"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/infrastructure/database"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/infrastructure/database/models"
	"github.com/google/uuid"
)

func TestIntegrationCriarEIniciarFechamento(t *testing.T) {
	databaseURL, err := config.DatabaseURL()
	if err != nil {
		t.Fatal(err)
	}
	connection, err := database.Open(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()

	repository := NewNotaFiscalRepository(connection.Gorm)
	item, err := domain.NewItemNotaFiscal(uuid.New(), "SKU-INT", "Produto integrado", 2, 35.90)
	if err != nil {
		t.Fatal(err)
	}
	numero, err := repository.ProximoNumero(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	nota, err := domain.NewNotaFiscal(numero, "Cliente Integrado", "Rua Integrada, 10", []domain.ItemNotaFiscal{*item})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		connection.Gorm.Where("aggregate_id = ?", nota.ID()).Delete(&models.OutboxEvent{})
		connection.Gorm.Delete(&models.NotaFiscal{}, "id = ?", nota.ID())
	})

	if err := repository.Criar(context.Background(), nota); err != nil {
		t.Fatal(err)
	}
	persistida, err := repository.BuscarPorID(context.Background(), nota.ID())
	if err != nil {
		t.Fatal(err)
	}
	if len(persistida.Itens()) != 1 || persistida.Itens()[0].Quantidade() != 2 {
		t.Fatalf("itens persistidos incorretamente: %+v", persistida.Itens())
	}
	novoItem, err := domain.NewItemNotaFiscal(
		uuid.New(),
		"SKU-ATUALIZADO",
		"Produto atualizado",
		3,
		40,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := nota.Atualizar(
		"Cliente Atualizado",
		"Avenida Atualizada, 20",
		[]domain.ItemNotaFiscal{*novoItem},
	); err != nil {
		t.Fatal(err)
	}
	if err := repository.Atualizar(context.Background(), nota); err != nil {
		t.Fatal(err)
	}
	persistida, err = repository.BuscarPorID(context.Background(), nota.ID())
	if err != nil {
		t.Fatal(err)
	}
	if persistida.NomeCliente() != "CLIENTE ATUALIZADO" ||
		len(persistida.Itens()) != 1 ||
		persistida.Itens()[0].CodigoProduto() != "SKU-ATUALIZADO" {
		t.Fatal("cabecalho e itens da nota nao foram substituidos")
	}

	if err := nota.IniciarFechamento(); err != nil {
		t.Fatal(err)
	}
	if err := repository.IniciarFechamento(context.Background(), nota); err != nil {
		t.Fatal(err)
	}
	var events int64
	if err := connection.Gorm.Model(&models.OutboxEvent{}).
		Where("aggregate_id = ?", nota.ID()).Count(&events).Error; err != nil {
		t.Fatal(err)
	}
	if events != 1 {
		t.Fatalf("esperava um evento Outbox, recebeu %d", events)
	}
	outboxRepository := NewOutboxRepository(connection.Gorm)
	pendentes, err := outboxRepository.ListarPendentes(context.Background(), 100)
	if err != nil {
		t.Fatal(err)
	}
	var eventID uuid.UUID
	for _, event := range pendentes {
		if event.ID == uuid.Nil || event.Type != "estoque.baixa.solicitada" {
			continue
		}
		var record models.OutboxEvent
		if err := connection.Gorm.First(&record, "id = ?", event.ID).Error; err == nil &&
			record.AggregateID == nota.ID() {
			eventID = event.ID
			break
		}
	}
	if eventID == uuid.Nil {
		t.Fatal("evento da nota nao foi encontrado entre os pendentes")
	}
	t.Cleanup(func() {
		connection.Gorm.Delete(&models.MensagemProcessada{}, "correlation_id = ?", eventID)
	})
	if err := outboxRepository.MarcarPublicado(
		context.Background(),
		eventID,
		time.Now().UTC(),
	); err != nil {
		t.Fatal(err)
	}
	var published models.OutboxEvent
	if err := connection.Gorm.First(&published, "id = ?", eventID).Error; err != nil {
		t.Fatal(err)
	}
	if published.PublishedAt == nil {
		t.Fatal("evento deveria estar marcado como publicado")
	}
	processed, err := repository.ProcessarResultadoBaixa(
		context.Background(),
		uuid.New(),
		eventID,
		nota.ID(),
		func(nota *domain.NotaFiscal) error { return nota.ConfirmarFechamento() },
	)
	if err != nil || !processed {
		t.Fatalf("resultado deveria fechar a nota: processed=%v err=%v", processed, err)
	}
	processed, err = repository.ProcessarResultadoBaixa(
		context.Background(),
		uuid.New(),
		eventID,
		nota.ID(),
		func(nota *domain.NotaFiscal) error { return nota.ConfirmarFechamento() },
	)
	if err != nil || processed {
		t.Fatalf("resultado duplicado deveria ser ignorado: processed=%v err=%v", processed, err)
	}
	finalizada, err := repository.BuscarPorID(context.Background(), nota.ID())
	if err != nil {
		t.Fatal(err)
	}
	if finalizada.Status() != domain.StatusFechada {
		t.Fatalf("nota deveria estar fechada: status=%v", finalizada.Status())
	}
}
