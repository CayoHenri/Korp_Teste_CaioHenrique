import { PortuguesePaginatorIntl } from './paginator-intl';

describe('PortuguesePaginatorIntl', () => {
  const paginator = new PortuguesePaginatorIntl();

  it('traduz os controles de navegação', () => {
    expect(paginator.itemsPerPageLabel).toBe('Itens por página:');
    expect(paginator.nextPageLabel).toBe('Próxima página');
    expect(paginator.previousPageLabel).toBe('Página anterior');
    expect(paginator.firstPageLabel).toBe('Primeira página');
    expect(paginator.lastPageLabel).toBe('Última página');
  });

  it('formata o intervalo da página em português', () => {
    expect(paginator.getRangeLabel(0, 5, 61)).toBe('1 – 5 de 61');
    expect(paginator.getRangeLabel(12, 5, 61)).toBe('61 – 61 de 61');
    expect(paginator.getRangeLabel(0, 5, 0)).toBe('0 de 0');
  });
});
