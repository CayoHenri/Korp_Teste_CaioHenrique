package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	notafiscal "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/domain/nota_fiscal"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/infrastructure/database/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotaFiscalRepository struct {
	db *gorm.DB
}

func NewNotaFiscalRepository(db *gorm.DB) *NotaFiscalRepository {
	return &NotaFiscalRepository{db: db}
}

func (r *NotaFiscalRepository) ProximoNumero(ctx context.Context) (int64, error) {
	var numero int64
	err := r.db.WithContext(ctx).
		Raw("SELECT nextval('faturamento.nota_fiscal_numero_seq')").Scan(&numero).Error
	return numero, err
}

func (r *NotaFiscalRepository) Criar(ctx context.Context, nota *notafiscal.NotaFiscal) error {
	model := models.NewNotaFiscalModel(nota)
	return r.db.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		return transaction.Create(&model).Error
	})
}

func (r *NotaFiscalRepository) Atualizar(ctx context.Context, nota *notafiscal.NotaFiscal) error {
	model := models.NewNotaFiscalModel(nota)
	return r.db.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		result := transaction.Model(&models.NotaFiscal{}).
			Where("id = ? AND status = ?", nota.ID(), string(notafiscal.StatusAberta)).
			Updates(map[string]any{
				"nome_cliente":     nota.NomeCliente(),
				"endereco_cliente": nota.EnderecoCliente(),
				"data_atualizacao": nota.DataAtualizacao(),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return notafiscal.ErrNotaNaoEstaAberta
		}
		if err := transaction.Where("nota_fiscal_id = ?", nota.ID()).
			Delete(&models.ItemNotaFiscal{}).Error; err != nil {
			return err
		}
		if len(model.Itens) == 0 {
			return nil
		}
		return transaction.Create(&model.Itens).Error
	})
}

func (r *NotaFiscalRepository) BuscarPorID(
	ctx context.Context,
	id uuid.UUID,
) (*notafiscal.NotaFiscal, error) {
	var model models.NotaFiscal
	err := r.db.WithContext(ctx).Preload("Itens").First(&model, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, notafiscal.ErrNotaNaoEncontrada
	}
	if err != nil {
		return nil, err
	}
	return model.ToDomain()
}

func (r *NotaFiscalRepository) Listar(ctx context.Context) ([]notafiscal.NotaFiscal, error) {
	var records []models.NotaFiscal
	if err := r.db.WithContext(ctx).
		Preload("Itens").Order("numero DESC").Find(&records).Error; err != nil {
		return nil, err
	}
	notas := make([]notafiscal.NotaFiscal, 0, len(records))
	for index := range records {
		nota, err := records[index].ToDomain()
		if err != nil {
			return nil, err
		}
		notas = append(notas, *nota)
	}
	return notas, nil
}

func (r *NotaFiscalRepository) IniciarFechamento(ctx context.Context, nota *notafiscal.NotaFiscal) error {
	eventID := uuid.New()
	payload, err := criarPayloadBaixa(eventID, nota)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		result := transaction.Model(&models.NotaFiscal{}).
			Where("id = ? AND status = ?", nota.ID(), string(notafiscal.StatusAberta)).
			Updates(map[string]any{
				"status":           nota.Status(),
				"data_atualizacao": nota.DataAtualizacao(),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return notafiscal.ErrNotaNaoEstaAberta
		}
		event := models.OutboxEvent{
			ID:          eventID,
			EventType:   "faturamento.nota.fechamento_solicitado",
			AggregateID: nota.ID(),
			Payload:     payload,
			CreatedAt:   time.Now().UTC()}
		return transaction.Create(&event).Error
	})
}

func criarPayloadBaixa(eventID uuid.UUID, nota *notafiscal.NotaFiscal) ([]byte, error) {
	type itemPayload struct {
		ProdutoID  uuid.UUID `json:"produtoId"`
		Quantidade int       `json:"quantidade"`
	}
	type eventoPayload struct {
		EventID      uuid.UUID     `json:"eventId"`
		NotaFiscalID uuid.UUID     `json:"notaFiscalId"`
		Itens        []itemPayload `json:"itens"`
	}
	itens := nota.Itens()
	payload := eventoPayload{
		EventID:      eventID,
		NotaFiscalID: nota.ID(),
		Itens:        make([]itemPayload, 0, len(itens)),
	}
	for _, item := range itens {
		payload.Itens = append(
			payload.Itens,
			itemPayload{ProdutoID: item.ProdutoID(), Quantidade: item.Quantidade()},
		)
	}
	return json.Marshal(payload)
}
