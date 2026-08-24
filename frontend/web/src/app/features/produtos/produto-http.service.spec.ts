import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { API_CONFIG } from '../../core/config/api.config';
import { Produto } from './produto.model';
import { ProdutoHttpService } from './produto-http.service';

const produto: Produto = {
  id: 'b723729d-d046-4d6e-9327-cfba7e39834a',
  codigo: 'SKU-001',
  descricao: 'TECLADO MECANICO',
  saldo: 10,
  valor: 159.9,
  ativo: true,
  dataCadastro: '2026-08-24T10:00:00Z',
  dataAtualizacao: '2026-08-24T10:00:00Z',
};

describe('ProdutoHttpService', () => {
  let service: ProdutoHttpService;
  let httpController: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        {
          provide: API_CONFIG,
          useValue: { estoqueUrl: '/api/estoque', faturamentoUrl: '/api/faturamento' },
        },
      ],
    });

    service = TestBed.inject(ProdutoHttpService);
    httpController = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpController.verify());

  it('lista produtos com filtros e extrai os dados do envelope', () => {
    service
      .listar({
        codigo: 'SKU',
        descricao: 'TECLADO',
        ativo: true,
        pagina: 2,
        tamanhoPagina: 10,
      })
      .subscribe((response) => expect(response.itens).toEqual([produto]));

    const request = httpController.expectOne(
      (candidate) =>
        candidate.url === '/api/estoque/produtos' &&
        candidate.params.get('pagina') === '2' &&
        candidate.params.get('tamanhoPagina') === '10' &&
        candidate.params.get('codigo') === 'SKU' &&
        candidate.params.get('descricao') === 'TECLADO' &&
        candidate.params.get('ativo') === 'true',
    );

    expect(request.request.method).toBe('GET');
    request.flush({
      success: true,
      data: { itens: [produto], total: 1, pagina: 2, tamanhoPagina: 10, totalPaginas: 1 },
    });
  });

  it('cadastra um produto', () => {
    const input = { codigo: 'SKU-001', descricao: 'TECLADO MECANICO', saldo: 10, valor: 159.9 };
    service.criar(input).subscribe((response) => expect(response).toEqual(produto));

    const request = httpController.expectOne('/api/estoque/produtos');
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual(input);
    request.flush({ success: true, data: produto });
  });

  it('atualiza somente os campos permitidos', () => {
    const input = { descricao: 'TECLADO RGB', saldo: 20, valor: 199.9 };
    service.atualizar(produto.id, input).subscribe();

    const request = httpController.expectOne(`/api/estoque/produtos/${produto.id}`);
    expect(request.request.method).toBe('PUT');
    expect(request.request.body).toEqual(input);
    request.flush({ success: true, data: { ...produto, ...input } });
  });

  it('escolhe a operação de status a partir do estado atual', () => {
    service.alterarStatus(produto).subscribe();

    const request = httpController.expectOne(`/api/estoque/produtos/${produto.id}/inativar`);
    expect(request.request.method).toBe('PATCH');
    request.flush({ success: true, data: { ...produto, ativo: false } });
  });
});
