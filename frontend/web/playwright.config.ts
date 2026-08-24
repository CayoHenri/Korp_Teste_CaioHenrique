import { defineConfig, devices } from '@playwright/test';

const baseURL = process.env['E2E_WEB_URL'] ?? 'http://localhost:4200';

export default defineConfig({
  testDir: './e2e',
  outputDir: './test-results/e2e',
  timeout: 60_000,
  expect: {
    timeout: 15_000,
  },
  fullyParallel: false,
  workers: 1,
  retries: process.env['CI'] ? 2 : 0,
  reporter: process.env['CI']
    ? [['line'], ['html', { open: 'never', outputFolder: 'playwright-report' }]]
    : [['list'], ['html', { open: 'never', outputFolder: 'playwright-report' }]],
  use: {
    baseURL,
    locale: 'pt-BR',
    timezoneId: 'America/Sao_Paulo',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  webServer: {
    command: 'npm start -- --host localhost',
    url: baseURL,
    reuseExistingServer: !process.env['CI'],
    timeout: 120_000,
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
});
