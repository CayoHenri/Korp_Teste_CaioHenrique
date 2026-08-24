import { registerLocaleData } from '@angular/common';
import localePt from '@angular/common/locales/pt';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Produto } from '../produto.model';
import { ProdutosList } from './produtos-list';

registerLocaleData(localePt);

const produto: Produto = {
  id: 'produto-1', codigo: 'SKU-001', descricao: 'TECLADO', saldo: 10,
  valor: 159.9, ativo: true, dataCadastro: '2026-08-24T10:00:00Z',
  dataAtualizacao: '2026-08-24T11:00:00Z',
};

describe('ProdutosList', () => {
  let fixture: ComponentFixture<ProdutosList>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [ProdutosList] }).compileComponents();
    fixture = TestBed.createComponent(ProdutosList);
    fixture.componentRef.setInput('produtos', [produto]);
    fixture.componentRef.setInput('total', 1);
    fixture.componentRef.setInput('pagina', 1);
    fixture.componentRef.setInput('tamanhoPagina', 10);
    fixture.detectChanges();
  });

  it('renderiza os dados e emite cada ação da linha', () => {
    const actions: string[] = [];
    fixture.componentInstance.verMovimentacoes.subscribe(() => actions.push('historico'));
    fixture.componentInstance.editar.subscribe(() => actions.push('editar'));
    fixture.componentInstance.alterarStatus.subscribe(() => actions.push('status'));

    const element: HTMLElement = fixture.nativeElement;
    expect(element.textContent).toContain('SKU-001');
    expect(element.textContent).toContain('R$ 159,90');
    (element.querySelector('[aria-label="Ver movimentações"]') as HTMLButtonElement).click();
    (element.querySelector('[aria-label="Editar produto"]') as HTMLButtonElement).click();
    (element.querySelector('[aria-label="Inativar produto"]') as HTMLButtonElement).click();

    expect(actions).toEqual(['historico', 'editar', 'status']);
  });
});
