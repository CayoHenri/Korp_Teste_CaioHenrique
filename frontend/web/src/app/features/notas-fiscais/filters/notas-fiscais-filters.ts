import { Component, inject, output } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { NotaFiscalFiltros, NotaFiscalStatus } from '../nota-fiscal.model';

export type NotaFiscalFilterValue = Pick<NotaFiscalFiltros, 'numero' | 'status' | 'nomeCliente'>;

@Component({
  selector: 'app-notas-fiscais-filters',
  imports: [
    MatButtonModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatSelectModule,
    ReactiveFormsModule,
  ],
  templateUrl: './notas-fiscais-filters.html',
  styleUrl: './notas-fiscais-filters.scss',
})
export class NotasFiscaisFilters {
  private readonly formBuilder = inject(FormBuilder);
  readonly filterChange = output<NotaFiscalFilterValue>();

  protected readonly form = this.formBuilder.group({
    numero: [null as number | null],
    nomeCliente: [''],
    status: [null as NotaFiscalStatus | null],
  });

  protected filtrar(): void {
    const value = this.form.getRawValue();
    this.filterChange.emit({
      numero: value.numero,
      nomeCliente: value.nomeCliente?.trim() ?? '',
      status: value.status,
    });
  }

  protected limpar(): void {
    this.form.reset({ numero: null, nomeCliente: '', status: null });
    this.filtrar();
  }
}
