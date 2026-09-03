---
name: web-search-scraping
description: Engines bot-blocking? Prefer Bing RSS curl, then scrapling.
version: 1.0.0
metadata:
  hermes:
    tags: [web-search, scraping, bing-rss, anti-bot, vps, scrapling, research]
    related_skills: [scrapling, domain-intel, grounded-citations]
---

# Web Search & Scraping from a Blocked VPS

When your VPS IP is flagged as a datacenter/bot (very common on Google, DuckDuckGo, Bing, Ecosia, Brave — they serve CAPTCHA/"verify you are human"/Cloudflare challenges), do NOT fight the CAPTCHA. Use the fast paths below. The browser stack will usually just sit on a challenge page.

## Multi-Source Research Workflow (Mandatory Quality Rule)
When answering queries requiring factual research or current information, do NOT rely on a single search engine or unverified assumptions. Use multi-source queries across available search tools and providers (Serper, Linkup, Firecrawl, Jina Reader from 9router DB `/root/.9router/db/data.sqlite`) to cross-verify facts across platforms before generating responses.

## 1. 9router Web Search & Fetch APIs (Serper, Linkup, Firecrawl, Jina Reader)
Extract active API keys directly from 9router's SQLite database (`/root/.9router/db/data.sqlite` table `providerConnections`):

```python
import sqlite3, json, requests

conn = sqlite3.connect('/root/.9router/db/data.sqlite')
def get_key(provider_name):
    row = conn.execute("SELECT data FROM providerConnections WHERE provider = ? AND isActive = 1", (provider_name,)).fetchone()
    return json.loads(row[0]).get('apiKey') if row else None

# Serper Search:
serper_key = get_key('serper')
res = requests.post('https://google.serper.dev/search', headers={'X-API-KEY': serper_key}, json={'q': 'query string', 'gl': 'id', 'hl': 'id'}).json()

# Linkup Search:
linkup_key = get_key('linkup')
res = requests.post('https://api.linkup.so/v1/search', headers={'Authorization': f'Bearer {linkup_key}'}, json={'q': 'query string', 'depth': 'standard', 'outputType': 'searchResults'}).json()

# Firecrawl Web Fetch / Scrape:
fc_key = get_key('firecrawl')
res = requests.post('https://api.firecrawl.dev/v1/scrape', headers={'Authorization': f'Bearer {fc_key}'}, json={'url': 'target_url'}).json()

# Jina Reader Web Fetch (Markdown conversion):
jina_key = get_key('jina-reader')
res = requests.get('https://r.jina.ai/' + 'target_url', headers={'Authorization': f'Bearer {jina_key}'}).text
```

## 1b. Bing RSS — the reliable search fallback
The single most dependable search path from a blocked IP. No browser, no CAPTCHA, returns parseable RSS.

```
curl -s --max-time 30 -A "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/124.0 Safari/537.36" \
  "https://www.bing.com/search?format=rss&q=<urlencoded query>" -o rss.xml
```
Parse `<item><title>`, `<link>`, `<description>` with regex/ElementTree. Adds ~0.5s per query. Do several queries if one set is thin.

Pitfalls:
- **Bing sometimes caches/returns identical result sets for different queries** — if two very different queries return the exact same items, refresh/re-execute or reword the query; don't trust stale results as per-query truth.
- **Generic domain roots on tech queries** — broad hardware search queries (e.g. "harga AMD EPYC" or "NVIDIA RTX 6000 Blackwell") in Bing RSS often return generic root domains (nvidia.com, wikipedia, tokopik) instead of specific price snippets. Reword with targeted parameters like specific model numbers ("EPYC 9654", "B200"), "MSRP price list", or "launch price".
- `format=rss` also bot-blocks SOMETIMES, but far less than the HTML/browser page; retry once before giving up.

## 1c. Local SearXNG Meta-Search (Port 8888)
A local SearXNG service runs directly on the VPS (`http://localhost:8888`). It aggregates multi-engine results without IP blocking, returning structured JSON:

```python
import urllib.request, urllib.parse, json

query = 'Nahida Genshin Impact cosplay'
url = f'http://localhost:8888/search?q={urllib.parse.quote(query)}&format=json&categories=images'
req = urllib.request.Request(url, headers={'User-Agent': 'Mozilla/5.0'})
res = json.loads(urllib.request.urlopen(req, timeout=5).read().decode('utf-8'))
results = res.get('results', [])
# Extract img_src / thumbnail_src / url
```
- Supported categories: `images`, `general`, `news`, `science`, etc.
- Bypasses datacenter IP blocks completely by routing locally. Always test port 8888 when external search APIs fail or return empty/CAPTCHA results.

