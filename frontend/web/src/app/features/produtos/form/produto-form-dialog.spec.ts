import { TestBed } from '@angular/core/testing';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { Produto } from '../produto.model';
import { ProdutoFormDialog } from './produto-form-dialog';

describe('ProdutoFormDialog', () => {
  const dialogRef = { close: vi.fn() };

  async function create(produto: Produto | null) {
    TestBed.resetTestingModule();
    await TestBed.configureTestingModule({
      imports: [ProdutoFormDialog],
      providers: [
        { provide: MAT_DIALOG_DATA, useValue: produto },
        { provide: MatDialogRef, useValue: dialogRef },
      ],
    }).compileComponents();
    return TestBed.createComponent(ProdutoFormDialog);
  }

  beforeEach(() => dialogRef.close.mockReset());

  it('não fecha com formulário inválido', async () => {
    const fixture = await create(null);
    fixture.componentInstance.salvar();
    expect(dialogRef.close).not.toHaveBeenCalled();
    expect(fixture.componentInstance.form.touched).toBe(true);
  });

  it('preserva o código desabilitado e fecha com os dados da edição', async () => {
    const produto: Produto = {
      id: 'produto-1', codigo: 'SKU-001', descricao: 'TECLADO', saldo: 10, valor: 100,
      ativo: true, dataCadastro: '2026-08-24T10:00:00Z', dataAtualizacao: '2026-08-24T10:00:00Z',
    };
    const fixture = await create(produto);
    fixture.componentInstance.form.patchValue({ descricao: 'TECLADO RGB', saldo: 20, valor: 150 });
    fixture.componentInstance.salvar();
    expect(fixture.componentInstance.form.controls.codigo.disabled).toBe(true);
    expect(dialogRef.close).toHaveBeenCalledWith({
      codigo: 'SKU-001', descricao: 'TECLADO RGB', saldo: 20, valor: 150,
    });
  });
});
