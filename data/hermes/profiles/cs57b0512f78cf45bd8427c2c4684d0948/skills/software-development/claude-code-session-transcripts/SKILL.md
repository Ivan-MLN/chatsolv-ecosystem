---
name: claude-code-session-transcripts
description: "Read Claude Code session history from JSONL files."
version: 1.0.0
author: Cathlyne (Hermes curator)
license: MIT
platforms: [linux, macos]
metadata:
  hermes:
    tags: [claude-code, sessions, jsonl, forensics, transcript]
    related_skills: [claude-code]
---

# Claude Code Session Transcripts

User meminta "ambilkan hasil sesi chat claude code" — tarik isi/riwayat sesi
Claude Code yang sudah lewat dari transcript JSONL di disk. Ini read-only
forensik transcript, beda dari skill `claude-code` (yang mengatur/delegasi
kerja ke CLI).

## Lokasi & format

- Session files: `~/.claude/projects/<project-dir-slug>/*.jsonl`
  (slug: `/root/enslaver-v6/...` → `-root-enslaver-v6-...`; ada juga
  `subagents/*.jsonl`).
- Satu baris JSON per pesan/event. Field penting: `message.role`
  (`user`/`assistant`), `message.content` (string ATAU array of blocks),
  block types: `text`, `tool_use`, `tool_result`, `thinking`,
  `local-command-*` (mis. `/model`).

## Menemukan sesi yang tepat

1. List by mtime terbaru dulu: `ls -lt ~/.claude/projects/<slug>/ | head`.
2. Filter yang menyebut topik: `grep -rl "<topik>" ~/.claude/projects/<slug>/*.jsonl`.
3. Cari yang paling baru: loop `grep -l` + `stat -c%y` → sort desc.
4. User bilang "barusan" = ambil mtime terbaru yang match, bukan yang dulu.

## Menampilkan isi sesi (python, satu execute_code)

```python
import json
p = '.../xxx.jsonl'
for i, line in enumerate(open(p, encoding='utf-8', errors='replace')):
    try: d = json.loads(line)
    except: continue
    m = d.get('message', {})
    role = m.get('role', '?')
    if role in ('user', 'assistant'):
        c = m.get('content'); txt = ''
        if isinstance(c, str): txt = c
        elif isinstance(c, list):
            for b in c:
                if isinstance(b, dict) and b.get('type') == 'text':
                    txt += ' ' + b.get('text', '')
        txt = ' '.join(txt.split())
        if txt: print(f'[{i}][{role}]', txt[:400])
```

Command-command lokal (`/model`, `<command-name>`) muncul sebagai pesan user
dengan prefix `<local-command-*>` — boleh dilewati saat meringkas isi.

## Menampilkan tool activity (hasil kerja, bukan teks)

Untuk lihat apa yang dieksekusi / dihasilkan:

```python
for i, line in enumerate(open(p, encoding='utf-8', errors='replace')):
    try: d = json.loads(line)
    except: continue
    m = d.get('message', {})
    for b in (m.get('content') or []):
        if isinstance(b, dict) and b.get('type') == 'tool_use':
            print(f'--- [{i}] TOOL_USE: {b.get("name")}')
            print(json.dumps(b.get('input', {}))[:400])
        elif isinstance(b, dict) and b.get('type') == 'tool_result':
            r = b.get('content')
            if isinstance(r, list): r = ' '.join(x.get('text','') for x in r if isinstance(x,dict))
            print(f'--- [{i}] TOOL_RESULT:', ' '.join(str(r).split())[:500])
```

## Pitfalls

- **Sesi yang di-interrupt tidak punya jawaban final** — cek baris terakhir:
  kalau berakhir di `tool_use` tanpa `assistant` text setelahnya, sesi belum
  kelar. Lapor ke user: analysis-nya sempat ke-generate tapi summary final
  belum ada, tawarkan selesaikan sendiri dari data yang sama.
- `message.content` bisa string (teks langsung) atau list (blocks) — selalu
  handle dua-duanya, jangan asumsi satu bentuk.
- Jangan echo token/kredensial yang muncul di transcript.
- `errors='replace'` wajib kalau file mengandung byte aneh dari output binary.