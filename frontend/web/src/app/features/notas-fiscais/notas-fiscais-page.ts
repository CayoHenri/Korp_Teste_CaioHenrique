import { AsyncPipe } from '@angular/common';
import { Component } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatIconModule } from '@angular/material/icon';
import { PageHeader } from '../../shared/ui/page-header/page-header';
import { NotasFiscaisStore } from './notas-fiscais.store';

@Component({
  selector: 'app-notas-fiscais-page',
  imports: [AsyncPipe, MatButtonModule, MatCardModule, MatIconModule, PageHeader],
  providers: [NotasFiscaisStore],
  templateUrl: './notas-fiscais-page.html',
  styleUrl: '../shared-feature-page.scss',
})
export class NotasFiscaisPage {
  constructor(protected readonly store: NotasFiscaisStore) {}
}
