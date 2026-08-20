package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type produtoModel struct {
	ID              uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	Codigo          string    `gorm:"column:codigo"`
	Descricao       string    `gorm:"column:descricao"`
	Saldo           int       `gorm:"column:saldo"`
	DataCadastro    time.Time `gorm:"column:data_cadastro"`
	DataAtualizacao time.Time `gorm:"column:data_atualizacao"`
}

func (produtoModel) TableName() string {
	return "estoque.produtos"
}

type GormProdutoRepository struct {
	database *gorm.DB
}

func NewGormProdutoRepository(database *gorm.DB) *GormProdutoRepository {
	return &GormProdutoRepository{database: database}
}

func (repository *GormProdutoRepository) Criar(ctx context.Context, produto *domain.Produto) error {
	model := toModel(produto)
	if err := repository.database.WithContext(ctx).Create(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrCodigoJaExistente
		}
		return err
	}
	return nil
}

func (repository *GormProdutoRepository) BuscarPorID(ctx context.Context, id uuid.UUID) (*domain.Produto, error) {
	var model produtoModel
	err := repository.database.WithContext(ctx).First(&model, "id = ?", id).Error
	return result(model, err)
}

func (repository *GormProdutoRepository) BuscarPorCodigo(ctx context.Context, codigo string) (*domain.Produto, error) {
	var model produtoModel
	err := repository.database.WithContext(ctx).
		First(&model, "codigo = ?", strings.TrimSpace(codigo)).Error
	return result(model, err)
}

func (repository *GormProdutoRepository) Listar(ctx context.Context) ([]domain.Produto, error) {
	var models []produtoModel
	if err := repository.database.WithContext(ctx).Order("codigo ASC").Find(&models).Error; err != nil {
		return nil, err
	}

	produtos := make([]domain.Produto, 0, len(models))
	for _, model := range models {
		produtos = append(produtos, *toDomain(model))
	}
	return produtos, nil
}

func result(model produtoModel, err error) (*domain.Produto, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrProdutoNaoEncontrado
	}
	if err != nil {
		return nil, err
	}
	return toDomain(model), nil
}

func toModel(produto *domain.Produto) produtoModel {
	return produtoModel{
		ID: produto.ID, Codigo: produto.Codigo, Descricao: produto.Descricao,
		Saldo: produto.Saldo, DataCadastro: produto.DataCadastro,
		DataAtualizacao: produto.DataAtualizacao,
	}
}

func toDomain(model produtoModel) *domain.Produto {
	return &domain.Produto{
		ID: model.ID, Codigo: model.Codigo, Descricao: model.Descricao,
		Saldo: model.Saldo, DataCadastro: model.DataCadastro,
		DataAtualizacao: model.DataAtualizacao,
	}
}
