---
name: whatsapp-voice-replies
description: "Use when a Baileys bot needs TTS voice-note replies."
version: 1.0.0
platforms: [linux]
metadata:
  hermes:
    tags: [whatsapp, baileys, tts, cartesia, voice, nael-ai]
---

# Adding AI voice replies to a WhatsApp bot (voice notes)

When a bot should answer with BOTH a text reply AND a WhatsApp voice note
(`ptt:true`) spoken by an AI voice. Verified working in nael-ai
(`src/voice/tts.ts`, session Aug 2026).

## Disabling / Controlling Cartesia TTS
- Set `ENABLE_CARTESIA=false` in `.env` and comment out `#CARTESIA_API_KEY=...`.
- In `src/voice/tts.ts`, guard key check with `process.env.ENABLE_CARTESIA === "true" ? process.env.CARTESIA_API_KEY : undefined`.
- When key is undefined, `speak()` logs `"Cartesia not configured"` and returns `null` safely without attempting network requests.

## Key facts about Cartesia (api.cartesia.ai)
- **Cartesia does NOT output ogg/opus.** Asking for `container:"ogg"` returns
  HTTP 400. You MUST request **mp3** (`{container:"mp3",encoding:"mp3",
  sample_rate:44100}`) then re-encode locally.
- Auth: header `Cartesia-Version: 2024-06-10` + `X-API-Key`. Model `sonic-3.5;
  voice `voice:{mode:"id",id:"<voiceId>"}`.
- A 401 = wrong API key; 400 on a valid key usually means unsupported output_format.
- Read the key from `.env` via `process.env`, never hardcode/echo. Confirm it
  exists without leaking: `assert "$KEY" == "$(grep -o ..."=)" .env)"` style.

## The  voice-note shape (Baileys)
```
sock.sendMessage(jid, { audio:oggBuffer, ptt:true, mimetype:"audio/ogg;codecs=opus", seconds:1 }, {quoted:ctx.msg});
```
`ptt?:boolean` exists in Baileys Message.d.ts — the voice-note flag.

## Re-encode mp3→Ogg Opus (ffmpeg)
- `execFile`/`execFileP` have NO `input` option — use `spawn(...,{stdio:["pipe","pipe","ignore"]})`,
  collect `child.stdout.on("data")` chunks, `child.stdin.end(mp3)`, await `close`, concat, check exit.
- Command: `ffmpeg -y -loglevel error -i pipe:0 -c:a libopus -b:a 48k -ar 48000 -f ogg pipe:1`

## Normalize text before TTS
Strip markdown `#*`[]{}``, emoji/pictographs, `|...|`, URLs. Then collapse
`?!?{2,}`→`?`, clock `"07:30"`→`"pukul 7 lewat 30 menit"`, `a/b`→`a atau b`,
em-dash→comma, newline→pause, caps→lowercase unless KEEP_SPELLED, slice to cap.

## Integration
- One module does normalize+fetch+convert+send.
- Call AFTER `sock.replyAi`, in the same try, guarded by a length cap (voice only
  SHORT replies e.g. ≤180 chars) and `speak()` returning null when key unset.

## Pitfalls
- Cartesia sonic-3.5 SPEAKS, never sings — no melody. Real singing needs Suno
  (free 50 credits/day) or a GPU model. Be honest; don't fake singing.
- Only voice short texts — quota + essay-as-speech.
- A protected `.env` rejects `patch` → append via terminal sed/heredoc.

## Reference
- `references/cartesia-details.md` — exact proven request body, the ogg=400
  quirk, verified byte sizes, and the `.env` key-handling discipline.