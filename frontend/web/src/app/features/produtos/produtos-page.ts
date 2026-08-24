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
import { DataFeedback } from '../../shared/ui/data-feedback/data-feedback';
import {
  ConfirmationDialog,
  ConfirmationDialogData,
} from '../../shared/ui/confirmation-dialog/confirmation-dialog';
import { PaginationChange } from '../../shared/ui/pagination/pagination';
import { PageHeader } from '../../shared/ui/page-header/page-header';
import { ProdutoFilterValue, ProdutosFilters } from './filters/produtos-filters';
import { ProdutoFormDialog } from './form/produto-form-dialog';
import { ProdutosList } from './list/produtos-list';
import { ProdutoMovimentacoesDialog } from './movements/produto-movimentacoes-dialog';
import { AtualizarProdutoInput, CriarProdutoInput, Produto } from './produto.model';
import { ProdutosStore } from './produtos.store';

type ProdutoFormValue = CriarProdutoInput;

function isProdutoFormValue(value: unknown): value is ProdutoFormValue {
  if (typeof value !== 'object' || value === null) {
    return false;
  }

  const candidate = value as Partial<ProdutoFormValue>;
  return (
    typeof candidate.codigo === 'string' &&
    typeof candidate.descricao === 'string' &&
    typeof candidate.saldo === 'number' &&
    typeof candidate.valor === 'number'
  );
}

@Component({
  selector: 'app-produtos-page',
  imports: [
    AsyncPipe,
    DataFeedback,
    MatButtonModule,
    MatCardModule,
    MatDialogModule,
    MatIconModule,
    MatProgressBarModule,
    MatSnackBarModule,
    PageHeader,
    ProdutosFilters,
    ProdutosList,
  ],
  providers: [ProdutosStore],
  templateUrl: './produtos-page.html',
  styleUrl: './produtos-page.scss',
})
export class ProdutosPage implements OnInit {
  private readonly dialog = inject(MatDialog);
  private readonly snackBar = inject(MatSnackBar);
  private readonly destroyRef = inject(DestroyRef);

  protected readonly store = inject(ProdutosStore);

  ngOnInit(): void {
    this.store.carregar();
  }

  protected filtrar(filtros: ProdutoFilterValue): void {
    this.store.filtrar(filtros);
  }

  protected paginar(event: PaginationChange): void {
    this.store.paginar(event.pagina, event.tamanhoPagina);
  }

  protected abrirFormulario(produto: Produto | null = null): void {
    this.dialog
      .open<ProdutoFormDialog, Produto | null, ProdutoFormValue>(ProdutoFormDialog, {
        data: produto,
        disableClose: true,
        width: '36rem',
        maxWidth: 'calc(100vw - 2rem)',
        autoFocus: false,
      })
      .afterClosed()
      .pipe(
        filter(isProdutoFormValue),
        switchMap((value) => {
          if (!produto) {
            return this.store.criar(value);
          }

          const atualizacao: AtualizarProdutoInput = {
            descricao: value.descricao,
            saldo: value.saldo,
            valor: value.valor,
          };
          return this.store.atualizar(produto.id, atualizacao);
        }),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe({
        next: () => this.mostrarSucesso(produto ? 'Produto atualizado.' : 'Produto cadastrado.'),
        error: (error: unknown) => this.mostrarErro(error),
      });
  }

  protected abrirMovimentacoes(produto: Produto): void {
    this.dialog.open<ProdutoMovimentacoesDialog, Produto>(ProdutoMovimentacoesDialog, {
      data: produto,
      width: '48rem',
      maxWidth: 'calc(100vw - 2rem)',
      autoFocus: false,
    });
  }

  protected alterarStatus(produto: Produto): void {
    const acao = produto.ativo ? 'inativar' : 'ativar';

    this.dialog
      .open<ConfirmationDialog, ConfirmationDialogData, boolean>(ConfirmationDialog, {
        data: {
          title: `${produto.ativo ? 'Inativar' : 'Ativar'} produto`,
          message: `Deseja ${acao} o produto ${produto.codigo}?`,
          confirmLabel: produto.ativo ? 'Inativar' : 'Ativar',
          icon: produto.ativo ? 'toggle_off' : 'toggle_on',
        },
        autoFocus: false,
      })
      .afterClosed()
      .pipe(
        filter((confirmed) => confirmed === true),
        switchMap(() => this.store.alterarStatus(produto)),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe({
        next: () => this.mostrarSucesso(`Produto ${acao === 'ativar' ? 'ativado' : 'inativado'}.`),
        error: (error: unknown) => this.mostrarErro(error),
      });
  }

  private mostrarSucesso(message: string): void {
    this.snackBar.open(message, 'Fechar', { duration: 3500 });
  }

  private mostrarErro(error: unknown): void {
    this.snackBar.open(apiErrorMessage(error), 'Fechar', {
      duration: 5000,
      panelClass: ['error-snackbar'],
    });
  }
}
