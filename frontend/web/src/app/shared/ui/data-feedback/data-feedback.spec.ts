import { ComponentFixture, TestBed } from '@angular/core/testing';
import { DataFeedback } from './data-feedback';

describe('DataFeedback', () => {
  let fixture: ComponentFixture<DataFeedback>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [DataFeedback] }).compileComponents();
    fixture = TestBed.createComponent(DataFeedback);
  });

  it('apresenta erro e emite nova tentativa', () => {
    fixture.componentRef.setInput('type', 'error');
    fixture.componentRef.setInput('title', 'Falha ao carregar');
    fixture.componentRef.setInput('description', 'API indisponível');
    let retries = 0;
    fixture.componentInstance.retry.subscribe(() => retries++);
    fixture.detectChanges();

    const element: HTMLElement = fixture.nativeElement;
    expect(element.querySelector('[role="alert"]')?.textContent).toContain('API indisponível');
    (element.querySelector('button') as HTMLButtonElement).click();
    expect(retries).toBe(1);
  });

  it('não apresenta botão no estado vazio', () => {
    fixture.componentRef.setInput('type', 'empty');
    fixture.componentRef.setInput('title', 'Sem registros');
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('button')).toBeNull();
  });
});
