# Video CDN Extraction: Streamrizz Pattern

## Problem
Sites like `streamrizz.com` obfuscate video CDN URLs through multiple layers:
- Main page loads an iframe with dynamic token
- Iframe loads another endpoint (`stream.php`) that renders a player page
- Player page contains the actual CDN URL in a `<source>` tag
- Direct CDN downloads require proper `Referer` header to bypass 403 Forbidden

Simple static scraping (curl HTML → grep for .mp4) returns nothing because the video URL is only present deep in the iframe chain.

## Working Extraction Path

### 1. Network interception with Playwright
Use Playwright to intercept network requests and catch the intermediate `stream.php` endpoint:

```python
from playwright.sync_api import sync_playwright

video_urls = []

def handle_request(request):
    url = request.url
    # Filter for stream endpoints (not ads/analytics)
    if 'stream.php' in url or any(ext in url.lower() for ext in ['.mp4', '.m3u8']):
        if 'google-analytics' not in url:
            video_urls.append(url)
            print(f"[STREAM] {url}")

with sync_playwright() as p:
    browser = p.chromium.launch(headless=True)
    page = browser.new_page()
    page.on('request', handle_request)
    
    page.goto('https://streamrizz.com/d/<id>', wait_until='domcontentloaded', timeout=30000)
    time.sleep(3)  # Let iframe/scripts load
    
    browser.close()

# Look for stream.php?bucket=... in captured URLs
```

### 2. Extract CDN URL from stream.php
The `stream.php` endpoint returns a full HTML player page with the video source:

```bash
curl -s 'https://streamrizz.com/stream.php?bucket=vidoycdn&id=<id>' \
  -H 'User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36' \
  -H 'Referer: https://streamrizz.com/' \
  | grep -i 'source' | grep -oP 'src="([^"]+)"'
```

Typical result:
```html
<source src="https://mp4-02.overfetch.video/C0rzg-TELE @NEWASUPAN102.mp4" playsinline webkit-playsinline/>
```

### 3. Download with proper referer
CDN checks the Referer header — direct wget/curl without it returns 403:

```bash
wget -O output.mp4 'https://mp4-02.overfetch.video/<hash>-<filename>.mp4' \
  --user-agent='Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36' \
  --referer='https://streamrizz.com/'
```

**Without `--referer`**: `403 Forbidden`  
**With `--referer`**: Download succeeds

## Key Patterns

1. **Iframe chain**: Main page → iframe with token → `stream.php` → CDN URL
2. **Network interception beats static HTML parsing**: The CDN URL never appears in the main page source
3. **Referer requirement**: Video CDNs commonly enforce referer checks to prevent hotlinking
4. **Stream endpoint naming**: Look for paths like `stream.php`, `player.php`, `embed.php` with `bucket=` or `id=` params

## Similar Sites
This pattern applies to many video hosting/streaming aggregators that use:
- Token-gated iframes
- Separate player rendering endpoints
- CDN referer protection

Adapt the network interception approach and always check CDN download requirements (referer, token headers, cookies).

## Session Reference
Extracted from streamrizz.com/d/adq048k7s3mg (2026-08-16). Video: 103MB MP4, required referer for download.
