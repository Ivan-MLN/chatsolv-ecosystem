# TikTok API & Scraping Notes

## Profile Lookup via oEmbed (No Auth / No Bot Block)

TikTok's official `oembed` endpoint works reliably without proxies, login cookies, or stealth scrapers for basic user validation:

```bash
curl -s "https://www.tiktok.com/oembed?url=https://www.tiktok.com/@princessvionaa"
```

Response JSON shape:
```json
{
  "version": "1.0",
  "type": "rich",
  "title": "cio's Creator Profile",
  "author_url": "https://www.tiktok.com/@princessvionaa",
  "author_name": "cio",
  "width": "100%",
  "height": "100%",
  "provider_url": "https://www.tiktok.com",
  "provider_name": "TikTok",
  "embed_product_id": "princessvionaa",
  "embed_type": "profile"
}
```

This easily verifies display name (e.g. `@princessvionaa` -> `cio`) without hitting 403 blocks from tikwm/countik/raw HTML scrapers.

## tikwm API for Video Info & Download

tikwm API endpoint:
`https://www.tikwm.com/api/?url=<video_url>` (URL must be a video link, not a user profile link).
