package notafiscal

import (
	"context"
	"testing"

	domain "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/domain/nota_fiscal"
	"github.com/google/uuid"
)

type resultadoRepositoryStub struct {
	nota          *domain.NotaFiscal
	processed     bool
	correlationID uuid.UUID
}

func (repository *resultadoRepositoryStub) ProcessarResultadoBaixa(
	_ context.Context,
	_, correlationID, _ uuid.UUID,
	transition func(*domain.NotaFiscal) error,
) (bool, error) {
	repository.correlationID = correlationID
	if !repository.processed {
		return false, nil
	}
	return true, transition(repository.nota)
}

func TestProcessarResultadoBaixaFechaNotaNoSucesso(t *testing.T) {
	item, _ := domain.NewItemNotaFiscal(uuid.New(), "SKU", "Produto", 1, 10)
	nota, _ := domain.NewNotaFiscal(1, "Cliente", "Rua", []domain.ItemNotaFiscal{*item})
	_ = nota.IniciarFechamento()
	repository := &resultadoRepositoryStub{nota: nota, processed: true}
	correlationID := uuid.New()

	processed, err := NewProcessarResultadoBaixaUseCase(repository).Execute(
		context.Background(),
		ResultadoBaixaInput{
			EventID:       uuid.New(),
			CorrelationID: correlationID,
			NotaFiscalID:  nota.ID(),
			Type:          EventTypeBaixaRealizada,
		},
	)
	if err != nil || !processed || nota.Status() != domain.StatusFechada {
		t.Fatalf("resultado inesperado: processed=%v status=%s err=%v", processed, nota.Status(), err)
	}
	if repository.correlationID != correlationID {
		t.Fatal("correlationId nao foi usado como chave de idempotencia")
	}
}

func TestProcessarResultadoBaixaReabreNotaNaRejeicao(t *testing.T) {
	item, _ := domain.NewItemNotaFiscal(uuid.New(), "SKU", "Produto", 1, 10)
	nota, _ := domain.NewNotaFiscal(1, "Cliente", "Rua", []domain.ItemNotaFiscal{*item})
	_ = nota.IniciarFechamento()
	repository := &resultadoRepositoryStub{nota: nota, processed: true}

	_, err := NewProcessarResultadoBaixaUseCase(repository).Execute(
		context.Background(),
		ResultadoBaixaInput{
			EventID:       uuid.New(),
			CorrelationID: uuid.New(),
			NotaFiscalID:  nota.ID(),
			Type:          EventTypeBaixaRejeitada,
			Motivo:        "ESTOQUE_INSUFICIENTE",
		},
	)
	if err != nil || nota.Status() != domain.StatusAberta {
		t.Fatalf("resultado inesperado: status=%s err=%v", nota.Status(), err)
	}
	if nota.MotivoRejeicao() != "ESTOQUE_INSUFICIENTE" {
		t.Fatalf("motivo inesperado: %s", nota.MotivoRejeicao())
	}
}
