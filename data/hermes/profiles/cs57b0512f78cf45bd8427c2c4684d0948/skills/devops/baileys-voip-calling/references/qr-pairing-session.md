# VoIP QR Pairing — Validated Workflow

## Context
August 7, 2026 session — first successful VoIP client pairing for nael-ai bot.

## Validated Sequence

1. **Clean state** (if re-pairing):
   ```bash
   rm -rf /tmp/nael-voip-auth/*
   rm -f /root/nael-ai/voip-qr.png
   ```

2. **Start init script in background**:
   ```bash
   cd /root/nael-ai && npx tsx scripts/init-voip.ts
   ```
   Watch for "✅ QR saved to: /root/nael-ai/voip-qr.png"

3. **Send QR immediately** (30-second validity window):
   - Target: RecoVerse group `120363186235853203@g.us` (user preference: shared visibility, not personal chat)
   - Tool: `mcp__baileys__wa_send_file`
   - Caption: "QR VoIP client — scan cepet sebelum expire (30 detik)"

4. **User scans** — init script prints "✅ VOIP client connected!" when pairing succeeds

5. **Keep init script running** to maintain connection during testing, or Ctrl+C after confirming connection

## Confirmed Details
- QR validity: ~30 seconds (Baileys default)
- Auth dir: `/tmp/nael-voip-auth/`
- QR file: `/root/nael-ai/voip-qr.png` (~6.8KB PNG)
- Connection check: look for `connection === "open"` in init script output

## Pitfall Avoided
Original attempt sent QR to user's personal chat; user redirected to group. Group delivery provides shared visibility for team-based operations.
