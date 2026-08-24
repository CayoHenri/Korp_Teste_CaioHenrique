import { AsyncPipe, DatePipe } from '@angular/common';
import { Component, inject } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatTableModule } from '@angular/material/table';
import { MatTooltipModule } from '@angular/material/tooltip';
import { catchError, map, of, startWith, Subject, switchMap } from 'rxjs';
import { apiErrorMessage } from '../../../core/http/api-error';
import { DataFeedback } from '../../../shared/ui/data-feedback/data-feedback';
import { Produto, ProdutoMovimentacao } from '../produto.model';
import { ProdutoHttpService } from '../produto-http.service';

interface MovimentacoesState {
  carregando: boolean;
  itens: readonly ProdutoMovimentacao[];
  erro: string | null;
}

@Component({
  selector: 'app-produto-movimentacoes-dialog',
  imports: [
    AsyncPipe,
    DataFeedback,
    DatePipe,
    MatButtonModule,
    MatDialogModule,
    MatIconModule,
    MatProgressBarModule,
    MatTableModule,
    MatTooltipModule,
  ],
  templateUrl: './produto-movimentacoes-dialog.html',
  styleUrl: './produto-movimentacoes-dialog.scss',
})
export class ProdutoMovimentacoesDialog {
  private readonly service = inject(ProdutoHttpService);
  private readonly recarregarSubject = new Subject<void>();

  protected readonly produto = inject<Produto>(MAT_DIALOG_DATA);
  protected readonly colunas = ['data', 'tipo', 'quantidade', 'referencia'];
  protected readonly state$ = this.recarregarSubject.pipe(
    startWith(undefined),
    switchMap(() =>
      this.service.listarMovimentacoes(this.produto.id).pipe(
        map(
          (itens): MovimentacoesState => ({
            carregando: false,
            itens,
            erro: null,
          }),
        ),
        catchError((error: unknown) =>
          of<MovimentacoesState>({
            carregando: false,
            itens: [],
            erro: apiErrorMessage(error, 'Estoque'),
          }),
        ),
        startWith<MovimentacoesState>({ carregando: true, itens: [], erro: null }),
      ),
    ),
  );

  protected recarregar(): void {
    this.recarregarSubject.next();
  }
}
