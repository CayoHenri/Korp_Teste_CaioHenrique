package notafiscal

import (
	"context"
	"errors"

	domain "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/domain/nota_fiscal"
	"github.com/google/uuid"
)

const (
	EventTypeBaixaRealizada = "estoque.baixa.realizada"
	EventTypeBaixaRejeitada = "estoque.baixa.rejeitada"
)

var ErrTipoResultadoInvalido = errors.New("tipo do resultado da baixa e invalido")

type ResultadoBaixaInput struct {
	EventID       uuid.UUID
	CorrelationID uuid.UUID
	NotaFiscalID  uuid.UUID
	Type          string
	Motivo        string
}

type resultadoBaixaRepository interface {
	ProcessarResultadoBaixa(
		context.Context,
		uuid.UUID,
		uuid.UUID,
		uuid.UUID,
		func(*domain.NotaFiscal) error,
	) (bool, error)
}

type ProcessarResultadoBaixaUseCase struct {
	repository resultadoBaixaRepository
}

func NewProcessarResultadoBaixaUseCase(
	repository resultadoBaixaRepository,
) *ProcessarResultadoBaixaUseCase {
	return &ProcessarResultadoBaixaUseCase{repository: repository}
}

func (useCase *ProcessarResultadoBaixaUseCase) Execute(
	ctx context.Context,
	input ResultadoBaixaInput,
) (bool, error) {
	var transition func(*domain.NotaFiscal) error
	switch input.Type {
	case EventTypeBaixaRealizada:
		transition = func(nota *domain.NotaFiscal) error {
			return nota.ConfirmarFechamento()
		}
	case EventTypeBaixaRejeitada:
		transition = func(nota *domain.NotaFiscal) error {
			return nota.ReabrirAposRejeicao(input.Motivo)
		}
	default:
		return false, ErrTipoResultadoInvalido
	}
	return useCase.repository.ProcessarResultadoBaixa(
		ctx,
		input.EventID,
		input.CorrelationID,
		input.NotaFiscalID,
		transition,
	)
}
