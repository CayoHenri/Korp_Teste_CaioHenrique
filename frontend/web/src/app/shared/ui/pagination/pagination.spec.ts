import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MatPaginator } from '@angular/material/paginator';
import { Pagination, PaginationChange } from './pagination';

describe('Pagination', () => {
  let fixture: ComponentFixture<Pagination>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [Pagination] }).compileComponents();
    fixture = TestBed.createComponent(Pagination);
    fixture.componentRef.setInput('total', 42);
    fixture.componentRef.setInput('pagina', 2);
    fixture.componentRef.setInput('tamanhoPagina', 10);
    fixture.detectChanges();
  });

  it('converte o índice zero-based do Material para página iniciada em um', () => {
    const changes: PaginationChange[] = [];
    fixture.componentInstance.mudou.subscribe((value) => changes.push(value));

    const paginator = fixture.debugElement.query(
      (node) => node.componentInstance instanceof MatPaginator,
    ).componentInstance as MatPaginator;
    paginator.page.emit({ pageIndex: 2, pageSize: 20, length: 42, previousPageIndex: 1 });

    expect(changes).toEqual([{ pagina: 3, tamanhoPagina: 20 }]);
  });
});
