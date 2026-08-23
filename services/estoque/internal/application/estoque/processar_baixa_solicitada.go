package estoque

import (
	"context"
	"errors"

	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
	"github.com/google/uuid"
)

const (
	EventTypeBaixaSolicitada = "estoque.baixa.solicitada"
	EventTypeBaixaRealizada  = "estoque.baixa.realizada"
	EventTypeBaixaRejeitada  = "estoque.baixa.rejeitada"
)

type ResultadoBaixa struct {
	EventID       uuid.UUID
	CorrelationID uuid.UUID
	NotaFiscalID  uuid.UUID
	Type          string
	Motivo        string
}

type resultadoPublisher interface {
	PublicarResultado(context.Context, ResultadoBaixa) error
}

type ProcessarBaixaSolicitadaUseCase struct {
	baixarEstoque *BaixarEstoqueUseCase
	publisher     resultadoPublisher
}

func NewProcessarBaixaSolicitadaUseCase(
	baixarEstoque *BaixarEstoqueUseCase,
	publisher resultadoPublisher,
) *ProcessarBaixaSolicitadaUseCase {
	return &ProcessarBaixaSolicitadaUseCase{
		baixarEstoque: baixarEstoque,
		publisher:     publisher,
	}
}

func (useCase *ProcessarBaixaSolicitadaUseCase) Execute(
	ctx context.Context,
	input BaixarEstoqueInput,
) error {
	_, err := useCase.baixarEstoque.Execute(ctx, input)
	resultado := ResultadoBaixa{
		EventID:       uuid.New(),
		CorrelationID: input.EventID,
		NotaFiscalID:  input.NotaFiscalID,
		Type:          EventTypeBaixaRealizada,
	}
	if err != nil {
		motivo, rejeitada := motivoRejeicao(err)
		if !rejeitada {
			return err
		}
		resultado.Type = EventTypeBaixaRejeitada
		resultado.Motivo = motivo
	}
	return useCase.publisher.PublicarResultado(ctx, resultado)
}

func motivoRejeicao(err error) (string, bool) {
	switch {
	case errors.Is(err, domain.ErrEstoqueInsuficiente):
		return "ESTOQUE_INSUFICIENTE", true
	case errors.Is(err, domain.ErrProdutoNaoEncontrado):
		return "PRODUTO_NAO_ENCONTRADO", true
	case errors.Is(err, domain.ErrProdutoInativo):
		return "PRODUTO_INATIVO", true
	case errors.Is(err, domain.ErrQuantidadeInvalida),
		errors.Is(err, domain.ErrEventoInvalido),
		errors.Is(err, domain.ErrNotaInvalida),
		errors.Is(err, domain.ErrItensVazios):
		return "SOLICITACAO_INVALIDA", true
	default:
		return "", false
	}
}
