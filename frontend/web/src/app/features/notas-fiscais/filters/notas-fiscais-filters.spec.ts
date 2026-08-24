import { TestBed } from '@angular/core/testing';
import { NotaFiscalFilterValue, NotasFiscaisFilters } from './notas-fiscais-filters';

describe('NotasFiscaisFilters', () => {
  it('emite filtros preenchidos e depois os limpa', async () => {
    await TestBed.configureTestingModule({ imports: [NotasFiscaisFilters] }).compileComponents();
    const fixture = TestBed.createComponent(NotasFiscaisFilters);
    const component = fixture.componentInstance;
    const values: NotaFiscalFilterValue[] = [];
    component.filterChange.subscribe((value) => values.push(value));
    const internal = component as unknown as { form: any; filtrar(): void; limpar(): void };

    internal.form.setValue({ numero: 10, nomeCliente: ' Cliente ', status: 'ABERTA' });
    internal.filtrar();
    internal.limpar();

    expect(values).toEqual([
      { numero: 10, nomeCliente: 'Cliente', status: 'ABERTA' },
      { numero: null, nomeCliente: '', status: null },
    ]);
  });
});
