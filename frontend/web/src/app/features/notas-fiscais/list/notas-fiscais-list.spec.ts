import { registerLocaleData } from '@angular/common';
import localePt from '@angular/common/locales/pt';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NotaFiscal } from '../nota-fiscal.model';
import { NotasFiscaisList } from './notas-fiscais-list';

registerLocaleData(localePt);

const nota: NotaFiscal = {
  id: 'nota-1', numero: 55, status: 'ABERTA', nomeCliente: 'CLIENTE', enderecoCliente: 'RUA 1',
  quantidadeTotal: 2, valorTotal: 100, dataCadastro: '2026-08-24T10:00:00Z',
  dataAtualizacao: '2026-08-24T11:00:00Z',
  itens: [{ id: 'item-1', produtoId: 'produto-1', codigoProduto: 'SKU-001',
    descricaoProduto: 'PRODUTO', quantidade: 2, valor: 50, valorTotal: 100 }],
};

describe('NotasFiscaisList', () => {
  let fixture: ComponentFixture<NotasFiscaisList>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [NotasFiscaisList] }).compileComponents();
    fixture = TestBed.createComponent(NotasFiscaisList);
    fixture.componentRef.setInput('notas', [nota]);
    fixture.componentRef.setInput('total', 1);
    fixture.componentRef.setInput('pagina', 1);
    fixture.componentRef.setInput('tamanhoPagina', 10);
    fixture.detectChanges();
  });

  it('habilita edição e impressão para nota aberta e emite as ações', () => {
    const actions: string[] = [];
    fixture.componentInstance.editar.subscribe(() => actions.push('editar'));
    fixture.componentInstance.imprimir.subscribe(() => actions.push('imprimir'));
    const element: HTMLElement = fixture.nativeElement;
    const editar = element.querySelector('[aria-label="Editar nota"]') as HTMLButtonElement;
    const imprimir = element.querySelector('[aria-label="Imprimir nota"]') as HTMLButtonElement;

    expect(editar.disabled).toBe(false);
    expect(imprimir.disabled).toBe(false);
    editar.click();
    imprimir.click();
    expect(actions).toEqual(['editar', 'imprimir']);
  });

  it('bloqueia as ações de nota fechada', () => {
    fixture.componentRef.setInput('notas', [{ ...nota, status: 'FECHADA' }]);
    fixture.detectChanges();
    const element: HTMLElement = fixture.nativeElement;
    expect((element.querySelector('[aria-label="Editar nota"]') as HTMLButtonElement).disabled).toBe(true);
    expect((element.querySelector('[aria-label="Imprimir nota"]') as HTMLButtonElement).disabled).toBe(true);
  });
});
