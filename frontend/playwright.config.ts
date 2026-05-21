import { defineConfig, devices } from "@playwright/test";

const baseURL = process.env.PLAYWRIGHT_BASE_URL ?? "http://127.0.0.1:5173";
/** Dedicated port so e2e does not reuse a dev server on :8080 without /ws. */
const e2eHttp = process.env.E2E_HTTP_ADDR ?? "http://127.0.0.1:18080";
const e2eWsUrl = process.env.VITE_WS_URL ?? "ws://127.0.0.1:18080/ws";

export default defineConfig({
  testDir: "./e2e",
  timeout: 120_000,
  expect: { timeout: 60_000 },
  fullyParallel: false,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? "github" : "list",
  use: {
    ...devices["Desktop Chrome"],
    baseURL,
    trace: "on-first-retry",
  },
  webServer: [
    {
      command: "bash ../scripts/e2e-backend.sh",
      url: `${e2eHttp}/healthz`,
      timeout: 90_000,
      reuseExistingServer: !process.env.CI,
      env: {
        HTTP_ADDR: ":18080",
        DATABASE_URL:
          process.env.DATABASE_URL ??
          "postgres://herogame:herogame@localhost:5432/herogame?sslmode=disable",
        REDIS_URL: process.env.REDIS_URL ?? "redis://localhost:6379/0",
      },
    },
    {
      command: `VITE_E2E=1 VITE_WS_URL=${e2eWsUrl} npm run dev -- --host 127.0.0.1 --port 5173`,
      url: "http://127.0.0.1:5173",
      timeout: 60_000,
      // Always start e2e-mode Vite so test IDs are present.
      reuseExistingServer: false,
    },
  ],
});
