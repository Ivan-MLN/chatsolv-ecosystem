---
name: whatsapp-scheduled-jobs
description: "Schedule WhatsApp bot actions via Hermes cron + eval MCP."
version: 1.0.0
platforms: [linux]
metadata:
  hermes:
    tags: [nael-ai, hermes, cron, whatsapp, baileys, mcp, scheduling]
---

# Deliver scheduled WhatsApp jobs (Hermes cron -> eval MCP)

Use when the user wants the nael-ai WhatsApp bot to act automatically at a
schedule — reminders, @all broadcasts, wake-up pings, periodic reports.
Wires the durable Hermes cron daemon to the live Baileys bot through its eval
MCP. The user-owned `nael-eval-mcp` skill covers immediate eval; this one is
about *scheduled* execution through the daemon.

## Model: script POSTs JSON-RPC to the bot MCP
A cron job runs a script; the script POSTs to the eval MCP on the already-
connected Baileys socket (no restart, no second connection — a 2nd socket kills
the live bot).

```bash
hermes cron create "30 6 * * *" "daily wakeup @all" \
  --name wakeup-0630 --script nael_wakeup_all.py --no-agent
```

- Scripts live in `~/.hermes/scripts/` (root host: `/root/.hermes/scripts/`).
- `--no-agent`: the script IS the job, stdout delivered verbatim. For a pure
  side-effect script, print nothing → silent no-agent job.
- Works regardless of deliver target since the script sends its own WA message.

## Mandatory setup steps (or cron silently never fires)

1. **Timezone is the SYSTEM zone.** If the user says "WIB", set it first:
   `timedatectl set-timezone Asia/Jakarta` → verify `date`. This host is WIB,
   so `30 6 * * *` fires 06:30 WIB. Cron is NOT UTC.
2. **The Hermes gateway must be running.** `hermes cron status` tells you.
   On this root/LXC-style host install it as a system service:
   `sudo hermes gateway install --system --run-as-user root`
   Without `--run-as-user root` it refuses ("Refusing to install the gateway
   system service as root"). Then `systemctl is-active hermes-gateway` →
   `active`, and `hermes cron status` shows "✓ Gateway is running". Confirmed.
3. Confirm: `hermes cron list` shows the job with its `Next run`.

## Sending a real @all mention
The `wa_send_message` MCP tool has only `jid`+`text` — **no `mentions` field**,
so `@` tags never render through it. To actually tag all members use the MCP's
`eval_code` with Baileys group metadata:
```js
const meta = await sock.groupMetadata(jid);
const parts = (meta.participants || []).map(p => p.id);
await sock.sendMessage(jid, { text: "...", mentions: parts });
```
`participants[].id` may be `@lid` handles — mention as-is.

See `scripts/spam_reminder.py` for a complete working Python example that wraps
this pattern in a cron-ready script.

## MCP call shape
POST to `http://127.0.0.1:5778/mcp`:
```json
{"jsonrpc":"2.0","id":1,"method":"tools/call",
 "params":{"name":"eval_code","arguments":{"code":"...","jid":"<grp>"}}}
```
- Result at `.result.content[0].text`; `.isError` true on failure.

## Pitfalls
- `--run-as-user root` is mandatory on the root host, else the gateway install
  hard-fails.
- Cron schedule uses system TZ — set the TZ before choosing the cron string.
- **Never print secrets** — a `--no-agent` script prints stdout verbatim and
  empty stdout is silent. Read tokens/keys from env or file; return status only.
- For a no-spam smoke test, eval a safe no-op (e.g. `return 42`) instead of the
  real broadcast.
- **Do NOT call `wa_group_metadata` then `wa_send_message` separately** — the
  MCP tool returns empty `.result` (no participants), so mentions will be `[]`.
  Always use `eval_code` with `sock.groupMetadata(jid)` inside the code block;
  participants come back populated and `sock.sendMessage()` has the mentions
  param. The standalone tools are read-only wrappers; real actions need eval.

## Verify
`hermes cron list` (job present, next run) and `journalctl -u hermes-gateway -f`
(ticker heartbeat).

## Validate the model backend before relying on a broadcast job
A scheduled job depends on the router/model the bot talks through. If the user
switched models around the same time, probe 9router before trusting the next run:
```
curl -s http://127.0.0.1:20128/v1/models                 # lists model ids
curl -s --max-time 90 http://127.0.0.1:20128/v1/chat/completions \
  -H 'content-type: application/json' \
  -d '{"model":"claude-gratisan","messages":[{"role":"user","content":"ok"}],"max_tokens":20}'
```
~31s (HTTP 200) is NORMAL for the puppeteer-backed `claude-gratisan`; if it errors,
revert to `deepseek-gratisan` before the scheduled broadcast fires.