import { TestBed } from '@angular/core/testing';
import { MAT_DIALOG_DATA } from '@angular/material/dialog';
import { ConfirmationDialog, ConfirmationDialogData } from './confirmation-dialog';

describe('ConfirmationDialog', () => {
  it('renderiza textos personalizados e valores booleanos das ações', async () => {
    const data: ConfirmationDialogData = {
      title: 'Inativar produto',
      message: 'Deseja continuar?',
      confirmLabel: 'Inativar',
      cancelLabel: 'Voltar',
      icon: 'toggle_off',
    };
    await TestBed.configureTestingModule({
      imports: [ConfirmationDialog],
      providers: [{ provide: MAT_DIALOG_DATA, useValue: data }],
    }).compileComponents();

    const fixture = TestBed.createComponent(ConfirmationDialog);
    fixture.detectChanges();
    const buttons = [...fixture.nativeElement.querySelectorAll('button')] as HTMLButtonElement[];

    expect(fixture.nativeElement.textContent).toContain('Inativar produto');
    expect(buttons.map((button) => button.textContent?.trim())).toEqual(['Voltar', 'Inativar']);
    expect(buttons[0].getAttribute('ng-reflect-dialog-result')).not.toBe('true');
  });
});
