import { inject, Injectable, OnDestroy } from '@angular/core';
import {
  BehaviorSubject,
  catchError,
  EMPTY,
  exhaustMap,
  finalize,
  Observable,
  Subject,
  switchMap,
  takeUntil,
  takeWhile,
  tap,
  timer,
} from 'rxjs';
import { apiErrorMessage } from '../../core/http/api-error';
import { NotaFiscalHttpService } from './nota-fiscal-http.service';
import {
  NotaFiscal,
  NotaFiscalFiltros,
  NotaFiscalInput,
  ResultadoFechamento,
} from './nota-fiscal.model';

export interface NotasFiscaisState {
  itens: readonly NotaFiscal[];
  filtros: NotaFiscalFiltros;
  total: number;
  totalPaginas: number;
  carregando: boolean;
  acoesEmAndamento: readonly string[];
  erro: string | null;
}

const filtrosIniciais: NotaFiscalFiltros = {
  numero: null,
  status: null,
  nomeCliente: '',
  pagina: 1,
  tamanhoPagina: 10,
};

const initialState: NotasFiscaisState = {
  itens: [],
  filtros: filtrosIniciais,
  total: 0,
  totalPaginas: 0,
  carregando: false,
  acoesEmAndamento: [],
  erro: null,
};

@Injectable()
export class NotasFiscaisStore implements OnDestroy {
  private readonly service = inject(NotaFiscalHttpService);
  private readonly stateSubject = new BehaviorSubject<NotasFiscaisState>(initialState);
  private readonly recarregarSubject = new Subject<void>();
  private readonly resultadoFechamentoSubject = new Subject<ResultadoFechamento>();
  private readonly destroySubject = new Subject<void>();
  private readonly notasMonitoradas = new Set<string>();

  readonly state$ = this.stateSubject.asObservable();
  readonly resultadoFechamento$ = this.resultadoFechamentoSubject.asObservable();

  constructor() {
    this.recarregarSubject
      .pipe(
        tap(() => this.patch({ carregando: true, erro: null })),
        switchMap(() =>
          this.service.listar(this.stateSubject.value.filtros).pipe(
            catchError((error: unknown) => {
              this.patch({ erro: apiErrorMessage(error, 'Faturamento') });
              return EMPTY;
            }),
            finalize(() => this.patch({ carregando: false })),
          ),
        ),
        takeUntil(this.destroySubject),
      )
      .subscribe((pagina) => {
        this.patch({
          itens: pagina.itens,
          total: pagina.total,
          totalPaginas: pagina.totalPaginas,
          filtros: {
            ...this.stateSubject.value.filtros,
            pagina: pagina.pagina,
            tamanhoPagina: pagina.tamanhoPagina,
          },
        });

        pagina.itens
          .filter((nota) => nota.status === 'PROCESSANDO')
          .forEach((nota) => this.acompanharFechamento(nota.id));
      });
  }

  carregar(): void {
    this.recarregarSubject.next();
  }

  filtrar(filtros: Pick<NotaFiscalFiltros, 'numero' | 'status' | 'nomeCliente'>): void {
    this.patch({ filtros: { ...this.stateSubject.value.filtros, ...filtros, pagina: 1 } });
    this.carregar();
  }

  paginar(pagina: number, tamanhoPagina: number): void {
    this.patch({ filtros: { ...this.stateSubject.value.filtros, pagina, tamanhoPagina } });
    this.carregar();
  }

  criar(input: NotaFiscalInput): Observable<NotaFiscal> {
    return this.executarMutacao(this.service.criar(input));
  }

  atualizar(id: string, input: NotaFiscalInput): Observable<NotaFiscal> {
    return this.executarMutacao(this.service.atualizar(id, input));
  }

  iniciarFechamento(nota: NotaFiscal): Observable<NotaFiscal> {
    this.adicionarAcao(nota.id);

    return this.service.iniciarFechamento(nota.id).pipe(
      tap((atualizada) => {
        this.atualizarItem(atualizada);
        this.acompanharFechamento(atualizada.id);
      }),
      finalize(() => this.removerAcao(nota.id)),
    );
  }

  ngOnDestroy(): void {
    this.destroySubject.next();
    this.destroySubject.complete();
    this.resultadoFechamentoSubject.complete();
    this.stateSubject.complete();
  }

  private executarMutacao(operacao: Observable<NotaFiscal>): Observable<NotaFiscal> {
    return operacao.pipe(tap(() => this.carregar()));
  }

  private acompanharFechamento(id: string): void {
    if (this.notasMonitoradas.has(id)) {
      return;
    }

    this.notasMonitoradas.add(id);
    timer(1500, 1500)
      .pipe(
        exhaustMap(() => this.service.buscar(id).pipe(catchError(() => EMPTY))),
        tap((nota) => {
          this.atualizarItem(nota);
          if (nota.status !== 'PROCESSANDO') {
            this.resultadoFechamentoSubject.next({
              nota,
              sucesso: nota.status === 'FECHADA',
            });
            if (
              this.stateSubject.value.filtros.status !== null &&
              this.stateSubject.value.filtros.status !== nota.status
            ) {
              this.carregar();
            }
          }
        }),
        takeWhile((nota) => nota.status === 'PROCESSANDO', true),
        finalize(() => this.notasMonitoradas.delete(id)),
        takeUntil(this.destroySubject),
      )
      .subscribe();
  }

  private atualizarItem(nota: NotaFiscal): void {
    this.patch({
      itens: this.stateSubject.value.itens.map((item) => (item.id === nota.id ? nota : item)),
    });
  }

  private adicionarAcao(id: string): void {
    this.patch({
      acoesEmAndamento: [...this.stateSubject.value.acoesEmAndamento, id],
    });
  }

  private removerAcao(id: string): void {
    this.patch({
      acoesEmAndamento: this.stateSubject.value.acoesEmAndamento.filter((item) => item !== id),
    });
  }

  private patch(partialState: Partial<NotasFiscaisState>): void {
    this.stateSubject.next({ ...this.stateSubject.value, ...partialState });
  }
}
