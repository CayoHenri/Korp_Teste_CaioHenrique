package notafiscal

import (
	"context"
	domain "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/domain/nota_fiscal"
	"github.com/google/uuid"
)

type IniciarFechamentoUseCase struct{ repository domain.Repository }

func NewIniciarFechamentoUseCase(repository domain.Repository) *IniciarFechamentoUseCase {
	return &IniciarFechamentoUseCase{repository: repository}
}
func (useCase *IniciarFechamentoUseCase) Execute(
	ctx context.Context,
	id uuid.UUID,
) (*domain.NotaFiscal, error) {
	nota, err := useCase.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := nota.IniciarFechamento(); err != nil {
		return nil, err
	}
	if err := useCase.repository.IniciarFechamento(ctx, nota); err != nil {
		return nil, err
	}
	return nota, nil
}
