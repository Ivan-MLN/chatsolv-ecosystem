---
name: claude-code-session-history
description: "Read past Claude Code sessions from local JSONL logs."
version: 1.0.0
platforms: [linux, macos]
metadata:
  hermes:
    tags: [claude-code, forensics, jsonl, history, session]
    related_skills: [claude-code]
---

# Reading Claude Code session history from disk

Use when the user asks to recover/recap what a **past Claude Code session** did
(e.g. "ambilkan hasil sesi chat claude code barusan di /root", "previous session
on project X"). The `claude-code` skill is about *delegating new* work; this is
about *inspecting old transcripts* Claude already persisted.

## Where the transcripts live

`~/.claude/projects/<project-dir>/<session-uuid>.jsonl` — one directory per
working directory Claude Code was launched from (dirs are name-mangled from the
path, e.g. `-root-enslaver-v6-psycohack`). The most recent session is just the
newest `.jsonl`. There is also a `~/.claude/projects/<dir>/subagents/*.jsonl`
tree with any subagent transcripts.

## Finding the right session in 1-2 commands

```bash
# Most recent session per project dir
ls -lt ~/.claude/projects/-root-*/ 2>/dev/null | head
find ~/.claude -name "*.jsonl" -newermt "2026-08-01" 2>/dev/null | head

# Which transcripts mention a keyword (project, file, feature)
grep -rl "enslaver-v6\|<description>" ~/.claude/projects/*/*.jsonl 2>/dev/null

# Of the keyword-matching ones, order by modification time, newest first
grep -l "<description>" ~/.claude/projects/-root/*.jsonl 2>/dev/null \
  | xargs stat -c'%Y %M' | sort -rn | head
```

## Extracting a readable transcript

Each line is one JSON object `{ type, message: {...} }`. `message.content` is
either a string or an array of blocks; the block types you actually want are:
- `text` — actual user/assistant prose
- `tool_use` — a tool Claude called: `{ name, input }` (Bash commands, etc.)
- `tool_result` — its output: `content` may be a string OR array of `{text}`
- `thinking` — reasoning; usually suppress unless the user wants it

A robust one-shot dump:

```bash
python3 -c "
import json
p='<endpoint>.jsonl'  # or loop over files
for i,line in enumerate(open(p,encoding='utf-8',errors='replace')):
    try: d=json.loads(line)
    except: continue
    m=d.get('message',{})
    role=m.get('role','?')
    if role not in ('user','assistant'): continue
    c=m.get('content'); txt=''
    if isinstance(c,str): txt=c
    elif isinstance(c,list):
        for b in c:
            if isinstance(b,dict) and b.get('type')=='text': txt+=' '+b.get('text','')
    txt=' '.join(txt.split())
    if txt: print(f'[{i}][{role}]', txt[:400])
"
```

To see **tool calls + results** (what the agent actually executed/generated)
instead of prose: iterate blocks, print `tool_use` name+input and `tool_result`
content. The `tool_result.content` may be a list of `{text}` — join them.

## Pitfalls
- Skip `[0]/[1]` header lines (no `message.role`) — first real message is index 2.
- Local-command lines (e.g. `<command-name>/model</command-name>`) are user noise;
  a user often models mid-session (`/model opus` → "Set model to claude...") —
  not part of the task itself.
- A session may be **interrupted mid-task** (no final assistant answer). If the
  last user event is `tool_use` with no resolution, tell the user the analysis
  was generated but not finalized — do NOT fabricate a completion.
- Recover first, then answer from the transcript content: the record is truth,
  not your memory of what "probably" happened.
- When the user names a specific repo, grep transcripts to find sessions touching
  it, read the newest matching one, and report both the conversation and the tool
  actions (e.g. the Bash `wc -c`/`head` and the `collections.Counter` summary
  outputs) with their actual values.