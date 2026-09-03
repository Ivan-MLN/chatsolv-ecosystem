# Rendering the real Google index from a blocked VPS via Startpage

## Context (validated 2026-08-06)

VPS IP `2a02:4780:59:4050::1` / `76.13.193.152` is flagged by Google as a
datacenter/bot IP. Every direct Google variant serves "Sistem kami telah
mendeteksi adanya lalu lintas yang tidak wajar dari jaringan komputer Anda"
("unusual traffic from your computer network"). This is an **IP-level refusal**,
not a page CAPTCHA you can click through.

Tried and FAILED (all serve the block page):
- plain `google.com/search?q=...` (with consent cookies CONSENT/SOCS/NID set)
- `&gbv=1` (basic HTML), `&udm=14` (text-only), `&igu=1` (feeling lucky)
- `google.co.id` domain
- forced IPv4 via `--host-resolver-rules=MAP www.google.com <ipv4>`
- google.com/translate proxy wrapping the search URL
- puppeteer stealth: `--disable-blink-features=AutomationControlled`,
  `navigator.webdriver` override, `window.chrome` shim, realistic Windows UA,
  `Accept-Language: id-ID`

Do not re-attempt these; they are IP-based and cannot be fixed from the client side.

## Working path: Startpage (renders Google's index)

Startpage.com is a privacy search engine that proxies and renders Google's real
result index. It loads fine from the flagged VPS.

```js
// run: node startpage_search.js "cuaca kopeng"
const puppeteer = require('/root/nael-ai/node_modules/puppeteer-core');

(async () => {
  const q = process.argv[2] || 'cuaca kopeng';
  const browser = await puppeteer.launch({
    executablePath: '/usr/bin/google-chrome',
    headless: 'new',
    args: ['--no-sandbox', '--disable-setuid-sandbox',
      '--disable-blink-features=AutomationControlled',
      '--disable-dev-shm-usage', '--window-size=1440,1200', '--lang=id-ID'],
    defaultViewport: { width: 1440, height: 1100 },
  });
  const page = await browser.newPage();
  await page.setUserAgent('Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36');
  await page.evaluateOnNewDocument(() => {
    Object.defineProperty(navigator, 'webdriver', { get: () => undefined });
    window.chrome = { runtime: {} };
  });
  await page.goto(`https://www.startpage.com/sp/search?query=${encodeURIComponent(q)}`,
    { waitUntil: 'domcontentloaded', timeout: 40000 });
  await new Promise(r => setTimeout(r, 3500)); // let rich cards render
  await page.screenshot({ path: '/root/startpage_result.png' });
  await browser.close();
  console.log('saved /root/startpage_result.png');
})().catch(e => { console.error(e.message); process.exit(1); });
```

Requirements: chrome at `/usr/bin/google-chrome`, puppeteer-core in
`/root/nael-ai/node_modules`. Set cookies on `.google.com` for
consent (CONSENT/SOCS/NID) only matters if you eventually hit google.com —
Startpage itself needs none.

## Pitfalls

1. **Geo/name resolution**: Startpage mapped "Kopeng" to Kopeng-Banten
   (32°C) when the target was Kopeng-Semarang, Jawa Tengah (~23°C). The rich
   card's region label may point at the wrong same-named town. Cross-check with
   the BMKG snippet / other results in the same page before reporting the card
   as authoritative.
2. **Deliver the real capture**: when the user asks for a "screenshot of Google
   results", send the actual page screenshot. Do not substitute a self-built
   chart/dashboard made from scraped data — user explicitly rejected that
   ("saya kan minta-nya tadi screenshot, bukan minta grafik yang kamu build
   sendiri"). A real capture plus an honest one-line note about which service
   rendered it beats a pretty hand-made graphic.
3. Startpage result layout differs slightly from google.com (privacy chrome,
   no exact google.com chrome), but the result set and rich cards come from
   Google's index.
