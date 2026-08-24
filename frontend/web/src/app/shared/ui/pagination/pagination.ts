import { Component, input, output } from '@angular/core';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';

export interface PaginationChange {
  pagina: number;
  tamanhoPagina: number;
}

@Component({
  selector: 'app-pagination',
  imports: [MatPaginatorModule],
  template: `
    <mat-paginator
      [length]="total()"
      [pageIndex]="pagina() - 1"
      [pageSize]="tamanhoPagina()"
      [pageSizeOptions]="opcoesTamanho()"
      [showFirstLastButtons]="true"
      (page)="onPage($event)"
    />
  `,
  styles: `
    :host {
      display: block;
      border-top: 1px solid var(--app-border);
      background: var(--app-surface-muted);
      --mat-paginator-container-size: 40px;
    }

    mat-paginator {
      background: transparent;
      color: var(--app-muted);
      font-size: 0.75rem;
    }
  `,
})
export class Pagination {
  readonly total = input.required<number>();
  readonly pagina = input.required<number>();
  readonly tamanhoPagina = input.required<number>();
  readonly opcoesTamanho = input<number[]>([5, 10, 20, 50]);
  readonly mudou = output<PaginationChange>();

  protected onPage(event: PageEvent): void {
    this.mudou.emit({
      pagina: event.pageIndex + 1,
      tamanhoPagina: event.pageSize,
    });
  }
}
