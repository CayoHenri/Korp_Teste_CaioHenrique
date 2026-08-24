import { AsyncPipe } from '@angular/common';
import { Component, DestroyRef, inject, OnInit } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { filter, switchMap } from 'rxjs';
import { apiErrorMessage } from '../../core/http/api-error';
import {
  ConfirmationDialog,
  ConfirmationDialogData,
} from '../../shared/ui/confirmation-dialog/confirmation-dialog';
import { DataFeedback } from '../../shared/ui/data-feedback/data-feedback';
import { PageHeader } from '../../shared/ui/page-header/page-header';
import { PaginationChange } from '../../shared/ui/pagination/pagination';
import { NotaFiscalFilterValue, NotasFiscaisFilters } from './filters/notas-fiscais-filters';
import { NotaFiscalFormDialog } from './form/nota-fiscal-form-dialog';
import { NotasFiscaisList } from './list/notas-fiscais-list';
import { NotaFiscal, NotaFiscalInput } from './nota-fiscal.model';
import { NotasFiscaisStore } from './notas-fiscais.store';

function isNotaFiscalInput(value: unknown): value is NotaFiscalInput {
  if (typeof value !== 'object' || value === null) {
    return false;
  }

  const candidate = value as Partial<NotaFiscalInput>;
  return (
    typeof candidate.nomeCliente === 'string' &&
    typeof candidate.enderecoCliente === 'string' &&
    Array.isArray(candidate.itens) &&
    candidate.itens.length > 0
  );
}

@Component({
  selector: 'app-notas-fiscais-page',
  imports: [
    AsyncPipe,
    DataFeedback,
    MatButtonModule,
    MatCardModule,
    MatDialogModule,
    MatIconModule,
    MatProgressBarModule,
    MatSnackBarModule,
    NotasFiscaisFilters,
    NotasFiscaisList,
    PageHeader,
  ],
  providers: [NotasFiscaisStore],
  templateUrl: './notas-fiscais-page.html',
  styleUrl: './notas-fiscais-page.scss',
})
export class NotasFiscaisPage implements OnInit {
  private readonly dialog = inject(MatDialog);
  private readonly snackBar = inject(MatSnackBar);
  private readonly destroyRef = inject(DestroyRef);

  protected readonly store = inject(NotasFiscaisStore);

  ngOnInit(): void {
    this.store.carregar();
    this.store.resultadoFechamento$
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe((resultado) => {
        if (resultado.sucesso) {
          this.mostrarSucesso(`Nota #${resultado.nota.numero} fechada com sucesso.`);
          return;
        }

        this.snackBar.open(
          `Nota #${resultado.nota.numero} rejeitada: ${resultado.nota.motivoRejeicao || 'motivo não informado'}`,
          'Fechar',
          { duration: 7000, panelClass: ['error-snackbar'] },
        );
      });
  }

  protected filtrar(filtros: NotaFiscalFilterValue): void {
    this.store.filtrar(filtros);
  }

  protected paginar(event: PaginationChange): void {
    this.store.paginar(event.pagina, event.tamanhoPagina);
  }

  protected abrirFormulario(nota: NotaFiscal | null = null): void {
    if (nota && nota.status !== 'ABERTA') {
      return;
    }

    this.dialog
      .open<NotaFiscalFormDialog, NotaFiscal | null, NotaFiscalInput>(NotaFiscalFormDialog, {
        data: nota,
        disableClose: true,
        width: '48rem',
        maxWidth: 'calc(100vw - 2rem)',
        autoFocus: false,
      })
      .afterClosed()
      .pipe(
        filter(isNotaFiscalInput),
        switchMap((input) =>
          nota ? this.store.atualizar(nota.id, input) : this.store.criar(input),
        ),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe({
        next: () => this.mostrarSucesso(nota ? 'Nota fiscal atualizada.' : 'Nota fiscal criada.'),
        error: (error: unknown) => this.mostrarErro(error),
      });
  }

  protected imprimir(nota: NotaFiscal): void {
    if (nota.status !== 'ABERTA') {
      return;
    }

    this.dialog
      .open<ConfirmationDialog, ConfirmationDialogData, boolean>(ConfirmationDialog, {
        data: {
          title: `Imprimir nota #${nota.numero}`,
          message:
            'A impressão iniciará o fechamento assíncrono e a nota não poderá mais ser editada durante o processamento.',
          confirmLabel: 'Iniciar impressão',
          icon: 'print',
        },
        autoFocus: false,
      })
      .afterClosed()
      .pipe(
        filter((confirmed) => confirmed === true),
        switchMap(() => this.store.iniciarFechamento(nota)),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe({
        next: () =>
          this.snackBar.open('Processamento da nota iniciado.', 'Fechar', { duration: 3000 }),
        error: (error: unknown) => this.mostrarErro(error),
      });
  }

  private mostrarSucesso(message: string): void {
    this.snackBar.open(message, 'Fechar', { duration: 3500 });
  }

  private mostrarErro(error: unknown): void {
    this.snackBar.open(apiErrorMessage(error, 'Faturamento'), 'Fechar', {
      duration: 5000,
      panelClass: ['error-snackbar'],
    });
  }
}
