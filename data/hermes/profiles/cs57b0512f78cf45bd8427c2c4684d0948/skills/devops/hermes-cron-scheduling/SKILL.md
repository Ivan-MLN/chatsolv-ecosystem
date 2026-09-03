---
name: hermes-cron-scheduling
description: "Use when scheduling recurring Hermes cron jobs."
version: 1.0.0
platforms: [linux]
metadata:
  hermes:
    tags: [hermes, cron, scheduling, gateway, systemd, automation, whatsapp]
---

# Hermes cron scheduling (VPS workflow)

Use when the user wants a recurring task: daily reminder, scheduled group
message, periodic report, wake-up automation. Covers the full path from
schedule creation to making sure it actually fires.

## Core workflow

1. **Timezone first.** Cron interprets the schedule in the SYSTEM timezone.
   For WIB: `sudo timedatectl set-timezone Asia/Jakarta` (verify with `date`).
   Then `30 6 * * *` means 06:30 WIB, no offset math needed.
2. **Create the job.** Two modes:
   - Agent mode (prompt runs as an agent with tools):
     `hermes cron create "30 6 * * *" "description" --name my-job`
   - Script mode (deterministic, no LLM):
     `hermes cron create "30 6 * * *" "description" --name my-job --script my_script.py --no-agent --deliver origin`
   Scripts live in `~/.hermes/scripts/`. With `--no-agent`, the script's
   stdout is delivered verbatim; side effects (HTTP calls, MCP calls, file
   writes) happen inside the script itself.
3. **The gateway MUST be running or jobs never fire.**
   `hermes cron status` shows `✗ Gateway is not running` when it isn't.
   Install as a system service:
   ```
   sudo hermes gateway install --system
   ```
   **Pitfall (LXC/root VPS):** as root it refuses with
   `Refusing to install the gateway system service as root` — re-run with
   `sudo hermes gateway install --system --run-as-user root`. This is the
   correct fix in containers, not a blocker.
4. **Verify:** `hermes cron status` → `✓ Gateway is running`, `Next run:`
   shown in local (WIB) time. `systemctl is-active hermes-gateway`.

## Calling the nael-ai Baileys MCP from a cron script

A `--no-agent` script can POST JSON-RPC to the baileys MCP (port 5778) to act
on WhatsApp without any LLM in the loop. Two reference examples:

- `scripts/mcp_send_jsonrpc.py` — minimal urllib-based sender (used by
  nael-wakeup-0630 @all job)
- `references/spam_reminder.py` — full example with `requests`, error handling,
  and @mentions (fetches group metadata, builds mention array with @lid format,
  loops with per-message error checks)

Pattern (minimal):
```python
payload = {"jsonrpc":"2.0","id":66,"method":"tools/call",
           "params":{"name":"eval_code","arguments":{"jid":GROUP,"code":code}}}
urllib.request.urlopen(urllib.request.Request(MCP, data=json.dumps(payload).encode(),
    headers={"content-type":"application/json","accept":"application/json, text/event-stream"}))
```

## Pitfalls

- `hermes cron remove <id>` accepts only one job ID per command. Passing multiple IDs (e.g. `hermes cron remove id1 id2`) fails with `unrecognized arguments`. Chain multiple removals with `&&` or execute sequential removals in Python/bash.
- Forgetting the gateway install → cron "created" but never fires. Always
  check `hermes cron status` after creating.
- `--no-agent` scripts are silent unless they print — print/exit non-zero on
  error so failures are visible in cron logs.
- @all mentions in Baileys need `mentions: participants.map(p=>p.id)` from
  `sock.groupMetadata(jid)`; the plain `wa_send_message` MCP tool has no
  mentions param (see nael-eval-mcp skill for eval_code details).
- Re-scheduling: `hermes cron list` shows job IDs; changes take effect
  immediately while the gateway runs.

## Verify after setup
`hermes cron status` (gateway ✓ + next run), then either wait for the fire
time or trigger manually with the script (`python3 ~/.hermes/scripts/<name>.py`)
to confirm the side effect works before trusting the schedule.