import { Component, input } from '@angular/core';

@Component({
  selector: 'app-page-header',
  template: `
    <header class="page-header">
      <div class="page-header__content">
        <p class="page-header__eyebrow">{{ eyebrow() }}</p>
        <h1 class="page-header__title">{{ title() }}</h1>
        <span class="page-header__description">{{ description() }}</span>
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
