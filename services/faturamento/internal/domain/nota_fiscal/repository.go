package notafiscal

import (
	"context"

	sharedquery "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/shared/query"
	"github.com/google/uuid"
)

type ListFilters struct {
	Numero      *int64
	Status      *Status
	NomeCliente string
}

type Repository interface {
	ProximoNumero(context.Context) (int64, error)
	Criar(context.Context, *NotaFiscal) error
	Atualizar(context.Context, *NotaFiscal) error
	BuscarPorID(context.Context, uuid.UUID) (*NotaFiscal, error)
	Listar(context.Context, sharedquery.Criteria[ListFilters]) (sharedquery.Page[NotaFiscal], error)
	IniciarFechamento(context.Context, *NotaFiscal) error
}
