import { Component, inject } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { NotaFiscal, NotaFiscalItemInput } from '../nota-fiscal.model';

@Component({
  selector: 'app-nota-fiscal-form-dialog',
  imports: [
    MatButtonModule,
    MatDialogModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    ReactiveFormsModule,
  ],
  templateUrl: './nota-fiscal-form-dialog.html',
  styleUrl: './nota-fiscal-form-dialog.scss',
})
export class NotaFiscalFormDialog {
  private readonly formBuilder = inject(FormBuilder);
  private readonly dialogRef = inject(MatDialogRef<NotaFiscalFormDialog>);

  protected readonly nota = inject<NotaFiscal | null>(MAT_DIALOG_DATA);
  protected readonly editando = this.nota !== null;
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

  protected adicionarItem(): void {
    this.form.controls.itens.push(this.criarItemForm());
  }

  protected removerItem(index: number): void {
    if (this.form.controls.itens.length > 1) {
      this.form.controls.itens.removeAt(index);
    }
  }

  protected salvar(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    this.dialogRef.close(this.form.getRawValue());
  }

  private criarItemForm(item?: NotaFiscalItemInput) {
    return this.formBuilder.nonNullable.group({
      codigoProduto: [item?.codigoProduto ?? '', Validators.required],
      quantidade: [item?.quantidade ?? 1, [Validators.required, Validators.min(1)]],
    });
  }
}
