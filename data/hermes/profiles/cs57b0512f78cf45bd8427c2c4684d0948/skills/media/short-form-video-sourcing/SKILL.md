---
name: short-form-video-sourcing
description: Find/download trending TikTok, Reels, Shorts videos.
tags: [tiktok, video, download, trending, yt-dlp, scraping]
---

# Short-Form Video Sourcing

Find trending short-form videos (TikTok / Reels / Shorts) and pull the raw MP4 to disk.

**OWNER MANDATE (nael-ai): for TikTok, the bot has a dedicated MCP tool `wa_tiktok`
or fallback to `www.tikwm.com/api/?url=` (desktop UA, wajib `www.tikwm.com`).**
Note: premierely.io can be down; fallback tikwm is verified working with `www.tikwm.com`.
For browser TikTok FYP navigation via Puppeteer with user cookies, puppeteer/chromium with user cookies can extract video links from `/foryou` or `@me`.

Two independent phases. Do not conflate them — discovery failing does not mean download will fail, and vice versa.

## Phase 1 — Discovery (what is trending)

TikTok Explore renders server-side and is readable **logged out**. No login wall for browsing.

1. `browser_navigate("https://www.tiktok.com/explore")`
2. The first snapshot is often `(empty page)` — hydration lag. Do **not** conclude "blocked". Take a screenshot or just run the DOM query; content is usually there.
3. Extract candidates with one console call:

```js
Array.from(document.querySelectorAll('a[href*="/video/"]'))
  .slice(0,20)
  .map(a => ({href: a.href, txt: a.innerText.slice(0,80)}))
```

`innerText` on the explore card is the **like count** (e.g. `10.9M`). Pick the highest to answer "most viral".

Explore is geo-shaped — an Indonesian-IP server returns mostly Indonesian/SEA content. Say so if the user expected global trends.

## For nael-ai (WhatsApp bot) — use wa_tiktok MCP tool or tikwm fallback
If this download is for the nael-ai bot:
1. First try `wa_tiktok` MCP tool (port 5778).
2. If `wa_tiktok` fails or returns `premierely.io` errors (e.g. "TikTok is temporarily unavailable"), fallback immediately to:
   - DDG HTML Search (`site:tiktok.com <query>`) to locate video URLs.
   - Query `https://www.tikwm.com/api/?url=<video_url>` (with Desktop UA).
   - Download `play` (video MP4) and `music` (audio MP4).
   - Convert audio to MP3: `ffmpeg -y -i audio.mp4 -vn -ar 44100 -ac 2 -b:a 192k audio.mp3`.
   - Send video and audio files via `wa_send_file` (`jid` & `path` parameters).

## Profile & User Video Listing

To resolve a TikTok username/handle (`@username`), check profile existence, or get display name (`author_name`) without hitting 403/bot walls:

```bash
curl -s "https://www.tiktok.com/oembed?url=https://www.tiktok.com/@<username>"
```

Returns JSON with `author_name` (display name), `title` (e.g. `"cio's Creator Profile"`), `author_url`, and `embed_product_id`.

To list user videos when tikwm direct user API returns 403 or profile scraping is blocked:
1. **Serper API (Fastest & recommended):** Query `site:tiktok.com/@<username> video` via Serper API (`https://google.serper.dev/search`). Returns exact TikTok video URLs for that specific handle.
2. **DuckDuckGo HTML Search:** Search `<username> tiktok video` on `https://html.duckduckgo.com/html/?q=` and extract video URLs using regex:

```python
import urllib.request, urllib.parse, re

query = f'{username} tiktok video'
url = 'https://html.duckduckgo.com/html/?q=' + urllib.parse.quote(query)
req = urllib.request.Request(url, headers={'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)'})
with urllib.request.urlopen(req) as resp:
    html = resp.read().decode('utf-8', errors='ignore')
    links = re.findall(r'https%3A%2F%2Fwww\.tiktok\.com%2F%40[^%&\"\']+%2Fvideo%2F\d+', html)
    video_links = list(set([urllib.parse.unquote(l) for l in links]))
```

