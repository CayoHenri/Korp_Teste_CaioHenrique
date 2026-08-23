package produto

import (
	"errors"
	"time"

	sharedtext "github.com/caiog/korp-notas-fiscais/services/estoque/internal/shared/text"
	"github.com/google/uuid"
)

var (
	ErrCodigoObrigatorio    = errors.New("codigo do produto e obrigatorio")
	ErrDescricaoObrigatoria = errors.New("descricao do produto e obrigatoria")
	ErrSaldoInvalido        = errors.New("saldo do produto nao pode ser negativo")
	ErrIDInvalido           = errors.New("id do produto e invalido")
	ErrProdutoNaoEncontrado = errors.New("produto nao encontrado")
	ErrCodigoJaExistente    = errors.New("codigo do produto ja existente")
	ErrProdutoInativo       = errors.New("produto esta inativo")
	ErrEstoqueInsuficiente  = errors.New("estoque insuficiente")
	ErrQuantidadeInvalida   = errors.New("quantidade deve ser positiva")
	ErrValorInvalido        = errors.New("valor unitario deve ser maior que zero")
)

type Produto struct {
	id              uuid.UUID
	codigo          string
	descricao       string
	saldo           int
	valor           float64
	ativo           bool
	dataCadastro    time.Time
	dataAtualizacao time.Time
}

func NewProduto(codigo, descricao string, saldo int, valor float64) (*Produto, error) {
	agora := time.Now().UTC()
	return NewProdutoWithState(uuid.New(), codigo, descricao, saldo, valor, true, agora, agora)
}

func NewProdutoWithState(
	id uuid.UUID,
	codigo string,
	descricao string,
	saldo int,
	valor float64,
	ativo bool,
	dataCadastro time.Time,
	dataAtualizacao time.Time,
) (*Produto, error) {
	codigo = sharedtext.NormalizeUpper(codigo)
	descricao = sharedtext.NormalizeUpper(descricao)

	if id == uuid.Nil {
		return nil, ErrIDInvalido
	}
	if codigo == "" {
		return nil, ErrCodigoObrigatorio
	}
	if descricao == "" {
		return nil, ErrDescricaoObrigatoria
	}
	if saldo < 0 {
		return nil, ErrSaldoInvalido
	}
	if valor <= 0 {
		return nil, ErrValorInvalido
	}

	return &Produto{
		id:              id,
		codigo:          codigo,
		descricao:       descricao,
		saldo:           saldo,
		valor:           valor,
		ativo:           ativo,
		dataCadastro:    dataCadastro.UTC(),
		dataAtualizacao: dataAtualizacao.UTC(),
	}, nil
}

func (produto *Produto) ID() uuid.UUID {
	return produto.id
}

func (produto *Produto) Codigo() string {
	return produto.codigo
}

func (produto *Produto) Descricao() string {
	return produto.descricao
}

func (produto *Produto) Saldo() int {
	return produto.saldo
}

func (produto *Produto) Valor() float64 {
	return produto.valor
}

func (produto *Produto) Ativo() bool {
	return produto.ativo
}

func (produto *Produto) DataCadastro() time.Time {
	return produto.dataCadastro
}

func (produto *Produto) DataAtualizacao() time.Time {
	return produto.dataAtualizacao
}

func (produto *Produto) AtualizarDescricao(descricao string) error {
	descricao = sharedtext.NormalizeUpper(descricao)
	if descricao == "" {
		return ErrDescricaoObrigatoria
	}
	if descricao == produto.descricao {
		return nil
	}

	produto.descricao = descricao
	produto.dataAtualizacao = time.Now().UTC()
	return nil
}

func (produto *Produto) AtualizarSaldo(saldo int) error {
	if saldo < 0 {
		return ErrSaldoInvalido
	}
	if saldo == produto.saldo {
		return nil
	}

	produto.saldo = saldo
	produto.dataAtualizacao = time.Now().UTC()
	return nil
}

func (produto *Produto) AtualizarValor(valor float64) error {
	if valor <= 0 {
		return ErrValorInvalido
	}
	if valor == produto.valor {
		return nil
	}
	produto.valor = valor
	produto.dataAtualizacao = time.Now().UTC()
	return nil
}

func (produto *Produto) ValidarBaixa(quantidade int) error {
	if quantidade <= 0 {
		return ErrQuantidadeInvalida
	}
	if !produto.ativo {
		return ErrProdutoInativo
	}
	if produto.saldo < quantidade {
		return ErrEstoqueInsuficiente
	}
	return nil
}

func (produto *Produto) Ativar() {
	if produto.ativo {
		return
	}
	produto.ativo = true
	produto.dataAtualizacao = time.Now().UTC()
}

func (produto *Produto) Inativar() {
	if !produto.ativo {
		return
	}
	produto.ativo = false
	produto.dataAtualizacao = time.Now().UTC()
}
