import { Injectable } from '@angular/core';
import { BehaviorSubject, distinctUntilChanged, map, Observable } from 'rxjs';

export interface AppState {
  loading: boolean;
  pageTitle: string;
}

const initialState: AppState = {
  loading: false,
  pageTitle: 'Visão geral',
};

@Injectable({ providedIn: 'root' })
export class AppStore {
  private readonly stateSubject = new BehaviorSubject<AppState>(initialState);

  readonly state$ = this.stateSubject.asObservable();
  readonly loading$ = this.select((state) => state.loading);
  readonly pageTitle$ = this.select((state) => state.pageTitle);

  setLoading(loading: boolean): void {
    this.patchState({ loading });
  }

  setPageTitle(pageTitle: string): void {
    this.patchState({ pageTitle });
  }

  private select<T>(project: (state: AppState) => T): Observable<T> {
    return this.state$.pipe(map(project), distinctUntilChanged());
  }

  private patchState(partialState: Partial<AppState>): void {
    this.stateSubject.next({ ...this.stateSubject.value, ...partialState });
  }
}
