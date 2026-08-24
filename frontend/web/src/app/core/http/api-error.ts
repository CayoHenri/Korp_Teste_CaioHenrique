import { HttpErrorResponse } from '@angular/common/http';
import { ApiErrorResponse } from './api-response.model';

export function apiErrorMessage(error: unknown, serviceName?: string): string {
  if (error instanceof HttpErrorResponse) {
    const response = error.error as Partial<ApiErrorResponse> | null;
    const message = response?.error?.message;

    if (message) return message;

    if (error.status === 0) {
      return serviceName
        ? `Não foi possível conectar à API de ${serviceName}.`
        : 'Não foi possível conectar à API.';
    }
  }

  return 'Não foi possível concluir a operação.';
}
