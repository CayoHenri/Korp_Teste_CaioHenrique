import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { map, Observable } from 'rxjs';
import { API_CONFIG } from '../../core/config/api.config';
import { ApiSuccessResponse } from '../../core/http/api-response.model';
import {
  AtualizarProdutoInput,
  CriarProdutoInput,
  Produto,
  ProdutoFiltros,
  ProdutoMovimentacao,
  ProdutosPaginados,
} from './produto.model';

@Injectable({ providedIn: 'root' })
export class ProdutoHttpService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${inject(API_CONFIG).estoqueUrl}/produtos`;

  listar(filtros: ProdutoFiltros): Observable<ProdutosPaginados> {
    let params = new HttpParams()
      .set('pagina', filtros.pagina)
      .set('tamanhoPagina', filtros.tamanhoPagina);

    if (filtros.codigo) {
      params = params.set('codigo', filtros.codigo);
    }
    if (filtros.descricao) {
      params = params.set('descricao', filtros.descricao);
    }
    if (filtros.ativo !== null) {
      params = params.set('ativo', filtros.ativo);
    }

    return this.http
      .get<ApiSuccessResponse<ProdutosPaginados>>(this.baseUrl, { params })
      .pipe(map((response) => response.data));
  }

  criar(input: CriarProdutoInput): Observable<Produto> {
    return this.http
      .post<ApiSuccessResponse<Produto>>(this.baseUrl, input)
      .pipe(map((response) => response.data));
  }

  atualizar(id: string, input: AtualizarProdutoInput): Observable<Produto> {
    return this.http
      .put<ApiSuccessResponse<Produto>>(`${this.baseUrl}/${id}`, input)
      .pipe(map((response) => response.data));
  }

  alterarStatus(produto: Produto): Observable<Produto> {
    const acao = produto.ativo ? 'inativar' : 'ativar';
    return this.http
      .patch<ApiSuccessResponse<Produto>>(`${this.baseUrl}/${produto.id}/${acao}`, {})
      .pipe(map((response) => response.data));
  }

  listarMovimentacoes(produtoId: string): Observable<ProdutoMovimentacao[]> {
    return this.http
      .get<ApiSuccessResponse<ProdutoMovimentacao[]>>(
        `${this.baseUrl}/${produtoId}/movimentacoes`,
      )
      .pipe(map((response) => response.data));
  }
}
