package movimentacao

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Tipo string

const (
	TipoEntrada Tipo = "ENTRADA"
	TipoSaida   Tipo = "SAIDA"
)

var (
	ErrQuantidadeInvalida = errors.New("quantidade da movimentacao deve ser positiva")
	ErrTipoInvalido       = errors.New("tipo da movimentacao e invalido")
)

type Movimentacao struct {
	id               uuid.UUID
	produtoID        uuid.UUID
	tipo             Tipo
	quantidade       int
	referencia       *uuid.UUID
	dataMovimentacao time.Time
}

func NewMovimentacao(
	produtoID uuid.UUID,
	tipo Tipo,
	quantidade int,
	referencia *uuid.UUID,
) (*Movimentacao, error) {
	return NewMovimentacaoWithState(
		uuid.New(),
		produtoID,
		tipo, quantidade,
		referencia,
		time.Now().UTC(),
	)
}

func NewAjusteSaldo(produtoID uuid.UUID, saldoAnterior, novoSaldo int) (*Movimentacao, error) {
	diferenca := novoSaldo - saldoAnterior
	if diferenca == 0 {
		return nil, nil
	}

	tipo := TipoEntrada
	quantidade := diferenca
	if diferenca < 0 {
		tipo = TipoSaida
		quantidade = -diferenca
	}

	return NewMovimentacao(produtoID, tipo, quantidade, nil)
}

func NewMovimentacaoWithState(
	id uuid.UUID,
	produtoID uuid.UUID,
	tipo Tipo,
	quantidade int,
	referencia *uuid.UUID,
	dataMovimentacao time.Time,
) (*Movimentacao, error) {
	if quantidade <= 0 {
		return nil, ErrQuantidadeInvalida
	}
	if tipo != TipoEntrada && tipo != TipoSaida {
		return nil, ErrTipoInvalido
	}
	return &Movimentacao{
		id:               id,
		produtoID:        produtoID,
		tipo:             tipo,
		quantidade:       quantidade,
		referencia:       referencia,
		dataMovimentacao: dataMovimentacao.UTC(),
	}, nil
}

func (movimentacao *Movimentacao) ID() uuid.UUID {
	return movimentacao.id
}

func (movimentacao *Movimentacao) ProdutoID() uuid.UUID {
	return movimentacao.produtoID
}

func (movimentacao *Movimentacao) Tipo() Tipo {
	return movimentacao.tipo
}

func (movimentacao *Movimentacao) Quantidade() int {
	return movimentacao.quantidade
}

func (movimentacao *Movimentacao) Referencia() *uuid.UUID {
	if movimentacao.referencia == nil {
		return nil
	}
	referencia := *movimentacao.referencia
	return &referencia
}

func (movimentacao *Movimentacao) DataMovimentacao() time.Time {
	return movimentacao.dataMovimentacao
}
