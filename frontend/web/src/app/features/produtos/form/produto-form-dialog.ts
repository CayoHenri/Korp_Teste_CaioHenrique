import { Component, inject } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatIconModule } from '@angular/material/icon';
import { Produto } from '../produto.model';

@Component({
  selector: 'app-produto-form-dialog',
  imports: [
    MatButtonModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatIconModule,
    ReactiveFormsModule,
  ],
  templateUrl: './produto-form-dialog.html',
  styleUrl: './produto-form-dialog.scss',
})
export class ProdutoFormDialog {
  private readonly formBuilder = inject(FormBuilder);
  private readonly dialogRef = inject(MatDialogRef<ProdutoFormDialog>);
  readonly produto = inject<Produto | null>(MAT_DIALOG_DATA);
  readonly editando = this.produto !== null;

  readonly form = this.formBuilder.nonNullable.group({
    codigo: [{ value: this.produto?.codigo ?? '', disabled: this.editando }, Validators.required],
    descricao: [this.produto?.descricao ?? '', [Validators.required, Validators.maxLength(255)]],
    saldo: [this.produto?.saldo ?? 0, [Validators.required, Validators.min(0)]],
    valor: [this.produto?.valor ?? 0, [Validators.required, Validators.min(0.01)]],
  });

  salvar(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    this.dialogRef.close(this.form.getRawValue());
  }
}
