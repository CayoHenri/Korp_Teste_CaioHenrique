import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, Subject, throwError } from 'rxjs';
import { ProdutoHttpService } from './produto-http.service';
import { Produto } from './produto.model';
import { ProdutosState, ProdutosStore } from './produtos.store';

const produto: Produto = {
  id: 'produto-1',
  codigo: 'SKU-001',
  descricao: 'TECLADO',
  saldo: 10,
  valor: 100,
  ativo: true,
  dataCadastro: '2026-08-24T10:00:00Z',
  dataAtualizacao: '2026-08-24T10:00:00Z',
};

describe('ProdutosStore', () => {
  let store: ProdutosStore;
  let state: ProdutosState;
  let service: {
    listar: ReturnType<typeof vi.fn>;
    criar: ReturnType<typeof vi.fn>;
    atualizar: ReturnType<typeof vi.fn>;
    alterarStatus: ReturnType<typeof vi.fn>;
  };

  beforeEach(() => {
    service = {
      listar: vi.fn(),
      criar: vi.fn(),
      atualizar: vi.fn(),
      alterarStatus: vi.fn(),
    };
    TestBed.configureTestingModule({
      providers: [ProdutosStore, { provide: ProdutoHttpService, useValue: service }],
    });
    store = TestBed.inject(ProdutosStore);
    store.state$.subscribe((value) => (state = value));
  });

  afterEach(() => store.ngOnDestroy());

  it('carrega a página e atualiza os metadados', () => {
    const response = new Subject<any>();
    service.listar.mockReturnValue(response);

    store.carregar();
    expect(state.carregando).toBe(true);
    expect(service.listar).toHaveBeenCalledWith(state.filtros);

    response.next({ itens: [produto], total: 1, pagina: 2, tamanhoPagina: 20, totalPaginas: 1 });
    response.complete();

    expect(state.itens).toEqual([produto]);
    expect(state.filtros.pagina).toBe(2);
    expect(state.filtros.tamanhoPagina).toBe(20);
    expect(state.carregando).toBe(false);
  });

  it('reinicia a página ao filtrar e preserva o erro da API', () => {
    service.listar.mockReturnValue(throwError(() => new HttpErrorResponse({ status: 0 })));

    store.paginar(3, 20);
    store.filtrar({ codigo: ' SKU ', descricao: 'TEC', ativo: true });

    expect(state.filtros).toMatchObject({ pagina: 1, tamanhoPagina: 20, ativo: true });
    expect(state.erro).toBe('Não foi possível conectar à API de Estoque.');
    expect(state.carregando).toBe(false);
  });

  it('marca salvamento, executa a mutação e recarrega a lista', () => {
    const mutation = new Subject<Produto>();
    service.criar.mockReturnValue(mutation);
    service.listar.mockReturnValue(
      of({ itens: [produto], total: 1, pagina: 1, tamanhoPagina: 10, totalPaginas: 1 }),
    );

    store
      .criar({ codigo: produto.codigo, descricao: produto.descricao, saldo: 10, valor: 100 })
      .subscribe();
    expect(state.salvando).toBe(true);

    mutation.next(produto);
    mutation.complete();

    expect(service.listar).toHaveBeenCalledTimes(1);
    expect(state.itens).toEqual([produto]);
    expect(state.salvando).toBe(false);
  });
});
