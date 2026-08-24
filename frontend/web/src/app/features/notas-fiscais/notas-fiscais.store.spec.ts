import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, Subject, throwError } from 'rxjs';
import { NotaFiscalHttpService } from './nota-fiscal-http.service';
import { NotaFiscal } from './nota-fiscal.model';
import { NotasFiscaisState, NotasFiscaisStore } from './notas-fiscais.store';

const notaAberta: NotaFiscal = {
  id: 'nota-1',
  numero: 1,
  status: 'ABERTA',
  nomeCliente: 'CLIENTE',
  enderecoCliente: 'RUA 1',
  quantidadeTotal: 1,
  valorTotal: 50,
  itens: [{
    id: 'item-1', produtoId: 'produto-1', codigoProduto: 'SKU-001',
    descricaoProduto: 'PRODUTO', quantidade: 1, valor: 50, valorTotal: 50,
  }],
  dataCadastro: '2026-08-24T10:00:00Z',
  dataAtualizacao: '2026-08-24T10:00:00Z',
};

describe('NotasFiscaisStore', () => {
  let store: NotasFiscaisStore;
  let state: NotasFiscaisState;
  let service: {
    listar: ReturnType<typeof vi.fn>;
    criar: ReturnType<typeof vi.fn>;
    atualizar: ReturnType<typeof vi.fn>;
    iniciarFechamento: ReturnType<typeof vi.fn>;
    buscar: ReturnType<typeof vi.fn>;
  };

  beforeEach(() => {
    vi.useFakeTimers();
    service = {
      listar: vi.fn(), criar: vi.fn(), atualizar: vi.fn(),
      iniciarFechamento: vi.fn(), buscar: vi.fn(),
    };
    TestBed.configureTestingModule({
      providers: [NotasFiscaisStore, { provide: NotaFiscalHttpService, useValue: service }],
    });
    store = TestBed.inject(NotasFiscaisStore);
    store.state$.subscribe((value) => (state = value));
  });

  afterEach(() => {
    store.ngOnDestroy();
    vi.useRealTimers();
  });

  it('carrega, filtra e pagina notas', () => {
    service.listar.mockReturnValue(
      of({ itens: [notaAberta], total: 1, pagina: 1, tamanhoPagina: 20, totalPaginas: 1 }),
    );

    store.paginar(2, 20);
    expect(service.listar).toHaveBeenLastCalledWith(expect.objectContaining({ pagina: 2 }));

    store.filtrar({ numero: 1, status: 'ABERTA', nomeCliente: 'CLIENTE' });
    expect(service.listar).toHaveBeenLastCalledWith(expect.objectContaining({ pagina: 1 }));
    expect(state.itens).toEqual([notaAberta]);
    expect(state.filtros.tamanhoPagina).toBe(20);
  });

  it('expõe erro e encerra o carregamento', () => {
    service.listar.mockReturnValue(
      throwError(() => new HttpErrorResponse({ status: 0 })),
    );
    store.carregar();
    expect(state.erro).toBe('Não foi possível conectar à API de Faturamento.');
    expect(state.carregando).toBe(false);
  });

  it('acompanha o fechamento sem duplicar polling e publica o resultado', () => {
    const processando = { ...notaAberta, status: 'PROCESSANDO' as const };
    const fechada = { ...notaAberta, status: 'FECHADA' as const };
    service.iniciarFechamento.mockReturnValue(of(processando));
    service.buscar.mockReturnValue(of(fechada));
    const resultados: boolean[] = [];
    store.resultadoFechamento$.subscribe((resultado) => resultados.push(resultado.sucesso));

    store.iniciarFechamento(notaAberta).subscribe();
    expect(state.acoesEmAndamento).toEqual([]);

    vi.advanceTimersByTime(1500);

    expect(service.buscar).toHaveBeenCalledTimes(1);
    expect(service.buscar).toHaveBeenCalledWith(notaAberta.id);
    expect(resultados).toEqual([true]);
  });

  it('recarrega a lista quando o resultado deixa de atender ao filtro atual', () => {
    const processando = { ...notaAberta, status: 'PROCESSANDO' as const };
    service.listar.mockReturnValue(
      of({ itens: [], total: 0, pagina: 1, tamanhoPagina: 10, totalPaginas: 0 }),
    );
    store.filtrar({ numero: null, status: 'PROCESSANDO', nomeCliente: '' });
    service.iniciarFechamento.mockReturnValue(of(processando));
    service.buscar.mockReturnValue(of(notaAberta));

    store.iniciarFechamento(notaAberta).subscribe();
    vi.advanceTimersByTime(1500);

    expect(service.listar).toHaveBeenCalledTimes(2);
  });
});
