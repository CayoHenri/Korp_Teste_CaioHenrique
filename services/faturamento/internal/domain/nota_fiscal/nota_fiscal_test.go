package notafiscal

import (
	"errors"
	"github.com/google/uuid"
	"testing"
)

func itemValido(t *testing.T) ItemNotaFiscal {
	t.Helper()
	item, err := NewItemNotaFiscal(uuid.New(), " sku-1 ", " teclado ", 2, 15.50)
	if err != nil {
		t.Fatal(err)
	}
	return *item
}

func TestNewNotaFiscalIniciaAberta(t *testing.T) {
	nota, err := NewNotaFiscal(
		1,
		"  Maria da Silva  ",
		"  Rua das Flores, 100  ",
		[]ItemNotaFiscal{itemValido(t)},
	)
	if err != nil {
		t.Fatal(err)
	}
	if nota.Status() != StatusAberta {
		t.Fatalf("esperava ABERTA, recebeu %s", nota.Status())
	}
	if nota.Numero() != 1 {
		t.Fatalf("esperava numero 1, recebeu %d", nota.Numero())
	}
	if nota.NomeCliente() != "MARIA DA SILVA" {
		t.Fatalf("nome do cliente nao foi normalizado: %q", nota.NomeCliente())
	}
	if nota.EnderecoCliente() != "RUA DAS FLORES, 100" {
		t.Fatalf("endereco do cliente nao foi normalizado: %q", nota.EnderecoCliente())
	}
}

func TestItemNotaFiscalNormalizaSnapshot(t *testing.T) {
	item := itemValido(t)
	if item.CodigoProduto() != "SKU-1" || item.DescricaoProduto() != "TECLADO" {
		t.Fatalf("snapshot nao normalizado: %s - %s", item.CodigoProduto(), item.DescricaoProduto())
	}
}

func TestNotaFiscalCalculaTotais(t *testing.T) {
	nota, err := NewNotaFiscal(1, "Cliente", "Rua A", []ItemNotaFiscal{itemValido(t)})
	if err != nil {
		t.Fatal(err)
	}
	if nota.QuantidadeTotal() != 2 || nota.ValorTotal() != 31.00 {
		t.Fatalf("totais inesperados: quantidade=%d valor=%.2f", nota.QuantidadeTotal(), nota.ValorTotal())
	}
}

func TestAtualizarNotaFiscalAberta(t *testing.T) {
	nota, _ := NewNotaFiscal(1, "Cliente", "Rua A", []ItemNotaFiscal{itemValido(t)})
	novoItem, _ := NewItemNotaFiscal(uuid.New(), "sku-2", "mouse", 3, 20)

	if err := nota.Atualizar(" novo cliente ", " avenida central ", []ItemNotaFiscal{*novoItem}); err != nil {
		t.Fatal(err)
	}
	if nota.NomeCliente() != "NOVO CLIENTE" || nota.EnderecoCliente() != "AVENIDA CENTRAL" {
		t.Fatal("dados do cliente nao foram atualizados e normalizados")
	}
	if nota.QuantidadeTotal() != 3 || nota.ValorTotal() != 60 {
		t.Fatal("itens e totalizadores nao foram atualizados")
	}
}

func TestAtualizarNotaFiscalRejeitaStatusDiferenteDeAberta(t *testing.T) {
	nota, _ := NewNotaFiscal(1, "Cliente", "Rua A", []ItemNotaFiscal{itemValido(t)})
	_ = nota.IniciarFechamento()
	nomeOriginal := nota.NomeCliente()

	err := nota.Atualizar("Outro cliente", "Outra rua", []ItemNotaFiscal{itemValido(t)})
	if !errors.Is(err, ErrNotaNaoEstaAberta) {
		t.Fatalf("esperava nota nao aberta, recebeu %v", err)
	}
	if nota.NomeCliente() != nomeOriginal {
		t.Fatal("nota nao aberta foi modificada")
	}
}

func TestCicloDeFechamento(t *testing.T) {
	nota, _ := NewNotaFiscal(1, "Cliente", "Rua A", []ItemNotaFiscal{itemValido(t)})
	if err := nota.IniciarFechamento(); err != nil {
		t.Fatal(err)
	}
	if nota.Status() != StatusProcessando {
		t.Fatalf("esperava PROCESSANDO, recebeu %s", nota.Status())
	}
	if err := nota.ConfirmarFechamento(); err != nil {
		t.Fatal(err)
	}
	if nota.Status() != StatusFechada || nota.DataFechamento() == nil {
		t.Fatal("nota deveria estar fechada e possuir data de fechamento")
	}
	if err := nota.IniciarFechamento(); !errors.Is(err, ErrNotaNaoEstaAberta) {
		t.Fatalf("esperava nota nao aberta, recebeu %v", err)
	}
}

func TestFechamentoExigeItens(t *testing.T) {
	nota, _ := NewNotaFiscal(1, "Cliente", "Rua A", nil)
	if err := nota.IniciarFechamento(); !errors.Is(err, ErrNotaSemItens) {
		t.Fatalf("esperava nota sem itens, recebeu %v", err)
	}
}

func TestReabrirAposRejeicao(t *testing.T) {
	nota, _ := NewNotaFiscal(1, "Cliente", "Rua A", []ItemNotaFiscal{itemValido(t)})
	_ = nota.IniciarFechamento()
	if err := nota.ReabrirAposRejeicao(); err != nil {
		t.Fatal(err)
	}
	if nota.Status() != StatusAberta {
		t.Fatalf("esperava ABERTA, recebeu %s", nota.Status())
	}
}