To list all recent video URLs and captions from a TikTok profile without scraping (if yt-dlp installed):

```bash
yt-dlp -j --flat-playlist "https://www.tiktok.com/@<username>"
```

Each line in stdout is JSON containing `url` / `webpage_url` and `title` (caption).

## Finding TikTok Videos by Keywords (Search without API keys)

Direct TikTok search API endpoints (e.g. tikwm search `/api/feed/search`) are frequently blocked by Cloudflare bot protection/JS challenges.

To reliably find TikTok video links by topic/keywords:
1. **Serper API (Fastest & most reliable):** Query Serper (`https://google.serper.dev/search`) with `site:tiktok.com <query>` (e.g. `site:tiktok.com cewek joget viral fyp`). Note: When searching for profane/explicit meme queries, raw profanity might return zero direct organic video results; broaden query terms (e.g., `sound toxic fyp` or `preset toxic am`), then filter/inspect snippet texts.
2. **DuckDuckGo HTML search (Fallback):** Use DuckDuckGo HTML search with `site:tiktok.com`.

### Python Sourcing & Download Pipeline (Serper + tikwm + wa_send_file)

When `wa_tiktok` MCP tool is unavailable or returning errors ("TikTok is temporarily unavailable"):

```python
import urllib.request, urllib.parse, json, random, os

SERPER_KEY = "bcdfe9657adfe971132dd3929c88f30f4ee812b1" # from 9router DB
MCP_URL = "http://127.0.0.1:5778"
TARGET_JID = "120363186235853203@g.us"

def get_tiktok_link(query):
    url = "https://google.serper.dev/search"
    payload = json.dumps({"q": f"site:tiktok.com {query}"}).encode('utf-8')
    headers = {"X-API-KEY": SERPER_KEY, "Content-Type": "application/json"}
    req = urllib.request.Request(url, data=payload, headers=headers)
    res = json.loads(urllib.request.urlopen(req).read().decode('utf-8'))
    links = [item['link'] for item in res.get('organic', []) if '/video/' in item.get('link', '')]
    return random.choice(links) if links else None

def download_and_send_wa(video_url, jid=TARGET_JID):
    api_url = f"https://www.tikwm.com/api/?url={urllib.parse.quote(video_url)}"
    headers = {"User-Agent": "Mozilla/5.0", "Referer": "https://www.tikwm.com/"}
    res = json.loads(urllib.request.urlopen(urllib.request.Request(api_url, headers=headers)).read().decode('utf-8'))
    if res.get('code') == 0:
        play_url = res['data']['play']
        title = res['data'].get('title', '')
        video_path = "/tmp/tiktok_temp.mp4"
        with urllib.request.urlopen(urllib.request.Request(play_url, headers=headers)) as resp, open(video_path, 'wb') as f:
            f.write(resp.read())
        
        # Send via wa_send_file MCP
        payload = json.dumps({
            "jsonrpc": "2.0", "id": 1, "method": "tools/call",
            "params": {"name": "wa_send_file", "arguments": {"jid": jid, "path": video_path, "caption": f"✨ *TikTok Video* ✨\n\n📌 {title}"}}
        }).encode('utf-8')
        mcp_req = urllib.request.Request(MCP_URL, data=payload, headers={"Content-Type": "application/json"})
        return urllib.request.urlopen(mcp_req).read().decode('utf-8')
```

### Bot vs Agent Sourcing Distinction (nael-ai)
- **Bot `/play-tiktok` command:** Uses YouTube (`yts`) + Savenow fallback for quick keyword search and high audio download reliability without scraping overhead.
- **Cathlyne / Agent workflow:** Uses DDG HTML search (`site:tiktok.com <query>`) -> extracts exact TikTok URL -> fetches via `tikwm.com/api/?url=` (or `yt-dlp`) -> extracts raw video/audio with FFmpeg -> sends via `wa_send_file`.

## Phase 2 — Download

Try in this order, stop at the first that works.

### 2a. yt-dlp (preferred)

