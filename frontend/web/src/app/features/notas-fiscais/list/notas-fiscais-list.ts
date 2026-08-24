import { CurrencyPipe, DatePipe } from '@angular/common';
import { Component, input, output } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatTableModule } from '@angular/material/table';
import { MatTooltipModule } from '@angular/material/tooltip';
import { Pagination, PaginationChange } from '../../../shared/ui/pagination/pagination';
import { NotaFiscal, NotaFiscalStatus } from '../nota-fiscal.model';

@Component({
  selector: 'app-notas-fiscais-list',
  imports: [
    CurrencyPipe,
    DatePipe,
    MatButtonModule,
    MatIconModule,
    MatProgressSpinnerModule,
    MatTableModule,
    MatTooltipModule,
    Pagination,
  ],
  templateUrl: './notas-fiscais-list.html',
  styleUrl: './notas-fiscais-list.scss',
})
export class NotasFiscaisList {
  readonly notas = input.required<readonly NotaFiscal[]>();
  readonly total = input.required<number>();
  readonly pagina = input.required<number>();
  readonly tamanhoPagina = input.required<number>();
  readonly acoesEmAndamento = input<readonly string[]>([]);
  readonly editar = output<NotaFiscal>();
  readonly imprimir = output<NotaFiscal>();
  readonly paginar = output<PaginationChange>();

  protected readonly colunas = [
    'numero',
    'cliente',
    'itens',
    'quantidade',
    'valor',
    'status',
    'atualizado',
    'acoes',
  ];

  protected processando(nota: NotaFiscal): boolean {
    return nota.status === 'PROCESSANDO' || this.acoesEmAndamento().includes(nota.id);
  }

  protected statusLabel(status: NotaFiscalStatus): string {
    const labels: Record<NotaFiscalStatus, string> = {
      ABERTA: 'Aberta',
      PROCESSANDO: 'Processando',
      FECHADA: 'Fechada',
    };
    return labels[status];
  }
}
