import { Component } from '@angular/core';
import { AppLayout } from './layout/app-layout';

@Component({
  imports: [AppLayout],
  selector: 'app-root',
  styleUrl: './app.scss',
  templateUrl: './app.html',
})
export class App {}
