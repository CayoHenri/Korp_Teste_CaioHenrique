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
import { DataFeedback } from '../../shared/ui/data-feedback/data-feedback';
import { PaginationChange } from '../../shared/ui/pagination/pagination';
import { PageHeader } from '../../shared/ui/page-header/page-header';
import { ProdutoFilterValue, ProdutosFilters } from './filters/produtos-filters';
import { ProdutoFormDialog } from './form/produto-form-dialog';
import { ProdutosList } from './list/produtos-list';
import { AtualizarProdutoInput, CriarProdutoInput, Produto } from './produto.model';
import { ProdutosStore } from './produtos.store';

type ProdutoFormValue = CriarProdutoInput;

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
      })
      .afterClosed()
      .pipe(
        filter((value): value is ProdutoFormValue => value !== undefined),
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
        error: () => undefined,
      });
  }

  protected alterarStatus(produto: Produto): void {
    const acao = produto.ativo ? 'inativar' : 'ativar';
    if (!window.confirm(`Deseja ${acao} o produto ${produto.codigo}?`)) {
      return;
    }

    this.store
      .alterarStatus(produto)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: () => this.mostrarSucesso(`Produto ${acao === 'ativar' ? 'ativado' : 'inativado'}.`),
        error: () => undefined,
      });
  }

  private mostrarSucesso(message: string): void {
    this.snackBar.open(message, 'Fechar', { duration: 3500 });
  }
}
