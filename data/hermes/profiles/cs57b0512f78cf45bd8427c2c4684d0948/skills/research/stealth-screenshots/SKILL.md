---
name: stealth-screenshots
description: Use when screenshotting Cloudflare-protected sites.
version: 1.0.0
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [Screenshot, Cloudflare, Stealth, Browser, Playwright]
    related_skills: [scrapling, browser-vision]
prerequisites:
  commands: [python3]
  python_packages: [scrapling]
---

# Stealth Screenshots

Capture screenshots from bot-protected or Cloudflare-challenged sites using scrapling's StealthyFetcher with Playwright page_action.

## When to Use

- Site shows Cloudflare challenge page to puppeteer/playwright
- Regular browser tools hit bot detection
- Need full-page + element screenshots from protected sites
- Site requires real browser fingerprint to load content

## Quick Example

```python
from scrapling.fetchers import StealthyFetcher

def take_screenshots(page):
    import time
    time.sleep(3)  # Wait for animations
    
    page.screenshot(path="/tmp/full_page.png", full_page=True)
    
    element = page.query_selector("#target")
    if element:
        element.screenshot(path="/tmp/element.png")
    
    return page  # MUST return page to prevent None error

page = StealthyFetcher.fetch(
    "https://protected-site.com",
    headless=True,
    network_idle=True,
    wait=0,
    timeout=60000,
    page_action=take_screenshots
)
```

## Parameters

| Parameter | Default | Purpose |
|-----------|---------|---------|
| `headless` | `True` | Run browser without UI |
| `network_idle` | `False` | Wait for no network activity (recommended: `True`) |
| `wait` | `0` | Extra wait in ms AFTER page_action (set `0` if using time.sleep inside) |
| `timeout` | `30000` | Total timeout in milliseconds |
| `block_webrtc` | `False` | Block WebRTC (helps stealth) |
| `disable_resources` | `False` | Block fonts/images/media for speed |

## Full-Page + Element Screenshot

```python
from scrapling.fetchers import StealthyFetcher

def capture_both(page):
    import time
    time.sleep(3)
    
    # Full page
    page.screenshot(path="/tmp/full.png", full_page=True)
    
    # Specific element by CSS selector
    card = page.query_selector("#cardStep1")
    if card:
        card.screenshot(path="/tmp/card.png")
    else:
        print("Element not found")
    
    return page

StealthyFetcher.fetch(
    "https://example.com",
    headless=True,
    network_idle=True,
    wait=0,
    timeout=60000,
    page_action=capture_both
)
```

## High-Resolution Screenshots

```python
def capture_hires(page):
    page.set_viewport_size({"width": 1560, "height": 1080})
    page.screenshot(path="/tmp/hires.png", full_page=True)
    return page

StealthyFetcher.fetch(url, headless=True, page_action=capture_hires)
```

## Pitfalls

- **MUST return page from page_action** — if you don't return `page`, scrapling raises `AttributeError: 'NoneType' object has no attribute 'wait_for_timeout'`
- **Use time.sleep(), not page.wait_for_timeout()** — Playwright's wait_for_timeout works but time.sleep is simpler
- **Set wait=0 if using time.sleep inside page_action** — otherwise you wait twice (inside + after)
- **network_idle=True recommended** — ensures page fully loads before screenshots
- **timeout in milliseconds** — 60000 = 60 seconds, not 60 like terminal timeout
- **scrapling parameter names differ from skill docs** — `solve_cloudflare` doesn't exist; stealth is automatic with StealthyFetcher
- **Element may not exist** — always check `if element:` before calling `.screenshot()`

## Verification

```bash
ls -lh /tmp/*.png
```

## See Also

- scrapling skill — general web scraping with stealth
- browser_vision tool — screenshot + OCR built-in tool (but hits Cloudflare)
