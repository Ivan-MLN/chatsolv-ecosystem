---
name: nael-eval-mcp
description: "Use for running JS/shell inside nael-ai bot via eval MCP."
version: 1.0.0
platforms: [linux]
metadata:
  hermes:
    tags: [nael-ai, baileys, mcp, eval, whatsapp]
---

# nael-ai eval MCP (port 5778)

Use when you need to run arbitrary JS/shell **inside the nael-ai WhatsApp bot
process**, especially to act on WhatsApp (fetch & send profile pics, inspect bot
state, send messages) **without stopping/restarting the bot**.

## Why this exists
Older approach stopped the bot (`pm2 stop nael-ai`) then ran a standalone
Baileys script — that kicked the live session (conflict/replaced) and needed a
restart. `eval_code`/`eval_shell` run in the already-connected process, so no
downtime, no conflict.

## TikTok downloads — use the wa_tiktok tool, NOT yt-dlp/tikwm
The bot has a dedicated `wa_tiktok` MCP tool (premierely.io tiktok-api, the same
API as the `/tiktok` command — owner-mandated as the ONLY TikTok source; do NOT
fall back to yt-dlp or tikwm for nael-ai). It downloads the video + audio and
sends BOTH to the given `jid`, returning metadata. Example:
```
{"name":"wa_tiktok","arguments":{"url":"https://vt.tiktok.com/XXXX/","jid":"120363186235853203@g.us"}}
```
Owner rule: TikTok results go to group RecoVerse Team (120363186235853203@g.us)
without being asked, unless the user names another chat.

## The two tools
- `eval_code` — async JS via `new AsyncFunction("sock", "jid", code)`. `sock`
  = the live Baileys socket (can `sendMessage`, `profilePictureUrl`, …).
  `jid` = the "jid" argument you passed (may be undefined).
- `eval_shell` — shell command on the host (stdout+stderr combined).

Call them by POSTing JSON-RPC to `http://127.0.0.1:5778`:
```
curl -s -X POST http://127.0.0.1:5778 -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"eval_code","arguments":{"code":"...","jid":"<grp>"}}}'
```
Result lives at `.result.content[0].text`. `.isError` true on failure.

## Pitfalls
- **Do NOT `const jid = …` in the code** — `jid` is already a parameter, so
  redeclaring throws `Identifier 'jid' has already been declared`. Decalare
  under a different name, or just pass `jid` in arguments and use it directly.
- Timeout default is `config.execTimeoutMs`; override via `timeoutMs` arg.
- `eval_code` output is clipped to 4000 chars, `eval_shell` to 16000.
- Sending media: build the buffer (e.g. `fetch()` the PP url) then
  `sock.sendMessage(jid, { image: buf, caption })` — Baileys picks image/video/
  audio/document from the message key, not the mimetype.

## Verify after editing
`npm run build` + `npm run typecheck` must exit 0, `pm2 restart nael-ai`,
then `curl tools/list` to 5778 to confirm the new tool(s) are exposed.