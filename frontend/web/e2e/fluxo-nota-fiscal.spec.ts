import { expect, Locator, Page, test } from '@playwright/test';

function unique(prefix: string): string {
  return `${prefix}-${Date.now()}-${Math.floor(Math.random() * 10_000)}`;
}

async function cadastrarProduto(
  page: Page,
  codigo: string,
  saldo = 10,
  valor = '25.50',
): Promise<void> {
  await page.goto('/produtos');
  await expect(page.getByRole('heading', { name: 'Produtos' })).toBeVisible();
  await page.getByRole('button', { name: 'Novo produto' }).click();

  const dialog = page.getByRole('dialog');
  await expect(dialog.getByRole('heading', { name: 'Novo produto' })).toBeVisible();
  await dialog.getByLabel('Código').fill(codigo);
  await dialog.getByLabel('Descrição').fill(`PRODUTO E2E ${codigo}`);
  await dialog.getByLabel('Saldo').fill(String(saldo));
  await dialog.getByLabel('Valor unitário').fill(valor);
  await dialog.getByRole('button', { name: 'Salvar', exact: true }).click();

  await expect(page.getByText('Produto cadastrado.')).toBeVisible();
  await filtrarProduto(page, codigo);
  await expect(linhaPorTexto(page, codigo)).toBeVisible();
}

async function filtrarProduto(page: Page, codigo: string): Promise<void> {
  const filtros = page.locator('app-produtos-filters');
  await filtros.getByLabel('Código').fill(codigo);
  await filtros.getByRole('button', { name: 'Filtrar' }).click();
}

async function criarNota(page: Page, cliente: string, codigoProduto: string): Promise<Locator> {
  await page.goto('/notas-fiscais');
  await expect(page.getByRole('heading', { name: 'Notas fiscais' })).toBeVisible();
  await page.getByRole('button', { name: 'Nova nota' }).click();

  const dialog = page.getByRole('dialog');
  await expect(dialog.getByRole('heading', { name: 'Nova nota fiscal' })).toBeVisible();
  await dialog.getByLabel('Nome do cliente').fill(cliente);
  await dialog.getByLabel('Endereço do cliente').fill('RUA DOS TESTES, 100');
  await dialog.getByRole('combobox', { name: 'Produto' }).click();
  await page.getByRole('searchbox', { name: 'Pesquisar produtos' }).fill(codigoProduto);
  await page.getByRole('option', { name: new RegExp(codigoProduto) }).click();
  await dialog.getByLabel('Quantidade').fill('2');
  await dialog.getByRole('button', { name: 'Salvar nota' }).click();

  await expect(page.getByText('Nota fiscal criada.')).toBeVisible();
  return filtrarNotaPorCliente(page, cliente);
}

async function filtrarNotaPorCliente(page: Page, cliente: string): Promise<Locator> {
  const filtros = page.locator('app-notas-fiscais-filters');
  await filtros.getByLabel('Cliente').fill(cliente);
  await filtros.getByRole('button', { name: 'Filtrar' }).click();
  const row = linhaPorTexto(page, cliente);
  await expect(row).toBeVisible();
  return row;
}

function linhaPorTexto(page: Page, value: string): Locator {
  return page.getByRole('row').filter({ hasText: value }).first();
}

test.describe.serial('Fluxo visual de notas fiscais', () => {
  test('cadastra produto, cria nota e acompanha o fechamento assíncrono', async ({ page }) => {
    const codigo = unique('E2E-FECHAMENTO');
    const cliente = unique('CLIENTE-FECHAMENTO');

    await cadastrarProduto(page, codigo, 10);
    const row = await criarNota(page, cliente, codigo);
    await expect(row.getByText('Aberta', { exact: true })).toBeVisible();
    await row.getByRole('button', { name: 'Imprimir nota' }).click();

    const confirmation = page.getByRole('dialog');
    await expect(
      confirmation.getByText('A impressão iniciará o fechamento assíncrono'),
    ).toBeVisible();
    await confirmation.getByRole('button', { name: 'Iniciar impressão' }).click();

    await expect(page.getByText('Processamento da nota iniciado.')).toBeVisible();
    await expect(row.getByText('Fechada', { exact: true })).toBeVisible({ timeout: 30_000 });
    await expect(row.getByRole('button', { name: 'Imprimir nota' })).toBeDisabled();
  });

  test('destaca produto inativo ao editar uma nota aberta', async ({ page }) => {
    const codigo = unique('E2E-INATIVO');
    const cliente = unique('CLIENTE-INATIVO');

    await cadastrarProduto(page, codigo, 10);
    await criarNota(page, cliente, codigo);

    await page.goto('/produtos');
    await filtrarProduto(page, codigo);
    const productRow = linhaPorTexto(page, codigo);
    await productRow.getByRole('button', { name: 'Inativar produto' }).click();
    await page.getByRole('dialog').getByRole('button', { name: 'Inativar', exact: true }).click();
    await expect(page.getByText('Produto inativado.')).toBeVisible();

    await page.goto('/notas-fiscais');
    const invoiceRow = await filtrarNotaPorCliente(page, cliente);
    await invoiceRow.getByRole('button', { name: 'Editar nota' }).click();

    const dialog = page.getByRole('dialog');
    await expect(dialog.getByText('Existem produtos inativos nesta nota')).toBeVisible();
    await expect(dialog.getByText('Inativo', { exact: true })).toBeVisible();
    await expect(dialog.getByRole('button', { name: 'Salvar nota' })).toBeDisabled();
  });
});
