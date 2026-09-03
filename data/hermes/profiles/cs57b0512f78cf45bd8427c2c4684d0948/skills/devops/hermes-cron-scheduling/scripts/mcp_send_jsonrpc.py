#!/usr/bin/env python3
"""Send a Baileys MCP tool call from a standalone/cron script (no Hermes agent loop).

Validated by the nael-wakeup-0630 cron (@all to RecoVerse group). The MCP
(nael-ai, port 5778) answers plain JSON-RPC POSTs, so `hermes cron create
... --script this.py --no-agent` works: side effects happen here, stdout is
delivered as the cron result.

Usage: edit MCP/GROUP/code below, or import send_mcp_call() from another script.
"""
import json
import urllib.request

MCP = "http://127.0.0.1:5778/mcp"
GROUP = "120363186235853203@g.us"

# Example: @all mention. eval_code runs `new AsyncFunction("sock","jid",code)`
# inside the live bot process — sock is the connected Baileys socket.
code = r"""
const meta = await sock.groupMetadata(jid);
const parts = (meta.participants || []).map(p => p.id);
const txt = "pesan ke semua member";
await sock.sendMessage(jid, { text: txt, mentions: parts });
return { sent: true, mentions: parts.length };
"""


def send_mcp_call(name, arguments, mcp=MCP, req_id=66, timeout=30):
    payload = {
        "jsonrpc": "2.0",
        "id": req_id,
        "method": "tools/call",
        "params": {"name": name, "arguments": arguments},
    }
    req = urllib.request.Request(
        mcp,
        data=json.dumps(payload).encode(),
        headers={"content-type": "application/json",
                 "accept": "application/json, text/event-stream"},
    )
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        return resp.read().decode()


if __name__ == "__main__":
    out = send_mcp_call("eval_code", {"jid": GROUP, "code": code})
    print(out[:600])
