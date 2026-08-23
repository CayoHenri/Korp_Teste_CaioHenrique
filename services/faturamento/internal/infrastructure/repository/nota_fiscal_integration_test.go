//go:build integration

package repository

import (
	"context"
	"testing"

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
}
