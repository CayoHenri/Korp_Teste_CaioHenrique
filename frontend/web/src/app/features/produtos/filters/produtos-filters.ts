import { Component, inject, output } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { ProdutoFiltros } from '../produto.model';

export type ProdutoFilterValue = Pick<ProdutoFiltros, 'codigo' | 'descricao' | 'ativo'>;

@Component({
  selector: 'app-produtos-filters',
  imports: [
    MatButtonModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatSelectModule,
    ReactiveFormsModule,
  ],
  templateUrl: './produtos-filters.html',
  styleUrl: './produtos-filters.scss',
})
export class ProdutosFilters {
  private readonly formBuilder = inject(FormBuilder);
  readonly filterChange = output<ProdutoFilterValue>();

  protected readonly form = this.formBuilder.group({
    codigo: [''],
    descricao: [''],
    ativo: [null as boolean | null],
  });

  protected filtrar(): void {
    const value = this.form.getRawValue();
    this.filterChange.emit({
      codigo: value.codigo?.trim() ?? '',
      descricao: value.descricao?.trim() ?? '',
      ativo: value.ativo,
    });
  }

  protected limpar(): void {
    this.form.reset({ codigo: '', descricao: '', ativo: null });
    this.filtrar();
  }
}