```bash
pip install -q yt-dlp curl_cffi
yt-dlp -t mp4 -o "viral.%(ext)s" "<video_url>"
```

Use `-t mp4`, not `-f mp4` (yt-dlp warns on the latter). `curl_cffi` supplies the impersonation target yt-dlp asks for.

If it errors `Unable to extract webpage video data`, that is TikTok rotating its page shape — go to 2b, do not debug it.

### 2b. tikwm API (reliable fallback, no watermark)

```bash
curl -s "https://www.tikwm.com/api/?url=<urlencoded_video_url>" -o /tmp/tw.json
```

Response `data` fields worth reading: `play` (no-watermark MP4 URL), `hdplay` (often `null`), `wmplay`, `duration`, `play_count`, `digg_count`, `region`, `author.unique_id`, `title`.

Then fetch, with a UA — the CDN 403s a bare curl:

```bash
curl -sL -A "Mozilla/5.0" -o /tmp/tiktok_viral.mp4 "<data.play>"
```

Verify: `file /tmp/tiktok_viral.mp4` should say `ISO Media, MP4`. A few-KB file means the CDN URL expired — re-hit the API, links are short-lived and signed.

## Delivery & Audio Processing

- **RecoVerse WhatsApp Workflow:** When requested to send a TikTok/short-form video to the WhatsApp group, send the video file directly to the RecoVerse group (`120363186235853203@g.us`) using `wa_send_file` along with its original TikTok URL included in the caption/message.
- **Audio Only Request:** If the user asks "mau denger", "lagunya kek apa", or "mana lagunya" (asking for audio rather than video), extract/convert to MP3 and send ONLY the audio (.mp3) file to the RecoVerse group along with the original link.
- `wa_send_file` can send local media files (MP4, PNG, MP3) directly to any WhatsApp JID (user or group) when explicitly requested.
- `wa_send_message` is **text-only** — it cannot attach media. Give the local path plus `MEDIA:<path>` so the client renders it inline when replying in chat session.

### FFmpeg Processing & Repair Snippets

- **Fix Sticking/Freezing Web Video (NAL Unit / Decode Corruption):**
  Web CDN videos (e.g., slicedrive) may suffer from corrupt NAL unit splits or non-standard H.264 streams, causing video playback to freeze on mobile/WhatsApp. Re-encode to clean standard H.264:
  `ffmpeg -y -i input.mp4 -c:v libx264 -preset fast -crf 22 -pix_fmt yuv420p -movflags +faststart -c:a copy output_fixed.mp4`
  *(See `references/ffmpeg-video-repair.md` for full diagnosis & verification commands).*

- **Convert video/audio to MP3:**
  `ffmpeg -y -i input.mp4 -vn -acodec libmp3lame -q:a 2 output.mp3`
- **Slowed + Reverb (without high-pitch/cempreng audio):**
  Use pitch reduction (`asetrate` low factor like 0.75-0.80) combined with a lowpass filter (e.g. `lowpass=f=3500`) to remove high-frequency squeaks/cempreng, and `aecho` for reverb:
  `ffmpeg -y -i input.ogg -af "asetrate=44100*0.75,aresample=44100,lowpass=f=3500,aecho=0.8:0.88:60:0.4" -acodec libmp3lame -q:a 2 output.mp3`

## Pitfalls

- Piping the tikwm JSON straight into `head -c` gives `Failed writing body` from curl. Write to a file first, then parse.
- Do not build the curl command as a nested Python one-liner inside an f-string — quoting collapses. Heredoc (`python3 - <<'EOF'`) or a real file.
- URL-encode the video URL in the API query string.

## Reference

`references/tiktok-notes.md` — worked example with real field values and observed API shape.
`references/tiktok-profile-oembed.md` — lightweight TikTok user profile lookup via official oEmbed API.
`references/ffmpeg-video-repair.md` — diagnosis and repair commands for corrupt video streams (NAL unit / decode errors).
`references/streamrizz-vidoy-extraction.md` — extraction workflow for Streamrizz / Vidoy, Epowner, and Slicedrive (vldey.ca) video CDN MP4 streams.
