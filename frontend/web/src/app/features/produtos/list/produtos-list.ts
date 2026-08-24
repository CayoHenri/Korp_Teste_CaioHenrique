import { CurrencyPipe, DatePipe } from '@angular/common';
import { Component, input, output } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTableModule } from '@angular/material/table';
import { MatTooltipModule } from '@angular/material/tooltip';
import {
  Pagination,
  PaginationChange,
} from '../../../shared/ui/pagination/pagination';
import { Produto } from '../produto.model';

@Component({
  selector: 'app-produtos-list',
  imports: [
    CurrencyPipe,
    DatePipe,
    MatButtonModule,
    MatIconModule,
    MatTableModule,
    MatTooltipModule,
    Pagination,
  ],
  templateUrl: './produtos-list.html',
  styleUrl: './produtos-list.scss',
})
export class ProdutosList {
  readonly produtos = input.required<readonly Produto[]>();
  readonly total = input.required<number>();
  readonly pagina = input.required<number>();
  readonly tamanhoPagina = input.required<number>();
  readonly editar = output<Produto>();
  readonly alterarStatus = output<Produto>();
  readonly paginar = output<PaginationChange>();

  protected readonly colunas = [
    'codigo',
    'descricao',
    'saldo',
    'valor',
    'status',
    'atualizado',
    'acoes',
  ];
}
