import { CurrencyPipe } from '@angular/common';
import { Component, computed, DestroyRef, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSelectModule } from '@angular/material/select';
import { apiErrorMessage } from '../../../core/http/api-error';
import { Produto } from '../../produtos/produto.model';
import { ProdutoHttpService } from '../../produtos/produto-http.service';
import { NotaFiscal, NotaFiscalItemInput } from '../nota-fiscal.model';

@Component({
  selector: 'app-nota-fiscal-form-dialog',
  imports: [
    CurrencyPipe,
    MatButtonModule,
    MatDialogModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatProgressSpinnerModule,
    MatSelectModule,
    ReactiveFormsModule,
  ],
  templateUrl: './nota-fiscal-form-dialog.html',
  styleUrl: './nota-fiscal-form-dialog.scss',
})
export class NotaFiscalFormDialog {
  private readonly formBuilder = inject(FormBuilder);
  private readonly dialogRef = inject(MatDialogRef<NotaFiscalFormDialog>);
  private readonly produtoHttpService = inject(ProdutoHttpService);
  private readonly destroyRef = inject(DestroyRef);

  protected readonly nota = inject<NotaFiscal | null>(MAT_DIALOG_DATA);
  protected readonly editando = this.nota !== null;
  protected readonly produtosAtivos = signal<readonly Produto[]>([]);
  protected readonly carregandoProdutos = signal(true);
  protected readonly erroProdutos = signal<string | null>(null);
  protected readonly pesquisaProduto = signal('');
  protected readonly produtosFiltrados = computed(() => {
    const termo = this.normalizarPesquisa(this.pesquisaProduto());
    if (!termo) return this.produtosAtivos();

    return this.produtosAtivos().filter((produto) =>
      this.normalizarPesquisa(`${produto.codigo} ${produto.descricao}`).includes(termo),
    );
  });
  protected readonly form = this.formBuilder.nonNullable.group({
    nomeCliente: [this.nota?.nomeCliente ?? '', Validators.required],
    enderecoCliente: [this.nota?.enderecoCliente ?? '', Validators.required],
    itens: this.formBuilder.array(
      this.nota?.itens.length
        ? this.nota.itens.map((item) =>
            this.criarItemForm({
              codigoProduto: item.codigoProduto,
              quantidade: item.quantidade,
            }),
          )
        : [this.criarItemForm()],
    ),
  });

  constructor() {
    this.carregarProdutosAtivos();
  }

  protected adicionarItem(): void {
    this.form.controls.itens.push(this.criarItemForm());
  }

  protected removerItem(index: number): void {
    if (this.form.controls.itens.length > 1) {
      this.form.controls.itens.removeAt(index);
    }
  }

  protected salvar(): void {
    if (this.form.invalid || this.possuiProdutosInativos()) {
      this.form.markAllAsTouched();
      return;
    }

    this.dialogRef.close(this.form.getRawValue());
  }

  protected carregarProdutosAtivos(): void {
    this.carregandoProdutos.set(true);
    this.erroProdutos.set(null);

    this.produtoHttpService
      .listarAtivosParaSelecao()
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (produtos) => {
          this.produtosAtivos.set(produtos);
          this.carregandoProdutos.set(false);
        },
        error: (error: unknown) => {
          this.erroProdutos.set(apiErrorMessage(error, 'Estoque'));
          this.carregandoProdutos.set(false);
        },
      });
  }

  protected aoAbrirProdutos(aberto: boolean): void {
    if (aberto) {
      this.pesquisaProduto.set('');
    }
  }

  protected produtoSelecionado(codigo: string): Produto | undefined {
    return this.produtosAtivos().find((produto) => produto.codigo === codigo);
  }

  protected itemInativo(codigo: string) {
    if (!codigo || this.carregandoProdutos() || this.produtoSelecionado(codigo)) {
      return undefined;
    }

    return this.nota?.itens.find((item) => item.codigoProduto === codigo);
  }

  protected possuiProdutosInativos(): boolean {
    if (this.carregandoProdutos() || this.erroProdutos()) return false;

    return this.form.controls.itens.controls.some((item) =>
      Boolean(this.itemInativo(item.controls.codigoProduto.value)),
    );
  }

  private criarItemForm(item?: NotaFiscalItemInput) {
    return this.formBuilder.nonNullable.group({
      codigoProduto: [item?.codigoProduto ?? '', Validators.required],
      quantidade: [item?.quantidade ?? 1, [Validators.required, Validators.min(1)]],
    });
  }

  private normalizarPesquisa(value: string): string {
    return value
      .normalize('NFD')
      .replace(/[\u0300-\u036f]/g, '')
      .trim()
      .toUpperCase();
  }
}
