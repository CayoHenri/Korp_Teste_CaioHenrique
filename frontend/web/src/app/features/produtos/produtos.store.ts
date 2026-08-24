import { Injectable } from '@angular/core';
import { BehaviorSubject, distinctUntilChanged, map } from 'rxjs';
import { Produto } from './produto.model';

interface ProdutosState {
  itens: readonly Produto[];
  carregando: boolean;
  erro: string | null;
}

const initialState: ProdutosState = {
  itens: [],
  carregando: false,
  erro: null,
};

@Injectable()
export class ProdutosStore {
  private readonly stateSubject = new BehaviorSubject<ProdutosState>(initialState);
  readonly state$ = this.stateSubject.asObservable();
  readonly itens$ = this.state$.pipe(
    map((state) => state.itens),
    distinctUntilChanged(),
  );
  readonly carregando$ = this.state$.pipe(
    map((state) => state.carregando),
    distinctUntilChanged(),
  );

  definirItens(itens: readonly Produto[]): void {
    this.patch({ itens, erro: null });
  }

  definirCarregando(carregando: boolean): void {
    this.patch({ carregando });
  }

  definirErro(erro: string): void {
    this.patch({ erro, carregando: false });
  }

  private patch(partialState: Partial<ProdutosState>): void {
    this.stateSubject.next({ ...this.stateSubject.value, ...partialState });
  }
}
