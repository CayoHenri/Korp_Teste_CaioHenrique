package repository

import (
	"context"
	"errors"
	"time"

	movimentacaoDomain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/movimentacao"
	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/infrastructure/database/models"
	sharedtext "github.com/caiog/korp-notas-fiscais/services/estoque/internal/shared/text"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProdutoRepository struct {
	db *gorm.DB
}

func NewProdutoRepository(db *gorm.DB) *ProdutoRepository {
	return &ProdutoRepository{db: db}
}

func (r *ProdutoRepository) Criar(ctx context.Context, produto *domain.Produto) error {
	model := models.NewProdutoModel(produto)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrCodigoJaExistente
		}
		return err
	}
	return nil
}

func (r *ProdutoRepository) Atualizar(ctx context.Context, produto *domain.Produto) error {
	return r.db.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		produtoPersistido, err := buscarProdutoParaAtualizacao(transaction, produto.ID())
		if err != nil {
			return err
		}

		if err := registrarAjusteSaldo(transaction, produtoPersistido, produto); err != nil {
			return err
		}

		return persistirCadastro(transaction, produto)
	})
}

func buscarProdutoParaAtualizacao(transaction *gorm.DB, produtoID uuid.UUID) (*models.Produto, error) {
	var produto models.Produto
	err := transaction.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&produto, "id = ?", produtoID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrProdutoNaoEncontrado
	}
	if err != nil {
		return nil, err
	}
	return &produto, nil
}

func registrarAjusteSaldo(
	transaction *gorm.DB,
	produtoPersistido *models.Produto,
	produtoAtualizado *domain.Produto,
) error {
	if produtoPersistido.Saldo == produtoAtualizado.Saldo() {
		return nil
	}

	movimentacao, err := movimentacaoDomain.NewAjusteSaldo(
		produtoAtualizado.ID(),
		produtoPersistido.Saldo,
		produtoAtualizado.Saldo(),
	)
	if err != nil {
		return err
	}

	model := models.NewMovimentacaoEstoqueModel(movimentacao)
	return transaction.Create(&model).Error
}

func persistirCadastro(transaction *gorm.DB, produto *domain.Produto) error {
	return transaction.Model(&models.Produto{}).
		Where("id = ?", produto.ID()).
		Updates(map[string]any{
			"descricao":        produto.Descricao(),
			"saldo":            produto.Saldo(),
			"ativo":            produto.Ativo(),
			"data_atualizacao": produto.DataAtualizacao(),
		}).Error
}

func (r *ProdutoRepository) BaixarEstoque(ctx context.Context, baixa domain.BaixaEstoque) (bool, error) {
	processada := false
	err := r.db.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		mensagemNova, err := registrarMensagemProcessada(transaction, baixa.EventID())
		if err != nil {
			return err
		}
		if !mensagemNova {
			return nil
		}
		processada = true

		for _, item := range baixa.Itens() {
			if err := baixarItem(transaction, item, baixa.NotaFiscalID()); err != nil {
				return err
			}
		}
		return nil
	})
	return processada, err
}

func registrarMensagemProcessada(transaction *gorm.DB, eventID uuid.UUID) (bool, error) {
	mensagem := models.MensagemProcessada{
		EventID:     eventID,
		ProcessedAt: time.Now().UTC(),
	}
	result := transaction.Clauses(clause.OnConflict{DoNothing: true}).Create(&mensagem)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func baixarItem(transaction *gorm.DB, item domain.BaixaItem, notaFiscalID uuid.UUID) error {
	result := transaction.Model(&models.Produto{}).
		Where("id = ? AND ativo = TRUE AND saldo >= ?", item.ProdutoID(), item.Quantidade()).
		Updates(map[string]any{
			"saldo":            gorm.Expr("saldo - ?", item.Quantidade()),
			"data_atualizacao": time.Now().UTC(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return validarFalhaNaBaixa(transaction, item)
	}
	return registrarMovimentacaoSaida(transaction, item, notaFiscalID)
}

func validarFalhaNaBaixa(transaction *gorm.DB, item domain.BaixaItem) error {
	var model models.Produto
	if err := transaction.First(&model, "id = ?", item.ProdutoID()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrProdutoNaoEncontrado
		}
		return err
	}

	produto, err := model.ToDomain()
	if err != nil {
		return err
	}
	return produto.ValidarBaixa(item.Quantidade())
}

func registrarMovimentacaoSaida(
	transaction *gorm.DB,
	item domain.BaixaItem,
	notaFiscalID uuid.UUID,
) error {
	movimentacao, err := movimentacaoDomain.NewMovimentacao(
		item.ProdutoID(),
		movimentacaoDomain.TipoSaida,
		item.Quantidade(),
		&notaFiscalID,
	)
	if err != nil {
		return err
	}
	model := models.NewMovimentacaoEstoqueModel(movimentacao)
	return transaction.Create(&model).Error
}

func (r *ProdutoRepository) ListarMovimentacoes(
	ctx context.Context,
	produtoID uuid.UUID,
) ([]movimentacaoDomain.Movimentacao, error) {
	var records []models.MovimentacaoEstoque
	if err := r.db.WithContext(ctx).
		Where("produto_id = ?", produtoID).
		Order("data_movimentacao DESC").Find(&records).Error; err != nil {
		return nil, err
	}

	movimentacoes := make([]movimentacaoDomain.Movimentacao, 0, len(records))
	for _, record := range records {
		movimentacao, err := record.ToDomain()
		if err != nil {
			return nil, err
		}
		movimentacoes = append(movimentacoes, *movimentacao)
	}
	return movimentacoes, nil
}

func (r *ProdutoRepository) BuscarPorID(ctx context.Context, id uuid.UUID) (*domain.Produto, error) {
	var model models.Produto
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	return result(model, err)
}

func (r *ProdutoRepository) BuscarPorCodigo(ctx context.Context, codigo string) (*domain.Produto, error) {
	var model models.Produto
	err := r.db.WithContext(ctx).First(&model, "codigo = ?", sharedtext.NormalizeUpper(codigo)).Error
	return result(model, err)
}

func (r *ProdutoRepository) Listar(ctx context.Context) ([]domain.Produto, error) {
	var records []models.Produto
	if err := r.db.WithContext(ctx).Order("codigo ASC").Find(&records).Error; err != nil {
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
