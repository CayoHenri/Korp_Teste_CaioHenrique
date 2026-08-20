package produto

import (
	"context"

	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
	"github.com/google/uuid"
)

type CriarInput struct {
	Codigo    string
	Descricao string
	Saldo     int
}

type Service struct {
	repository domain.Repository
}

func NewService(repository domain.Repository) *Service {
	return &Service{repository: repository}
}

func (service *Service) Criar(ctx context.Context, input CriarInput) (*domain.Produto, error) {
	produto, err := domain.Novo(input.Codigo, input.Descricao, input.Saldo)
	if err != nil {
		return nil, err
	}

	if err := service.repository.Criar(ctx, produto); err != nil {
		return nil, err
	}

	return produto, nil
}

func (service *Service) BuscarPorID(ctx context.Context, id uuid.UUID) (*domain.Produto, error) {
	return service.repository.BuscarPorID(ctx, id)
}

func (service *Service) BuscarPorCodigo(ctx context.Context, codigo string) (*domain.Produto, error) {
	return service.repository.BuscarPorCodigo(ctx, codigo)
}

func (service *Service) Listar(ctx context.Context) ([]domain.Produto, error) {
	return service.repository.Listar(ctx)
}