## 2. Direct site probing with `scrapling` (HTTP Fetcher)
Pure-python `Fetcher` handles static pages with no browser after `pip install scrapling`:

```python
from scrapling.fetchers import Fetcher
import re
page = Fetcher.get(url, timeout=25)                            # 0.2.x: NO `impersonate=` kwarg
txt = re.sub(r"\s+", " ", page.css_first("body").get().get())  # Adaptor -> string via .get().get()
```
- Some sites are fully JS-rendered (body tiny/near-empty via HTTP) — those need a real browser, not Fetcher.
- Sites whose body says "Enable JavaScript" / Cloudflare challenge — route to the normal browser tools.
- For pages with inline structured data (e.g. university PMB "tgl_mulai"/"tgl_selesai"/"nama_jalur"), parse the JSON blob directly rather than reading prose.

## 2b. Google from a datacenter IP → render the real index via Startpage
Google.com hard-blocks flagged datacenter IPs at the **network level** (serves "unusual traffic" / "lalu lintas yang tidak wajar"). This is NOT captcha-on-a-page — no UA, consent cookie, consent-flow, `gbv=1`/`udm=14`/`igu=1`, `google.co.id`, forced-IPv4 `--host-resolver-rules`, or puppeteer stealth fixes it; the IP itself is refused. Don't burn cycles "defeating" it — go straight to a Google-index renderer.

