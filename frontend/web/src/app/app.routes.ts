import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: '',
    pathMatch: 'full',
    redirectTo: 'inicio',
  },
  {
    path: 'inicio',
    loadComponent: () =>
      import('./features/inicio/inicio-page').then((component) => component.InicioPage),
    title: 'Visão geral',
  },
  {
    path: 'produtos',
    loadComponent: () =>
      import('./features/produtos/produtos-page').then((component) => component.ProdutosPage),
    title: 'Produtos',
  },
  {
    path: 'notas-fiscais',
    loadComponent: () =>
      import('./features/notas-fiscais/notas-fiscais-page').then(
        (component) => component.NotasFiscaisPage,
      ),
    title: 'Notas fiscais',
  },
  {
    path: '**',
    redirectTo: 'inicio',
  },
];
