package produto

import (
	"context"

	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/movimentacao"
	sharedquery "github.com/caiog/korp-notas-fiscais/services/estoque/internal/shared/query"
	"github.com/google/uuid"
)

type ListFilters struct {
	Codigo    string
	Descricao string
	Ativo     *bool
}

type Repository interface {
	Criar(context.Context, *Produto) error
	Atualizar(context.Context, *Produto) error
	BuscarPorID(context.Context, uuid.UUID) (*Produto, error)
	BuscarPorCodigo(context.Context, string) (*Produto, error)
	Listar(context.Context, sharedquery.Criteria[ListFilters]) (sharedquery.Page[Produto], error)
	BaixarEstoque(context.Context, BaixaEstoque) (bool, error)
	ListarMovimentacoes(context.Context, uuid.UUID) ([]movimentacao.Movimentacao, error)
}
