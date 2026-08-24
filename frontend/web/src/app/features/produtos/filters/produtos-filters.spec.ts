import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ProdutoFilterValue, ProdutosFilters } from './produtos-filters';

describe('ProdutosFilters', () => {
  let fixture: ComponentFixture<ProdutosFilters>;
  let component: ProdutosFilters;

  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [ProdutosFilters] }).compileComponents();
    fixture = TestBed.createComponent(ProdutosFilters);
    component = fixture.componentInstance;
  });

  it('normaliza espaços e emite os filtros', () => {
    const values: ProdutoFilterValue[] = [];
    component.filterChange.subscribe((value) => values.push(value));
    const internal = component as unknown as { form: any; filtrar(): void };
    internal.form.setValue({ codigo: ' SKU ', descricao: ' Teclado ', ativo: true });
    internal.filtrar();
    expect(values).toEqual([{ codigo: 'SKU', descricao: 'Teclado', ativo: true }]);
  });

  it('limpa todos os controles e dispara nova consulta', () => {
    const values: ProdutoFilterValue[] = [];
    component.filterChange.subscribe((value) => values.push(value));
    const internal = component as unknown as { form: any; limpar(): void };
    internal.form.setValue({ codigo: 'SKU', descricao: 'Teclado', ativo: false });
    internal.limpar();
    expect(values.at(-1)).toEqual({ codigo: '', descricao: '', ativo: null });
  });
});
