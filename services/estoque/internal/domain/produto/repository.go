package produto

import (
	"context"

	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/movimentacao"
	"github.com/google/uuid"
)

type Repository interface {
	Criar(context.Context, *Produto) error
	Atualizar(context.Context, *Produto) error
	BuscarPorID(context.Context, uuid.UUID) (*Produto, error)
	BuscarPorCodigo(context.Context, string) (*Produto, error)
	Listar(context.Context) ([]Produto, error)
	BaixarEstoque(context.Context, BaixaEstoque) (bool, error)
	ListarMovimentacoes(context.Context, uuid.UUID) ([]movimentacao.Movimentacao, error)
}
