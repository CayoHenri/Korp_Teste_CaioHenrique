import { registerLocaleData } from '@angular/common';
import localePt from '@angular/common/locales/pt';
import { TestBed } from '@angular/core/testing';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { of } from 'rxjs';
import { ProdutoHttpService } from '../../produtos/produto-http.service';
import { Produto } from '../../produtos/produto.model';
import { NotaFiscal } from '../nota-fiscal.model';
import { NotaFiscalFormDialog } from './nota-fiscal-form-dialog';

registerLocaleData(localePt);

const ativo: Produto = {
  id: 'produto-ativo', codigo: 'SKU-ATIVO', descricao: 'PRODUTO ATIVO', saldo: 8,
  valor: 25.5, ativo: true, dataCadastro: '2026-08-24T10:00:00Z',
  dataAtualizacao: '2026-08-24T10:00:00Z',
};

describe('NotaFiscalFormDialog', () => {
  const dialogRef = { close: vi.fn() };
  const produtoService = { listarAtivosParaSelecao: vi.fn(() => of([ativo])) };

  async function create(nota: NotaFiscal | null) {
    TestBed.resetTestingModule();
    await TestBed.configureTestingModule({
      imports: [NotaFiscalFormDialog],
      providers: [
        { provide: MAT_DIALOG_DATA, useValue: nota },
        { provide: MatDialogRef, useValue: dialogRef },
        { provide: ProdutoHttpService, useValue: produtoService },
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(NotaFiscalFormDialog);
    fixture.detectChanges();
    return fixture;
  }

  beforeEach(() => dialogRef.close.mockReset());

  it('filtra produtos por código ou descrição ignorando acentos e caixa', async () => {
    const fixture = await create(null);
    const component = fixture.componentInstance as unknown as {
      pesquisaProduto: { set(value: string): void };
      produtosFiltrados(): readonly Produto[];
    };
    component.pesquisaProduto.set('produto ativo');
    expect(component.produtosFiltrados()).toEqual([ativo]);
    component.pesquisaProduto.set('inexistente');
    expect(component.produtosFiltrados()).toEqual([]);
  });

  it('fecha somente com cliente, endereço, produto e quantidade válidos', async () => {
    const fixture = await create(null);
    const component = fixture.componentInstance as unknown as { form: any; salvar(): void };
    component.form.patchValue({ nomeCliente: 'CLIENTE', enderecoCliente: 'RUA 1' });
    component.form.controls.itens.at(0).setValue({ codigoProduto: ativo.codigo, quantidade: 2 });
    component.salvar();
    expect(dialogRef.close).toHaveBeenCalledWith({
      nomeCliente: 'CLIENTE', enderecoCliente: 'RUA 1',
      itens: [{ codigoProduto: ativo.codigo, quantidade: 2 }],
    });
  });

  it('destaca produto inativo e impede salvar até substituí-lo', async () => {
    const nota: NotaFiscal = {
      id: 'nota-1', numero: 1, status: 'ABERTA', nomeCliente: 'CLIENTE', enderecoCliente: 'RUA',
      quantidadeTotal: 1, valorTotal: 10, dataCadastro: '2026-08-24T10:00:00Z',
      dataAtualizacao: '2026-08-24T10:00:00Z',
      itens: [{ id: 'item-1', produtoId: 'inativo', codigoProduto: 'SKU-INATIVO',
        descricaoProduto: 'PRODUTO INATIVO', quantidade: 1, valor: 10, valorTotal: 10 }],
    };
    const fixture = await create(nota);
    const component = fixture.componentInstance as unknown as {
      possuiProdutosInativos(): boolean; salvar(): void;
    };

    expect(component.possuiProdutosInativos()).toBe(true);
    expect(fixture.nativeElement.textContent).toContain('Existem produtos inativos nesta nota');
    component.salvar();
    expect(dialogRef.close).not.toHaveBeenCalled();
  });
});
