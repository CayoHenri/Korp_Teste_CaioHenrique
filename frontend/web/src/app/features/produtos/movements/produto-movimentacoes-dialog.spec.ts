import { TestBed } from '@angular/core/testing';
import { MAT_DIALOG_DATA } from '@angular/material/dialog';
import { of, throwError } from 'rxjs';
import { ProdutoHttpService } from '../produto-http.service';
import { Produto } from '../produto.model';
import { ProdutoMovimentacoesDialog } from './produto-movimentacoes-dialog';

const produto: Produto = {
  id: 'produto-1', codigo: 'SKU-001', descricao: 'TECLADO', saldo: 8, valor: 50,
  ativo: true, dataCadastro: '2026-08-24T10:00:00Z', dataAtualizacao: '2026-08-24T10:00:00Z',
};

describe('ProdutoMovimentacoesDialog', () => {
  async function create(listarMovimentacoes: ReturnType<typeof vi.fn>) {
    TestBed.resetTestingModule();
    await TestBed.configureTestingModule({
      imports: [ProdutoMovimentacoesDialog],
      providers: [
        { provide: MAT_DIALOG_DATA, useValue: produto },
        { provide: ProdutoHttpService, useValue: { listarMovimentacoes } },
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(ProdutoMovimentacoesDialog);
    fixture.detectChanges();
    return fixture;
  }

  it('renderiza histórico e saldo atual', async () => {
    const fixture = await create(vi.fn(() => of([{ id: 'mov-1', produtoId: produto.id,
      tipo: 'ENTRADA', quantidade: 3, dataMovimentacao: '2026-08-24T12:00:00Z' }])));
    fixture.detectChanges();
    expect(fixture.nativeElement.textContent).toContain('Saldo atual');
    expect(fixture.nativeElement.textContent).toContain('Entrada');
    expect(fixture.nativeElement.textContent).toContain('+3');
  });

  it('apresenta feedback quando a consulta falha', async () => {
    const fixture = await create(vi.fn(() => throwError(() => new Error('falha'))));
    fixture.detectChanges();
    expect(fixture.nativeElement.textContent).toContain('Não foi possível carregar as movimentações');
    expect(fixture.nativeElement.querySelector('[role="alert"]')).not.toBeNull();
  });
});
