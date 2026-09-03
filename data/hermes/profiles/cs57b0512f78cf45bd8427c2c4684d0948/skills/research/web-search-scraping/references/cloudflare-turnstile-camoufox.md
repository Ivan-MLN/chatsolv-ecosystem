# Cloudflare Turnstile bypass with camoufox (scrapling StealthyFetcher) — worked recipe

Worked in Aug 2026 against `https://aerolink.lat` (Cloudflare Turnstile, non-interactive
"managed challenge"). VPS = flagged datacenter IP. Real Chrome (headless + headful/xvfb,
stealth flags, UA/cookie tricks) sat on the challenge for 60s+. Camoufox cleared it in one
~3-10s poll. This is the validated approach; keep it as a template and adapt the URL / query.

## Prereqs (install once)
```bash
pip install "scrapling[all]"
scrapling install        # downloads camoufox + GeoIP + UBO addon to ~/.cache/camoufox
```

## Working async script
```python
import asyncio, pathlib
from scrapling.fetchers import StealthyFetcher

async def main():
    out = pathlib.Path('/root/aerolink_camo.png')

    async def grab(page):                       # page_action MUST be async AND return page
        # poll until the localized challenge clears
        for i in range(20):
            await page.wait_for_timeout(3000)
            title = await page.title()
            body = ''
            try:
                body = (await page.evaluate("document.body.innerText.slice(0,300)")) or ''
            except Exception:
                pass
            still = any(k in (title + body).lower() for k in [
                'just a moment', 'tunggu sebentar', 'verifikasi keamanan',
                'security verification', 'checking your browser'])
            print(f"  poll {i+1}: title={title!r} challenge={still}", flush=True)
            if not still:
                break
        await page.screenshot(path=str(out), full_page=True)   # screenshot INSIDE page_action
        print("shot taken", flush=True)
        return page            # CRITICAL: must return page, else it becomes None downstream

    page = await StealthyFetcher.async_fetch(
        'https://aerolink.lat',
        headless=True,
        humanize=True,
        block_webrtc=True,
        timeout=120000,        # scrapling timeout is ms here
        wait=2000,
        page_action=grab,
    )
    title = page.css_first('title')
    print("TITLE:", title.text if title else None, flush=True)
    print("SHOT:", out.exists(), out.stat().st_size if out.exists() else 0, flush=True)

asyncio.run(main())
```

## Pitfalls that cost time (all hit in practice)

1. **`solve_cloudflare=True` does not exist on scrapling 0.2.x** (0.2.99). It's newer-version
   API. Omitting it is fine — camoufox + `humanize=True` + `block_webrtc=True` clears Turnstile
   on its own.

2. **`page_action` MUST `return page`.** The engine does `page = await self.page_action(page)`.
   If your async callback returns None (default omit), the page var becomes None and the next
   internal call dies with:
   `AttributeError: 'NoneType' object has no attribute 'wait_for_timeout'` at camo.py line ~305.
   Add `return page` at the end of the callback.

3. **Screenshot goes INSIDE page_action.** The returned `Response` proxy object has no
   `save_screenshot` method — calling it on the Response raises `AttributeError`. Screenshot
   from inside the callback (or via `page.screenshot`) while you still hold the real page.

4. **Localized challenge title.** On an Indonesian-locale site the title is
   `"Tunggu sebentar..."`, not `"Just a moment..."`. An English-only detection regex falsely
   reports "challenge passed" instantly. Always match the localized keywords list above.

5. **Chrome headless AND headful both fail Turnstile from a DC IP.** Don't burn cycles
   tweaking puppeteer flags (`--disable-blink-features=AutomationControlled`, UA, cookies) —
   the Turner is non-interactive and fingerprints the browser; Chrome loses. Go straight to
   camoufox.

## Verification
Confirm success visually — title ≠ challenge text, `SHOT exists: True` with a real byte size,
and (with vision) the page content is the actual site (e.g. marketing landing page), not a
"Melakukan verifikasi keamanan" interstitial with a `Ray ID`.