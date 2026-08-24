import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { API_CONFIG } from '../../core/config/api.config';
import { NotaFiscal } from './nota-fiscal.model';
import { NotaFiscalHttpService } from './nota-fiscal-http.service';

const nota: NotaFiscal = {
  id: '9df862bd-1f03-4ca8-bf97-971274291f6f',
  numero: 1001,
  status: 'ABERTA',
  nomeCliente: 'MARIA DA SILVA',
  enderecoCliente: 'RUA DAS FLORES, 100',
  quantidadeTotal: 2,
  valorTotal: 319.8,
  itens: [
    {
      id: '03ccfaf4-53b1-47c0-a621-6406811327b2',
      produtoId: '11fd3b6c-2de0-4ded-b388-a942247ebfea',
      codigoProduto: 'SKU-001',
      descricaoProduto: 'TECLADO MECANICO',
      quantidade: 2,
      valor: 159.9,
      valorTotal: 319.8,
    },
  ],
  dataCadastro: '2026-08-24T12:00:00Z',
  dataAtualizacao: '2026-08-24T12:00:00Z',
};

describe('NotaFiscalHttpService', () => {
  let service: NotaFiscalHttpService;
  let httpController: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        {
          provide: API_CONFIG,
          useValue: {
            estoqueUrl: 'http://localhost:8081',
            faturamentoUrl: 'http://localhost:8082',
          },
        },
      ],
    });
    service = TestBed.inject(NotaFiscalHttpService);
    httpController = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpController.verify());

  it('lista notas com paginação e filtros', () => {
    service
      .listar({ numero: 1001, status: 'ABERTA', nomeCliente: 'MARIA', pagina: 2, tamanhoPagina: 10 })
      .subscribe((response) => expect(response.itens).toEqual([nota]));

    const request = httpController.expectOne(
      (candidate) =>
        candidate.url === 'http://localhost:8082/notas-fiscais' &&
        candidate.params.get('numero') === '1001' &&
        candidate.params.get('status') === 'ABERTA' &&
        candidate.params.get('nomeCliente') === 'MARIA' &&
        candidate.params.get('pagina') === '2',
    );
    expect(request.request.method).toBe('GET');
    request.flush({
      success: true,
      data: { itens: [nota], total: 1, pagina: 2, tamanhoPagina: 10, totalPaginas: 1 },
    });
  });

  it('cria e atualiza uma nota com o corpo mínimo', () => {
    const input = {
      nomeCliente: 'MARIA DA SILVA',
      enderecoCliente: 'RUA DAS FLORES, 100',
      itens: [{ codigoProduto: 'SKU-001', quantidade: 2 }],
    };

    service.criar(input).subscribe();
    const createRequest = httpController.expectOne('http://localhost:8082/notas-fiscais');
    expect(createRequest.request.method).toBe('POST');
    expect(createRequest.request.body).toEqual(input);
    createRequest.flush({ success: true, data: nota });

    service.atualizar(nota.id, input).subscribe();
    const updateRequest = httpController.expectOne(
      `http://localhost:8082/notas-fiscais/${nota.id}`,
    );
    expect(updateRequest.request.method).toBe('PUT');
    updateRequest.flush({ success: true, data: nota });
  });

  it('busca uma nota para acompanhar o processamento', () => {
    service.buscar(nota.id).subscribe((response) => expect(response).toEqual(nota));
    const request = httpController.expectOne(`http://localhost:8082/notas-fiscais/${nota.id}`);
    expect(request.request.method).toBe('GET');
    request.flush({ success: true, data: nota });
  });

  it('inicia o fechamento assíncrono', () => {
    service.iniciarFechamento(nota.id).subscribe((response) =>
      expect(response.status).toBe('PROCESSANDO'),
    );
    const request = httpController.expectOne(
      `http://localhost:8082/notas-fiscais/${nota.id}/fechamento`,
    );
    expect(request.request.method).toBe('POST');
    request.flush({ success: true, data: { ...nota, status: 'PROCESSANDO' } });
  });
});
