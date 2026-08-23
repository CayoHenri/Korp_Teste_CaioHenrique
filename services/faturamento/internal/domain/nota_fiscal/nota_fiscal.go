package notafiscal

import (
	"errors"
	"time"

	sharedtext "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/shared/text"
	"github.com/google/uuid"
)

type Status string

const (
	StatusAberta      Status = "ABERTA"
	StatusProcessando Status = "PROCESSANDO"
	StatusFechada     Status = "FECHADA"
)

var (
	ErrIDInvalido                 = errors.New("id da nota fiscal e invalido")
	ErrNumeroInvalido             = errors.New("numero da nota fiscal e invalido")
	ErrStatusInvalido             = errors.New("status da nota fiscal e invalido")
	ErrNotaSemItens               = errors.New("nota fiscal deve possuir ao menos um item")
	ErrNotaNaoEncontrada          = errors.New("nota fiscal nao encontrada")
	ErrNotaNaoEstaAberta          = errors.New("nota fiscal nao esta aberta")
	ErrNotaNaoEstaProcessando     = errors.New("nota fiscal nao esta processando")
	ErrNomeClienteObrigatorio     = errors.New("nome do cliente e obrigatorio")
	ErrEnderecoClienteObrigatorio = errors.New("endereco do cliente e obrigatorio")
)

type NotaFiscal struct {
	id              uuid.UUID
	numero          int64
	status          Status
	nomeCliente     string
	enderecoCliente string
	itens           []ItemNotaFiscal
	dataCadastro    time.Time
	dataAtualizacao time.Time
	dataFechamento  *time.Time
}

func NewNotaFiscal(
	numero int64,
	nomeCliente, enderecoCliente string,
	itens []ItemNotaFiscal,
) (*NotaFiscal, error) {
	agora := time.Now().UTC()
	return NewNotaFiscalWithState(
		uuid.New(),
		numero,
		StatusAberta,
		nomeCliente,
		enderecoCliente,
		itens,
		agora,
		agora,
		nil,
	)
}

func NewNotaFiscalWithState(
	id uuid.UUID,
	numero int64,
	status Status,
	nomeCliente, enderecoCliente string,
	itens []ItemNotaFiscal,
	cadastro, atualizacao time.Time,
	fechamento *time.Time,
) (*NotaFiscal, error) {
	if id == uuid.Nil {
		return nil, ErrIDInvalido
	}
	if numero <= 0 {
		return nil, ErrNumeroInvalido
	}
	if !status.Valido() {
		return nil, ErrStatusInvalido
	}
	nomeCliente = sharedtext.NormalizeUpper(nomeCliente)
	enderecoCliente = sharedtext.NormalizeUpper(enderecoCliente)
	if nomeCliente == "" {
		return nil, ErrNomeClienteObrigatorio
	}
	if enderecoCliente == "" {
		return nil, ErrEnderecoClienteObrigatorio
	}
	return &NotaFiscal{
		id:              id,
		numero:          numero,
		status:          status,
		nomeCliente:     nomeCliente,
		enderecoCliente: enderecoCliente,
		itens:           append([]ItemNotaFiscal(nil), itens...),
		dataCadastro:    cadastro.UTC(),
		dataAtualizacao: atualizacao.UTC(),
		dataFechamento:  cloneTime(fechamento),
	}, nil
}

func (status Status) Valido() bool {
	return status == StatusAberta || status == StatusProcessando || status == StatusFechada
}

func (nota *NotaFiscal) ID() uuid.UUID {
	return nota.id
}

func (nota *NotaFiscal) Numero() int64 {
	return nota.numero
}

func (nota *NotaFiscal) Status() Status {
	return nota.status
}

func (nota *NotaFiscal) NomeCliente() string {
	return nota.nomeCliente
}

func (nota *NotaFiscal) EnderecoCliente() string {
	return nota.enderecoCliente
}

func (nota *NotaFiscal) QuantidadeTotal() int {
	total := 0
	for _, item := range nota.itens {
		total += item.Quantidade()
	}
	return total
}

func (nota *NotaFiscal) ValorTotal() float64 {
	var total float64
	for _, item := range nota.itens {
		total += item.ValorTotal()
	}
	return total
}

func (nota *NotaFiscal) Itens() []ItemNotaFiscal {
	return append([]ItemNotaFiscal(nil), nota.itens...)
}

func (nota *NotaFiscal) DataCadastro() time.Time {
	return nota.dataCadastro
}

func (nota *NotaFiscal) DataAtualizacao() time.Time {
	return nota.dataAtualizacao
}

func (nota *NotaFiscal) DataFechamento() *time.Time {
	return cloneTime(nota.dataFechamento)
}

func (nota *NotaFiscal) ValidarEdicao() error {
	if nota.status != StatusAberta {
		return ErrNotaNaoEstaAberta
	}
	return nil
}

func (nota *NotaFiscal) Atualizar(
	nomeCliente, enderecoCliente string,
	itens []ItemNotaFiscal,
) error {
	if err := nota.ValidarEdicao(); err != nil {
		return err
	}
	if len(itens) == 0 {
		return ErrNotaSemItens
	}
	nomeCliente = sharedtext.NormalizeUpper(nomeCliente)
	enderecoCliente = sharedtext.NormalizeUpper(enderecoCliente)
	if nomeCliente == "" {
		return ErrNomeClienteObrigatorio
	}
	if enderecoCliente == "" {
		return ErrEnderecoClienteObrigatorio
	}
	nota.nomeCliente = nomeCliente
	nota.enderecoCliente = enderecoCliente
	nota.itens = append([]ItemNotaFiscal(nil), itens...)
	nota.dataAtualizacao = time.Now().UTC()
	return nil
}

func (nota *NotaFiscal) IniciarFechamento() error {
	if nota.status != StatusAberta {
		return ErrNotaNaoEstaAberta
	}
	if len(nota.itens) == 0 {
		return ErrNotaSemItens
	}
	nota.status = StatusProcessando
	nota.dataAtualizacao = time.Now().UTC()
	return nil
}

func (nota *NotaFiscal) ConfirmarFechamento() error {
	if nota.status != StatusProcessando {
		return ErrNotaNaoEstaProcessando
	}
	agora := time.Now().UTC()
	nota.status = StatusFechada
	nota.dataAtualizacao = agora
	nota.dataFechamento = &agora
	return nil
}

func (nota *NotaFiscal) ReabrirAposRejeicao() error {
	if nota.status != StatusProcessando {
		return ErrNotaNaoEstaProcessando
	}
	nota.status = StatusAberta
	nota.dataAtualizacao = time.Now().UTC()
	return nil
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := value.UTC()
	return &cloned
}
