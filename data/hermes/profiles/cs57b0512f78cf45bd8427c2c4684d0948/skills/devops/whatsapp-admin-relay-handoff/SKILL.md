---
name: whatsapp-admin-relay-handoff
description: "WhatsApp human escalation handoff and anonymous admin proxy."
version: 1.0.0
platforms: [linux]
metadata:
  hermes:
    tags: [whatsapp, handoff, escalation, relay, proxy, chatsolv]
---

# WhatsApp Admin Relay & Escalation Handoff

Use when designing or implementing human handoff / escalation where store admins take over bot conversations via their personal WhatsApp without exposing their personal phone numbers to customers.

## Architectural Flow

1. **Trigger Fallback / Escalation**
   - Agent encounters low confidence / missing knowledge in Obsidian Second Brain.
   - Conversation mode switches to `human` (`conversations.mode = 'human'`).
   - Bot sends fallback message to customer and fires an HMAC-authenticated outbound notification to the registered Admin's WhatsApp.

2. **Admin Notification Template**
   ```text
   ⚠️ [ESKALASI PERCAKAPAN]
   Pelanggan: <nomor_pelanggan>
   Percakapan ID: #CNV-<id_short>
   Pesan Terakhir: "<isi_pesan>"

   Ketik #ACC <id_short> untuk mengambil alih percakapan ini.
   ```

3. **Admin Commands & Seamless 2-Way Relay**
   - `#ACC <ID>`: Admin accepts takeover. Backend verifies caller is a registered admin/owner in workspace, creates Redis relay session (`relay:admin:<phone> -> conv_id`, `relay:conversation:<conv_id> -> admin_phone` with TTL), and replies with confirmation.
   - **Direct 2-Way Relay (Seamless Chat):** Once `#ACC` is accepted, **all messages from admin (plain text, images, documents/PDF, audio, voice notes) are directly relayed to the customer without needing a `#` prefix**. Backend dispatches the message to customer using the business WhatsApp bot (`POST /internal/v1/messages/send`).
   - **Customer to Admin Forwarding:** During active relay, all customer incoming messages are automatically forwarded to the admin's personal WhatsApp (`📩 [PESAN DARI PELANGGAN: ...]`).
   - `#DONE`, `#CLOSE`, or `#SELESAI`: Admin releases conversation. Backend deletes Redis relay keys, sets `conversations.mode = 'agent'`, and informs admin that bot has resumed.

4. **Bot Gateway Integration (`POST /internal/v1/messages/send`)**
   - In whatsmeow / Baileys bot service, implement an HMAC-authenticated POST endpoint (`/internal/v1/messages/send`) with payload `{ "channel_id": "...", "recipient": "...", "text": "..." }`.
   - Allows backend to push proactive messages (escalation alerts and admin relays) over the active socket without restarting or opening a second connection.

## Pitfalls & Edge Cases
- **Seamless vs Command Disambiguation:** Admin control commands (`#ACC`, `#DONE`, `#CLOSE`, `#SELESAI`) must be checked first before relaying raw text to the customer. Once in an active session, non-command messages are seamlessly treated as customer-facing messages.
- **Anonymous Relay:** Always route messages to customer through the bot channel socket so the customer sees replies from the official business account, never the admin's personal number.
- **Session Expiry:** Use Redis TTL on relay session keys to prevent orphaned lockouts if an admin forgets to `#DONE`.
