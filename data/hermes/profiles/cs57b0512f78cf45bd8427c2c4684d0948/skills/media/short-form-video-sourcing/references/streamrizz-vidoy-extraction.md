# Streamrizz, Epowner, and Slicedrive (vldey.ca) Video CDN Extraction

Extracting direct CDN MP4 streams and metadata from video hosting players like Streamrizz (`streamrizz.com`), Epowner (`epowner.com`), and Slicedrive (`vldey.ca`).

## 1. Streamrizz / Vidoy Player (`streamrizz.com`)

Handles `/e/<id>`, `/d/<id>`, and `/s/<id>` embed/watch links.

### Structure & Extraction Workflow

1. **Initial Page URL:** `https://streamrizz.com/s/<id>`, `/d/<id>`, or `/e/<id>`
   - HTML contains `var iframeId = '...'` and `var embedToken = '...'`.
2. **Inner Iframe URL:** `https://streamrizz.com/ip129jk?id=<iframeId>&t=<embedToken>`
   - Request with `Referer: https://streamrizz.com/` and Desktop UA.
   - Extract `playerPath` from JS: `const playerPath = "...";`.
   - **Important:** Clean unicode escapes in `playerPath` using `json.loads(f'"{m.group(1)}"')` to avoid `\u0026` Bad Request errors.
3. **Stream Player HTML:** Load `playerPath` with `Referer: https://streamrizz.com/ip129jk...`
   - HTML contains `<source src="https://mp4-xx.overfetch.video/<id_hash>" />` or `.m3u8` master playlist.

### Robust Python Extraction Snippet

```python
import urllib.request, re, json

def extract_streamrizz_cdn(page_url):
    headers = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        "Referer": "https://streamrizz.com/"
    }
    html = urllib.request.urlopen(urllib.request.Request(page_url, headers=headers)).read().decode('utf-8', errors='ignore')

    iframe_id = re.search(r"var iframeId = '([^']+)';", html).group(1)
    embed_token = re.search(r"var embedToken = '([^']+)';", html).group(1)
    iframe_url = f"https://streamrizz.com/ip129jk?id={iframe_id}&t={embed_token}"

    html2 = urllib.request.urlopen(urllib.request.Request(iframe_url, headers={**headers, "Referer": page_url})).read().decode('utf-8', errors='ignore')
    m = re.search(r'const playerPath\s*=\s*\"([^\"]+)\"', html2)
    player_path = json.loads(f'"{m.group(1)}"')

    html3 = urllib.request.urlopen(urllib.request.Request(player_path, headers={**headers, "Referer": iframe_url})).read().decode('utf-8', errors='ignore')

    video_src = re.search(r'<source src=\"([^\"]+)\"', html3).group(1)

    return {
        "direct_cdn": video_src,
        "player_path": player_path
    }
```

*Note: Accessing direct `overfetch.video` (`VidoyCDN-05`) CDN MP4/HLS links requires `Referer: https://streamrizz.com` (or `https://vidoy.com`) AND a byte-range request header (`Range: bytes=0-`) to avoid HTTP 403 Forbidden from the CDN server.*

```bash
# Direct download command example:
curl -sL -H "Referer: https://streamrizz.com" -H "Range: bytes=0-" -A "Mozilla/5.0" -o video.mp4 "<direct_cdn_url>"
```

---

## 2. Slicedrive / Vldey (`vldey.ca`, `silicmove.store`, `silicmedia.blog`)

- **URL patterns:**
  - `https://vldey.ca/<id>` -> direct CDN `https://cdn2.slicedrive.com/<id>.mp4`
  - `https://play.silicmove.store/<id>.mp4?k=...` or `https://play.silicmedia.blog/<id>.mp4?k=...`
- **Extraction Logic:**
  - The HTML player page sets `const VIDEO_BASE = "https://cdn.slicedrive.com";` and extracts filename from path.
  - **Direct CDN URL:** `https://cdn.slicedrive.com/<id>.mp4`

---

## 3. Videy / Twimg Redirects (`media.twimg.co.in`, `media.twimg.date`, `videy.co`)

- **URL patterns:**
  - `https://media.twimg.co.in/<id>.mp4`
  - `https://media.twimg.date/<id>.mp4`
- **Extraction Logic:**
  - HTML page contains embedded `<source>` or direct regex match `(https?://cdn2\.videy\.co/[^\s\"\'<>]+\.mp4)`.
  - **Direct CDN URL:** `https://cdn2.videy.co/<id>.mp4`

---

## 4. Epowner (`epowner.com`)

- **URL pattern:** `https://epowner.com/v/?id=<id>`
- Direct CDN source is available in the initial HTML response body via standard regex:
  `re.findall(r'(https?://cdn\.aceimg\.com/[^\s\"\'<>]+\.mp4)', html)`
- **Direct CDN URL:** `https://cdn.aceimg.com/<file_id>.mp4`

---

## 5. Download Pitfalls & Resuming

- Large MP4 files on `cdn2.videy.co` or `cdn.slicedrive.com` (70-100MB+) can choke or hit foreground tool timeouts.
- Always pass `-A "Mozilla/5.0"` user-agent.
- Use `curl -C - -sL` to resume interrupted/partial downloads if a download times out midway.


