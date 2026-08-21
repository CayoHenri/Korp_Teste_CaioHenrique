package repository

import (
	"context"
	"errors"

	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/infrastructure/database/models"
	sharedtext "github.com/caiog/korp-notas-fiscais/services/estoque/internal/shared/text"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProdutoRepository struct {
	database *gorm.DB
}

func NewProdutoRepository(database *gorm.DB) *ProdutoRepository {
	return &ProdutoRepository{database: database}
}

func (repository *ProdutoRepository) Criar(ctx context.Context, produto *domain.Produto) error {
	model := models.NewProdutoModel(produto)
	if err := repository.database.WithContext(ctx).Create(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrCodigoJaExistente
		}
		return err
	}
	return nil
}

func (repository *ProdutoRepository) Atualizar(ctx context.Context, produto *domain.Produto) error {
	result := repository.database.WithContext(ctx).
		Model(&models.Produto{}).
		Where("id = ?", produto.ID()).
		Updates(map[string]any{
			"descricao":        produto.Descricao(),
			"saldo":            produto.Saldo(),
			"ativo":            produto.Ativo(),
			"data_atualizacao": produto.DataAtualizacao(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrProdutoNaoEncontrado
	}
	return nil
}

func (repository *ProdutoRepository) BuscarPorID(ctx context.Context, id uuid.UUID) (*domain.Produto, error) {
	var model models.Produto
	err := repository.database.WithContext(ctx).First(&model, "id = ?", id).Error
	return result(model, err)
}

func (repository *ProdutoRepository) BuscarPorCodigo(ctx context.Context, codigo string) (*domain.Produto, error) {
	var model models.Produto
	err := repository.database.WithContext(ctx).
		First(&model, "codigo = ?", sharedtext.NormalizeUpper(codigo)).Error
	return result(model, err)
}

func (repository *ProdutoRepository) Listar(ctx context.Context) ([]domain.Produto, error) {
	var records []models.Produto
	if err := repository.database.WithContext(ctx).Order("codigo ASC").Find(&records).Error; err != nil {
		return nil, err
	}

	produtos := make([]domain.Produto, 0, len(records))
	for _, model := range records {
		produto, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		produtos = append(produtos, *produto)
	}
	return produtos, nil
}

func result(model models.Produto, err error) (*domain.Produto, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrProdutoNaoEncontrado
	}
	if err != nil {
		return nil, err
	}
	return model.ToDomain()
}
