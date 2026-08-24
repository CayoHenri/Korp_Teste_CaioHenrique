import { Injectable } from '@angular/core';
import { BehaviorSubject, distinctUntilChanged, map } from 'rxjs';
import { NotaFiscalResumo } from './nota-fiscal.model';

interface NotasFiscaisState {
  itens: readonly NotaFiscalResumo[];
  carregando: boolean;
  erro: string | null;
}

const initialState: NotasFiscaisState = {
  itens: [],
  carregando: false,
  erro: null,
};

@Injectable()
export class NotasFiscaisStore {
  private readonly stateSubject = new BehaviorSubject<NotasFiscaisState>(initialState);
  readonly state$ = this.stateSubject.asObservable();
  readonly itens$ = this.state$.pipe(
    map((state) => state.itens),
    distinctUntilChanged(),
  );
  readonly carregando$ = this.state$.pipe(
    map((state) => state.carregando),
    distinctUntilChanged(),
  );

  definirItens(itens: readonly NotaFiscalResumo[]): void {
    this.patch({ itens, erro: null });
  }

  definirCarregando(carregando: boolean): void {
    this.patch({ carregando });
  }

  definirErro(erro: string): void {
    this.patch({ erro, carregando: false });
  }

  private patch(partialState: Partial<NotasFiscaisState>): void {
    this.stateSubject.next({ ...this.stateSubject.value, ...partialState });
  }
}
