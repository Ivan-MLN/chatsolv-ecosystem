# TikTok — worked example

Session: "find the currently viral TikTok video and send it to me".

## Discovery result

`https://www.tiktok.com/explore` logged out, Indonesian-IP server. First
`browser_navigate` snapshot came back `(empty page)` with `element_count: 0`.
A screenshot immediately after showed a fully-rendered 16-card grid — the
snapshot just raced hydration. The DOM query worked on the very next call
without any reload or wait.

Extracted 16 cards. `innerText` per card was the like count only. Top hit:

```
https://www.tiktok.com/@eggfamilydad/video/7665991461896310033   10.9M
https://www.tiktok.com/@pojokbogor/video/7666425615947435272     4.8M
https://www.tiktok.com/@digitalrealities1/video/7666987400568081672  3M
```

Explore also exposes category tabs (All, Singing & Dancing, Comedy, Sports,
Anime & Comics, Relationship, Shows, Lipsync) if the user wants a niche.

## yt-dlp outcome

`yt-dlp 2025.10.14`, both attempts failed identically:

```
[TikTok] 7665991461896310033: Downloading webpage
ERROR: [TikTok] 7665991461896310033: Unable to extract webpage video data
```

First attempt also warned:

```
WARNING: [TikTok] The extractor is attempting impersonation, but no
impersonate target is available.
```

Installing `curl_cffi` silenced that warning but did **not** fix extraction —
the failure is upstream page-shape drift, not a local dependency gap. Two
attempts is the right budget before falling through.

## tikwm response shape

```
GET https://www.tikwm.com/api/?url=<urlencoded>
{"code":0,"msg":"success","processed_time":0.51,"data":{...}}
```

`code` must be `0`. Observed `data` values for this video:

```
play        = https://v16m.tiktokcdn-us.com/...&btag=e000b8000   (no watermark)
hdplay      = None
play_count  = 163316831
digg_count  = 10924008
duration    = 15
region      = TH
author.unique_id = eggfamilydad
```

Note `digg_count` (10,924,008) matches the `10.9M` the explore card showed —
useful cross-check that you grabbed the right video.

Download with UA produced `3.8M`, `file` reported
`ISO Media, MP4 Base Media v1 [ISO 14496-12:2003]`.

## Quoting failure to avoid

This shape failed with `NameError: name 'k' is not defined` — the dict
comprehension inside a nested-quoted `python3 -c` inside an f-string:

```python
terminal(f'curl -s "..." -o /tmp/tw.json && python3 -c "import json;d=...;print(json.dumps({k:d.get(k) for k in [...]}))"')
```

Use a heredoc instead:

```bash
python3 - <<'EOF'
import json
d = json.load(open('/tmp/tw.json'))['data']
print(d['play'])
EOF
```

Also: `curl -s "<api>" | head -c 1200` emits `(23) Failed writing body`
because `head` closes the pipe early. Redirect to a file, then read.
