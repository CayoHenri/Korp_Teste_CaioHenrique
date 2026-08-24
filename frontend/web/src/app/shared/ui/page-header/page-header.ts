import { Component, input } from '@angular/core';

@Component({
  selector: 'app-page-header',
  template: `
    <header class="page-header">
      <div>
        <p>{{ eyebrow() }}</p>
        <h1>{{ title() }}</h1>
        <span>{{ description() }}</span>
      </div>
      <ng-content />
    </header>
  `,
  styleUrl: './page-header.scss',
})
export class PageHeader {
  readonly eyebrow = input.required<string>();
  readonly title = input.required<string>();
  readonly description = input.required<string>();
}
