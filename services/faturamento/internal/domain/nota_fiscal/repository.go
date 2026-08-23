package notafiscal

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	ProximoNumero(context.Context) (int64, error)
	Criar(context.Context, *NotaFiscal) error
	Atualizar(context.Context, *NotaFiscal) error
	BuscarPorID(context.Context, uuid.UUID) (*NotaFiscal, error)
	Listar(context.Context) ([]NotaFiscal, error)
	IniciarFechamento(context.Context, *NotaFiscal) error
}