**Startpage** (`startpage.com`) proxies and renders the genuine Google result index from a non-flagged IP. It works from a blocked VPS via puppeteer:
- Navigate `https://www.startpage.com/sp/search?query=<urlencoded>`
- Real Google weather/rich cards + web result snippets appear (rendered from Google's index, not hand-built).
- Do NOT hand-build a chart/dashboard to stand in for a requested screenshot — the user explicitly rejects that. If they said "screenshot the Google result", send the actual page capture.
- Important: Startpage may geo-resolve the query to the *wrong* place with the same name (e.g. it mapped "Kopeng" to Kopeng-Banten 32°C when the target was Kopeng-Semarang 23°C). Cross-check the result location against BMKG/other snippets before trusting the summary card; say so to the user.

Reference: `references/google-render-via-startpage.md` has the full puppeteer bypass script.

## 3. Cloudflare Turnstile ("Just a moment…" / "Tunggu sebentar…") → camoufox via scrapling StealthyFetcher
When a site sits behind Cloudflare Turnstile, real Chrome (headless OR headful via xvfb) with stealth flags will usually FAIL from a flagged DC IP — the Turnstile is non-interactive (no checkbox to click) and evaluates the browser fingerprint; Chrome sits on the challenge forever (60s+ of polling, `Ray ID` in footer). The reliable path that worked:

- Install once: `pip install "scrapling[all]"` then `scrapling install` (downloads the **camoufox** browser + GeoIP db + UBO addon to `~/.cache/camoufox`). Without it, StealthyFetcher raises `FileNotFoundError: Version information not found at ~/.cache/camoufox/version.json`.
- **scrapling 0.2.x has NO `solve_cloudflare` kwarg** — that's newer-version syntax (the hub skill docs show it, but installed 0.2.99 rejects it). The working substitute: plain `StealthyFetcher.async_fetch(..., headless=True, humanize=True, block_webrtc=True)` — camoufox itself clears the challenge in ~3-10s.
- Detect challenge pass by polling `page.title()` + body text for the *localized* keywords: `/tunggu sebentar|just a moment|verifikasi keamanan|security verification|checking your browser/i` (title is "Tunggu sebentar..." in Indonesian, not "Just a moment..." — an English-only regex falsely reports success).
- Screenshot the page from INSIDE the `page_action` callback (the returned `Response` object has no `save_screenshot`; and `page_action` MUST `return page` or the CWD page becomes None).
- Reference: `references/cloudflare-turnstile-camoufox.md` has the working async script.

## 3b. Fingerprint-based bot detection (NO Turnstile) — encrypted API payloads
Some sites (Next.js/SPA apps) don't use Cloudflare at all — they **fingerprint the browser** server-side (probe endpoints like `api/v1/fingerprint/sync`, `api/v1/geo/resolve`, `api/v1/telemetry/report`, `api/v1/config/fetch`) and *silently degrade* flagged browsers instead of showing a challenge. Signature: every API returns HTTP 200 but the body is **one encrypted field** `{"e":"<long base64>"}` in headless Chrome, while the **same flow in camoufox returns plain readable JSON** (`{"success":true,...}`). Client-side code then surfaces a dismissive error ("Security verification failed", "Session invalid", "Please try again") because it can't decrypt the payload. Key rules:
- Encrypted payload in Chrome + plaintext JSON in camoufox on the same URL/flow = **fingerprint gate**, not a bad token/URL. Diff the two network logs to prove it (pattern below) instead of guessing.
- Switch the WHOLE flow to camoufox (`StealthyFetcher`/`AsyncCamoufox`) immediately — don't burn cycles on Chrome stealth flags (UA, WebDriver, disable-blink); a real fingerprint sync beats those.
- If verification still hangs at "Establishing secure session…/Verifying…" even in camoufox with plaintext JSON, the wall is gone and the **token itself is invalid/expired** — stop chasing bypasses; the fix is a valid credential from the account owner.
- Full diagnostic + network-logging snippet: `references/fingerprint-bot-detection.md`.

## 3c. Short drama aggregators & dynamic edge stream sites (e.g. Narto-Drama)
Sites like `narto-drama.com` aggregate short dramas from multiple providers (DramaBox, BiliTV, Shortical, etc.) and resist static scraping (`requests`/`bs4`) completely:
- **No media URLs in static HTML:** Streams are generated dynamically via JavaScript through Cloudflare Worker edge fleets (e.g., `dramabox-edge.*.workers.dev`, `edge.narto-drama.com`).
- **Signed short-lived tokens & headers:** Video URLs require `XSRF-TOKEN`/session cookies and `X-Requested-With: XMLHttpRequest` header checks.
- **Client gating & popunder wrappers:** Local JavaScript uses `localStorage` gate checks and injected `NavGuard` popunder wrappers (Adsterra) that break basic click loops.
- **Scrape strategy:** Must use Playwright/Puppeteer with stealth (or Camoufox) and intercept XHR/Fetch network traffic to capture stream tokens during player init. Full architectural breakdown: `references/short-drama-aggregator-anti-scraping.md`.

## Pitfall: "Connection refused" (ERR_CONNECTION_REFUSED / Errno 111)
A foreign/DC VPS often cannot even OPEN the TCP connection to Indonesian domains (datacenter IP blocked at their edge, some also behind Cloudflare). This is NOT a content problem — skip those sites and tell the user to verify manually. Mark them as "website not reachable from here" rather than "closed".

## Reference detail
See `references/bing-rss-and-live-site-scrape.md` for worked parseable snippets, a live university-admissions example with `tgl_selesai` extraction, and the approach for "find unis with open admissions" style research.

## Recording interactive flows (Playwright video → WhatsApp)
When the user asks to "record and send a video" of a browser interaction (click button → popup → fill token → submit), or the site is a Next.js SPA (curl returns 308 + empty body, content renders client-side): use `record_video_dir` + `record_video_size` on the Playwright context, read `page.video.path()` only after `ctx.close()`, and ship the `.webm` via Baileys `wa_send_file`. Use `domcontentloaded` + fixed waits, never `networkidle` (SPAs never settle). Dump the DOM to find the real button text before recording (user said "verify token" but the element was labeled "Verify CypherCoin"). Full recipe: `references/playwright-video-recording-and-spa.md`.
- **Headless = no video.** In headless mode (both Playwright chromium and camoufox) `record_video_dir` stays EMPTY — the dir is created but no .webm is written. To actually capture video run headful (on a server: `xvfb-run`) and give `page.close()` + `await browser.close()` time to finalize before globbing for `*.webm`. Always also take a final `page.screenshot()` as fallback proof even when recording — it survives headless.

## 4. Video CDN extraction from obfuscated streaming sites
Sites like streamrizz.com hide video URLs behind iframe chains and token-gated endpoints. Static HTML scraping fails because the CDN URL only appears deep in the rendering flow. Use Playwright network interception to capture intermediate `stream.php` / `player.php` endpoints, then extract the `<source>` tag from that response. CDN downloads often require `--referer` header to bypass 403 Forbidden. Full technique: `references/video-cdn-extraction-streamrizz.md`.

## Overlap note
`scrapling` (hub skill for Cloudflare/stealth) overlaps with this and is the heavier tool; use Bing RSS first for simple search, scrapling Stealth only when a site needs real-browser anti-bot bypass.