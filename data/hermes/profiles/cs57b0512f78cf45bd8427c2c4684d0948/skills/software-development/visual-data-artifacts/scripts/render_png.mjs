#!/usr/bin/env node
// Render an HTML file to a FULL-PAGE PNG via puppeteer-core. Never clips tall
// content (replaces `chrome --screenshot`, which is fixed to --window-size and
// cut off charts -> user: "kepotong / offside / jelek banget").
//
// Usage:
//   node render_png.mjs /tmp/chart.html /tmp/chart.png           # wide default
//   node render_png.mjs /tmp/chart.html /tmp/chart.png 1600       # custom width
//
// Must run with cwd=/root/nael-ai (or node_modules resolved) OR set
// PUPPETEER_CORE path. No package install needed — puppeteer-core ships in
// nael-ai's node_modules on this box.
import fs from "node:fs";

const [, , inHtml, outPng, widthArg] = process.argv;
if (!inHtml || !outPng) {
  console.error("usage: node render_png.mjs <in.html> <out.png> [width]");
  process.exit(1);
}
const width = parseInt(widthArg || "1560", 10); // >=1500 => balanced landscape

const PUPPETEER_CORE =
  process.env.PUPPETEER_CORE ||
  "/root/nael-ai/node_modules/puppeteer-core/lib/puppeteer/puppeteer-core.js";

const puppeteer = await import(PUPPETEER_CORE);
const browser = await puppeteer.launch({
  executablePath: "/usr/bin/google-chrome",
  headless: true,
  args: ["--no-sandbox", "--disable-gpu", "--hide-scrollbars", "--force-device-scale-factor=2"],
});
try {
  const page = await browser.newPage();
  await page.setViewport({ width, height: 900, deviceScaleFactor: 2 });
  await page.goto(`file://${inHtml}`, { waitUntil: "networkidle0", timeout: 30_000 });
  await page.screenshot({ path: outPng, fullPage: true, type: "png" });
  console.log(`OK: ${fs.statSync(outPng).size} bytes`);
} finally {
  await browser.close().catch(() => undefined);
}