import { inject, Injectable, OnDestroy } from '@angular/core';
import {
  BehaviorSubject,
  catchError,
  EMPTY,
  finalize,
  map,
  Observable,
  Subject,
  switchMap,
  takeUntil,
  tap,
} from 'rxjs';
import { apiErrorMessage } from '../../core/http/api-error';
import { AtualizarProdutoInput, CriarProdutoInput, Produto, ProdutoFiltros } from './produto.model';
import { ProdutoHttpService } from './produto-http.service';

export interface ProdutosState {
  itens: readonly Produto[];
  filtros: ProdutoFiltros;
  total: number;
  totalPaginas: number;
  carregando: boolean;
  salvando: boolean;
  erro: string | null;
}

const filtrosIniciais: ProdutoFiltros = {
  codigo: '',
  descricao: '',
  ativo: null,
  pagina: 1,
  tamanhoPagina: 10,
};

const initialState: ProdutosState = {
  itens: [],
  filtros: filtrosIniciais,
  total: 0,
  totalPaginas: 0,
  carregando: false,
  salvando: false,
  erro: null,
};

@Injectable()
export class ProdutosStore implements OnDestroy {
  private readonly service = inject(ProdutoHttpService);
  private readonly stateSubject = new BehaviorSubject<ProdutosState>(initialState);
  private readonly recarregarSubject = new Subject<void>();
  private readonly destroySubject = new Subject<void>();

  readonly state$ = this.stateSubject.asObservable();
  readonly carregando$ = this.state$.pipe(map((state) => state.carregando));
  readonly salvando$ = this.state$.pipe(map((state) => state.salvando));

  constructor() {
    this.recarregarSubject
      .pipe(
        tap(() => this.patch({ carregando: true, erro: null })),
        switchMap(() =>
          this.service.listar(this.stateSubject.value.filtros).pipe(
            catchError((error: unknown) => {
              this.patch({ erro: apiErrorMessage(error, 'Estoque') });
              return EMPTY;
            }),
            finalize(() => this.patch({ carregando: false })),
          ),
        ),
        takeUntil(this.destroySubject),
      )
      .subscribe((pagina) =>
        this.patch({
          itens: pagina.itens,
          total: pagina.total,
          totalPaginas: pagina.totalPaginas,
          filtros: {
            ...this.stateSubject.value.filtros,
            pagina: pagina.pagina,
            tamanhoPagina: pagina.tamanhoPagina,
          },
        }),
      );
  }

  carregar(): void {
    this.recarregarSubject.next();
  }

  filtrar(filtros: Pick<ProdutoFiltros, 'codigo' | 'descricao' | 'ativo'>): void {
    this.patch({ filtros: { ...this.stateSubject.value.filtros, ...filtros, pagina: 1 } });
    this.carregar();
  }

  paginar(pagina: number, tamanhoPagina: number): void {
    this.patch({ filtros: { ...this.stateSubject.value.filtros, pagina, tamanhoPagina } });
    this.carregar();
  }

  criar(input: CriarProdutoInput): Observable<Produto> {
    return this.executarMutacao(this.service.criar(input));
  }

  atualizar(id: string, input: AtualizarProdutoInput): Observable<Produto> {
    return this.executarMutacao(this.service.atualizar(id, input));
  }

  alterarStatus(produto: Produto): Observable<Produto> {
    return this.executarMutacao(this.service.alterarStatus(produto));
  }

  ngOnDestroy(): void {
    this.destroySubject.next();
    this.destroySubject.complete();
    this.stateSubject.complete();
  }

  private executarMutacao(operacao: Observable<Produto>): Observable<Produto> {
    this.patch({ salvando: true });

    return operacao.pipe(
      tap(() => this.carregar()),
      finalize(() => this.patch({ salvando: false })),
    );
  }

  private patch(partialState: Partial<ProdutosState>): void {
    this.stateSubject.next({ ...this.stateSubject.value, ...partialState });
  }
}
