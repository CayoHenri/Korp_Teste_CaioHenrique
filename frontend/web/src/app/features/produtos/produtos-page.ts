import { AsyncPipe } from '@angular/common';
import { Component } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatIconModule } from '@angular/material/icon';
import { PageHeader } from '../../shared/ui/page-header/page-header';
import { ProdutosStore } from './produtos.store';

@Component({
  selector: 'app-produtos-page',
  imports: [AsyncPipe, MatButtonModule, MatCardModule, MatIconModule, PageHeader],
  providers: [ProdutosStore],
  templateUrl: './produtos-page.html',
  styleUrl: '../shared-feature-page.scss',
})
export class ProdutosPage {
  constructor(protected readonly store: ProdutosStore) {}
}
