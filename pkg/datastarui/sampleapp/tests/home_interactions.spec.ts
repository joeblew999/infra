import { test, expect } from '@playwright/test';

test.describe('DatastarUI sample home', () => {
  test('counter card increments and note is rendered', async ({ page }) => {
    await page.goto('/');

    const counterButton = page.getByTestId('home-counter');
    await expect(counterButton).toContainText('Clicked 0 times');

    await counterButton.click();
    await expect(counterButton).toContainText('Clicked 1 times');

    await expect(page.getByTestId('home-counter-note')).toContainText('Playwright verifies that the counter increments');
  });
});
