import { Component, input, output } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';

export type DataFeedbackType = 'empty' | 'error';

@Component({
  selector: 'app-data-feedback',
  imports: [MatButtonModule, MatIconModule],
  templateUrl: './data-feedback.html',
  styleUrl: './data-feedback.scss',
})
export class DataFeedback {
  readonly type = input.required<DataFeedbackType>();
  readonly title = input.required<string>();
  readonly description = input<string>('');
  readonly retry = output<void>();

  protected get icon(): string {
    return this.type() === 'error' ? 'error_outline' : 'inbox';
  }
}
