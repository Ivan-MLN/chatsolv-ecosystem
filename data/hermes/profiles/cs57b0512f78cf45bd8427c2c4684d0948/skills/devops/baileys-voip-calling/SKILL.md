---
name: baileys-voip-calling
description: "Place WhatsApp voice calls with audio via Baileys VoIP."
version: 1.0.0
platforms: [linux]
metadata:
  hermes:
    tags: [baileys, whatsapp, voip, calling, nael-ai]
---

# Baileys VoIP Voice Calling

Use when you need to place WhatsApp voice calls programmatically, optionally playing audio files (MP3/WAV) during the call.

## Architecture & Socket Modes

**1. Socket Reuse / Passing Active Socket (No Re-pairing)**
`baileys-caller`'s `SignalingBridge` uses the Baileys WebSocket (`sock.ws`) for WASM VoIP signaling. By accepting or reusing the main bot's active `sock`, VoIP calls can be routed over the existing WhatsApp connection without opening a second socket or creating a separate session.

**2. Dedicated VoIP Auth Session (Separate Socket)**
When `VoipClient` initializes its own socket with `new VoipClient({ authDir })`:
- **One auth session = one socket.** Pointing `authDir` to the main bot's session directory (`/var/lib/nael-workspace/.sessions.json`) causes socket conflicts and 515 disconnects.
- `authDir` must point to a separate directory (e.g. `/root/nael-ai/.voip-auth`) paired once via QR.

## Setup: Initial Pairing

VoIP client needs its own auth session. Use the init script pattern:

1. **Create init script** (`scripts/init-voip.ts`):
   - Use `useMultiFileAuthState()` with a dedicated auth dir (e.g. `/root/nael-ai/.voip-auth`)
   - Generate QR code with `qrcode` npm package when `connection.update` fires with `qr`
   - Save QR as PNG file
   - Send the QR image to user via the main bot socket (or manually)
   - Keep script running after pairing to maintain connection

2. **Dependencies**:
   ```bash
   npm install --save qrcode
   npm install --save-dev @types/qrcode
   ```

3. **Run pairing**:
   ```bash
   npx tsx scripts/init-voip.ts
   ```
   User scans QR with WhatsApp number designated for calling.

## MCP Tool: wa_voice_call

Once paired, the `wa_voice_call` tool (via nael-ai MCP baileys at port 5778) can place calls:

**Parameters**:
- `phoneNumber` (required) — Target phone number, digits only (e.g. `"6283893964069"`)
- `audioSource` (optional) — Absolute path to MP3/WAV file to play during call, or `"silence"` for silent call (default: silence). Custom TTS audio can be generated beforehand via Cartesia API (`https://api.cartesia.ai/tts/bytes` with `model_id: "sonic-3.5"`, `output_format: {container: "mp3", encoding: "mp3", sample_rate: 44100}`).
- `durationMs` (optional) — Auto-hangup timeout in milliseconds (default: 30000)

**Example**:
```json
{
  "name": "wa_voice_call",
  "arguments": {
    "phoneNumber": "6283893964069",
    "audioSource": "/root/nael-ai/audio-call-test.mp3",
    "durationMs": 60000
  }
}
```

The call rings the target. If they answer, the audio plays. Call auto-hangs up after `durationMs`.

## Group Voice Calling (Joinable Group Calls)

To place a group voice call in `baileys-caller`:
1. Fetch group participants via `sock.groupMetadata(groupJid)`.
2. Discover peer devices for all participants using `signaling.discoverPeerDevices(pJid)`.
3. **Filter out non-hosted device IDs**: Devices with ID `:99` (e.g., `...:99@s.whatsapp.net`) cause `jidToSignalProtocolAddress` in `libsignal` to throw `Unexpected non-hosted device JID with device 99`. Exclude them: `devices.filter(d => !d.includes(':99@'))`.
4. Ensure Signal sessions for all discovered peer devices via `signaling.ensureSessionsForPeers(allDeviceJids)`.
5. Trigger WASM engine startCall with `peerJid: groupJid, peerPn: groupJid, peerList: allDeviceJids, isLidCall: false`.

## Implementation Notes

- VoipClient is imported from `baileys-caller` package (assumes it's in `baileys-caller/dist/index.mjs`)
- Auth dir defaults to `/root/nael-ai/.voip-auth` in the nael-ai implementation
- VoipClient instance is created lazily on first `wa_voice_call` and reused
- Socket config for `VoipClient` should fetch latest Baileys version (`fetchLatestBaileysVersion()`) and match browser platform (`["Ubuntu", "Chrome", "22.04.4"]`) to prevent 405 Method Not Allowed / Connection Failure from WhatsApp signaling servers.

## Pitfalls

- **Connection Failure** on first call usually means VoIP client auth not paired yet. **Before attempting any call**, check if auth exists:
  ```bash
  ls -la /root/nael-ai/.voip-auth/creds.json 2>/dev/null || echo "Auth missing"
  ```
  If missing or dir is empty, run init script first:
  ```bash
  cd /root/nael-ai && npx tsx scripts/init-voip.ts
  ```
  Then send generated QR (`/root/nael-ai/voip-qr.png`) to user via main bot socket using `wa_send_file`
- **Socket conflict** — do NOT try to reuse main bot socket for calls; always separate auth session
- **QR sending from unauthenticated socket** — you can't send WA messages from a socket that hasn't completed pairing. Generate QR as file first, then send via main bot or deliver manually.
- **Silent audio on call** — If target answers but hears silence despite valid MP3 source, check if the audio format/sample rate was modified or if AudioFeeder/ffmpeg pipeline in WASM engine failed to feed Float32Array PCM frames. Standard Cartesia output format is `container: "mp3", encoding: "mp3", sample_rate: 44100`.
- **`timeout_waiting_state` status** — `wa_voice_call` returns `status: "timeout_waiting_state"` when the call rings successfully but the target user does not answer before the duration timeout. `success: true` with this status indicates the VoIP call signaling successfully reached and rang the target device.
- **Relaying call protobuf messages via chat relay fails** — Injected messages like `callLogMesssage`, `scheduledCallCreationMessage`, or `bcallMessage` via `sock.relayMessage()` will render as unsupported message placeholders (*"Anda menerima pesan, tetapi versi WhatsApp Anda tidak mendukungnya"*) on client apps. WhatsApp renders call UI and joinable group banners strictly through real-time VoIP signaling stanzas (`<call>` XMPP with `<destination>` Signal encryption to each participant) and WebRTC engine state, not chat message payloads.

## Related Files

- Implementation: `src/ai/mcp-baileys.ts` (tool def + executor)
- Init script: `scripts/init-voip.ts`
- Validated pairing workflow: see `references/qr-pairing-session.md`
- Baileys Auth & Store reference: see `references/baileys-store-and-auth.md`
- nael-ai reconnection handler bug (code 500): see `references/nael-ai-reconnect-bug.md`