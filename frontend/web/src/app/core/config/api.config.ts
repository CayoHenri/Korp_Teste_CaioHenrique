import { InjectionToken } from '@angular/core';
import { environment } from '../../../environments/environment';

export interface ApiConfig {
  estoqueUrl: string;
  faturamentoUrl: string;
}

export const API_CONFIG = new InjectionToken<ApiConfig>('API_CONFIG', {
  factory: () => ({
    estoqueUrl: environment.estoqueApiUrl,
    faturamentoUrl: environment.faturamentoApiUrl,
  }),
});
