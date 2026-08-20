package produto

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCodigoObrigatorio    = errors.New("codigo do produto e obrigatorio")
	ErrDescricaoObrigatoria = errors.New("descricao do produto e obrigatoria")
	ErrSaldoInvalido        = errors.New("saldo do produto nao pode ser negativo")
	ErrProdutoNaoEncontrado = errors.New("produto nao encontrado")
	ErrCodigoJaExistente    = errors.New("codigo do produto ja existente")
)

type Produto struct {
	ID              uuid.UUID
	Codigo          string
	Descricao       string
	Saldo           int
	DataCadastro    time.Time
	DataAtualizacao time.Time
}

func Novo(codigo, descricao string, saldo int) (*Produto, error) {
	codigo = strings.TrimSpace(codigo)
	descricao = strings.TrimSpace(descricao)

	if codigo == "" {
		return nil, ErrCodigoObrigatorio
	}
	if descricao == "" {
		return nil, ErrDescricaoObrigatoria
	}
	if saldo < 0 {
		return nil, ErrSaldoInvalido
	}

	agora := time.Now().UTC()
	return &Produto{
		ID:              uuid.New(),
		Codigo:          codigo,
		Descricao:       descricao,
		Saldo:           saldo,
		DataCadastro:    agora,
		DataAtualizacao: agora,
	}, nil
}
