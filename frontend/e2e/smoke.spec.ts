import { test, expect } from "@playwright/test";

/**
 * LEAD-001 / BACKLOG-006 scripted loop: connect → buy → move → combat modal.
 * Uses E2E DOM controls (VITE_E2E=1) instead of Konva canvas clicks.
 */
test("connect, recruit, march to bandit camp, combat modal", async ({ page }) => {
  await page.goto("/?playerId=1");

  await expect(page.getByTestId("connection-status")).toContainText("connected", {
    timeout: 20_000,
  });

  const recruitPlus1 = page.getByTestId("recruit-1-plus1");
  await expect(recruitPlus1).toBeVisible({ timeout: 5_000 });

  // Wait for spawn grace to expire; movement to creep nodes is locked during grace.
  await page.waitForTimeout(13_000);

  // Starting gold 200 → four Pikemen (50g each).
  for (let i = 0; i < 4; i++) {
    await recruitPlus1.click();
    await page.waitForTimeout(400);
  }

  const moveCrossroads = page.getByTestId("e2e-move-2");
  const moveBandit = page.getByTestId("e2e-move-5");

  await moveCrossroads.click();
  await expect(moveBandit).toBeEnabled({ timeout: 45_000 });

  await moveBandit.click();
  await expect(page.getByTestId("combat-modal")).toBeVisible({
    timeout: 90_000,
  });
  // CombatModal renders "Battle — Victory" / "Battle — Defeat" (HoMM
  // theme reskin) instead of the literal outcome words. Match either
  // of those rather than the legacy /win|loss/i.
  await expect(page.getByTestId("combat-modal")).toContainText(
    /victory|defeat/i,
  );
});
