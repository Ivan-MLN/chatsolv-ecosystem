# Fingerprint-Based Bot Detection: Detect & Handle

Covers the case where a site does NOT use Cloudflare/captcha but **fingerprints the
visiting browser** server-side and silently degrades flagged browsers. Distinct from
Turnstile (section 3): here there is no challenge page at all — every API returns HTTP
200, just with tampered payloads.

## Signature / diagnostic

The single most reliable signal is **payload shape changing between browsers (same URL,
same deterministic flow)**:

| Browser | XHR/fetch API response |
|---------|------------------------|
| headless Chrome (playwright/puppeteer) | every endpoint 200 but body = one field `{"e":"AejS...<long base64>"}` |
| camoufox (`StealthyFetcher`/`AsyncCamoufox`) | same endpoints return plain structured JSON `{"success":true,...}` |

Encrypted payload in Chrome + readable JSON in camoufox on the identical flow = the
server fingerprinted the browser and, for the flagged one, wrapped responses in a channel
the page's own JS cannot decrypt → the app shows a brittle error ("Security verification
failed", "Session invalid", "Please try again", "Establishing secure session…" stuck) even
though HTTP status is 200.

Look in the network log for a fingerprint probe set — presence strongly indicates
fingerprinting before real data is served:
- `api/v1/fingerprint/sync`
- `api/v1/geo/resolve`
- `api/v1/telemetry/report`
- `api/v1/config/fetch`
- `api/v1/user/session`, `api/auth/status`, `api/payment/validate`

## How to prove it instead of guessing

1. Run the IDENTICAL scripted flow in headless Chrome and in camoufox.
2. Log all XHR/fetch request+response bodies to a JSONL file in both runs.
3. Diff: Chrome shows `{"e":"..."}`/`<err>` on the sensitive endpoints; camoufox shows
   plain JSON. That diff IS the proof it was bot detection, not a bad token/URL.

## Handling

- **Go straight to camoufox** for the whole flow. Don't keep tuning Chrome stealth flags
  (randomize UA, `--disable-blink-features`, navigator.webdriver spoof, locale) — a real
  fingerprint sync beats all of those. Scrapling `StealthyFetcher` (`async_fetch`)
  `/ `AsyncCamoufox` is the reliable path (0.2.x has no `solve_cloudflare` kwarg).
- **Token guessing is the wrong fix.** Once the fingerprint gate falls (plaintext JSON)
  but the action still hangs ("Verifying…", "Establishing secure session…"), the wall is
  gone and the *credential* is invalid/expired. Stop chasing bypass now → the account
  owner must supply a valid token. Do not loop "keep trying".

## Network-logging snippet (Playwright / Camoufox)

```python
import asyncio, json

async def log_req(req):
    if req.resource_type in ('xhr', 'fetch'):
        post = ''
        try: post = req.post_data or ''
        except Exception: pass
        with open('/tmp/net.jsonl', 'a') as f:
            f.write(json.dumps({'t': 'req', 'url': req.url, 'post': post[:400]}) + '\n')

async def log_res(res):
    if res.request.resource_type in ('xhr', 'fetch'):
        body = ''
        try: body = (await res.body()).decode('utf-8', 'replace')[:600]
        except Exception as e: body = f'<err {str(e)[:80]}>'
        with open('/tmp/net.jsonl', 'a') as f:
            f.write(json.dumps({'t': 'resp', 'url': res.url,
                                'status': res.status, 'body': body}) + '\n')

page.on('request',  lambda r: asyncio.ensure_future(log_req(r)))
page.on('response', lambda r: asyncio.ensure_future(log_res(r)))
```

Notes:
- `resource_type in ('xhr','fetch')` — skips asset noise.
- `res.body()` can raise/empty for cache-served or already-consumed responses; that's why
  the request `post_data` is logged too. Read bodies immediately in the handler.
- Encrypted bodies are why you can't just grep for the submitted token/text in the log —
  the real payload rides inside the `{"e":...}` blob.

## Video recording pitfall (headless)

- `record_video_dir` produces **no .webm in headless mode** (dir stays empty) for both
  Playwright chromium and camoufox. Capture requires headful — on a server: run under
  `xvfb-run`, then `page.close()` + `await browser.close()`, then glob `*.webm` in the dir.
- Always also save a final `page.screenshot()` — it works headless and gives proof of the
  final state even when video comes up empty.