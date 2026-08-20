package produto

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Criar(context.Context, *Produto) error
	BuscarPorID(context.Context, uuid.UUID) (*Produto, error)
	BuscarPorCodigo(context.Context, string) (*Produto, error)
	Listar(context.Context) ([]Produto, error)
}
