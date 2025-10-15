import { test, expect } from '@playwright/test';

test.describe('DatastarUI dashboard', () => {
  test('dismisses alert and completes support ticket', async ({ page }) => {
    await page.goto('/dashboard');

    const alert = page.getByTestId('alert-container');
    await expect(alert).toBeVisible();
    await page.getByTestId('alert-dismiss').click();
    await expect(alert).toHaveAttribute('style', 'display: none;');

    await page.getByTestId('support-subject').fill('Streaming glitch');
    await page.getByTestId('support-description').fill('Video feed pauses every 30 seconds.');

    await page.getByTestId('support-submit').click();
    const status = page.getByTestId('support-status');
    await expect(status).toBeVisible();
    await expect(status).toContainText('Streaming glitch');

    await page.getByTestId('support-reset').click();
    await expect(status).toHaveAttribute('style', 'display: none;');
  });
});
