import { test, expect } from '@playwright/test';

test.describe('DatastarUI sample counter', () => {
  test('increments the click counter', async ({ page }) => {
    await page.goto('/');
    const button = page.getByRole('button', { name: /clicked/i });
    await expect(button).toContainText('0');
    await button.click();
    await expect(button).toContainText('1');
    await button.click();
    await expect(button).toContainText('2');
  });
});
