import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { map, Observable } from 'rxjs';
import { API_CONFIG } from '../../core/config/api.config';
import { ApiSuccessResponse } from '../../core/http/api-response.model';
import {
  NotaFiscal,
  NotaFiscalFiltros,
  NotaFiscalInput,
  NotasFiscaisPaginadas,
} from './nota-fiscal.model';

@Injectable({ providedIn: 'root' })
export class NotaFiscalHttpService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${inject(API_CONFIG).faturamentoUrl}/notas-fiscais`;

  listar(filtros: NotaFiscalFiltros): Observable<NotasFiscaisPaginadas> {
    let params = new HttpParams()
      .set('pagina', filtros.pagina)
      .set('tamanhoPagina', filtros.tamanhoPagina);

    if (filtros.numero !== null) {
      params = params.set('numero', filtros.numero);
    }
    if (filtros.status !== null) {
      params = params.set('status', filtros.status);
    }
    if (filtros.nomeCliente) {
      params = params.set('nomeCliente', filtros.nomeCliente);
    }

    return this.http
      .get<ApiSuccessResponse<NotasFiscaisPaginadas>>(this.baseUrl, { params })
      .pipe(map((response) => response.data));
  }

  buscar(id: string): Observable<NotaFiscal> {
    return this.http
      .get<ApiSuccessResponse<NotaFiscal>>(`${this.baseUrl}/${id}`)
      .pipe(map((response) => response.data));
  }

  criar(input: NotaFiscalInput): Observable<NotaFiscal> {
    return this.http
      .post<ApiSuccessResponse<NotaFiscal>>(this.baseUrl, input)
      .pipe(map((response) => response.data));
  }

  atualizar(id: string, input: NotaFiscalInput): Observable<NotaFiscal> {
    return this.http
      .put<ApiSuccessResponse<NotaFiscal>>(`${this.baseUrl}/${id}`, input)
      .pipe(map((response) => response.data));
  }

  iniciarFechamento(id: string): Observable<NotaFiscal> {
    return this.http
      .post<ApiSuccessResponse<NotaFiscal>>(`${this.baseUrl}/${id}/fechamento`, {})
      .pipe(map((response) => response.data));
  }
}
