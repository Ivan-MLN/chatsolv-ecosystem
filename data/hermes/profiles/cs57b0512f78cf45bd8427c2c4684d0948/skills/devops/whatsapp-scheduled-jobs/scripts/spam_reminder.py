#!/usr/bin/env python3
"""Spam reminder ke grup dengan tag semua member via eval_code.

Tested and working pattern for scheduled WhatsApp broadcasts with @all mentions.
Use this as a template for any cron job that needs to tag all group members.
"""

import json
import urllib.request
import sys

MCP = "http://127.0.0.1:5778/mcp"
GROUP_JID = "120363186235853203@g.us"  # Replace with target group JID

def send_spam_messages():
    """Kirim pesan spam dengan tag semua member via eval_code"""
    
    # Code runs inside the live bot process
    # sock = connected Baileys socket
    # jid = GROUP_JID passed in arguments
    code = r"""
const meta = await sock.groupMetadata(jid);
const parts = (meta.participants || []).map(p => p.id);
const mentionText = parts.map(id => `@${id.split('@')[0]}`).join(' ');
const message = `${mentionText}\n\nsudah saatnya`;

const results = [];
for (let i = 0; i < 10; i++) {
    await sock.sendMessage(jid, { text: message, mentions: parts });
    results.push(`Message ${i+1}/10 sent`);
}

return { sent: 10, mentions: parts.length, results };
"""
    
    payload = {
        "jsonrpc": "2.0",
        "id": 99,
        "method": "tools/call",
        "params": {
            "name": "eval_code",
            "arguments": {"jid": GROUP_JID, "code": code}
        }
    }
    
    req = urllib.request.Request(
        MCP,
        data=json.dumps(payload).encode(),
        headers={
            "content-type": "application/json",
            "accept": "application/json, text/event-stream"
        }
    )
    
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            result = json.loads(resp.read().decode())
            
        if result.get("error"):
            print(f"MCP Error: {result['error']}")
            return False
            
        content = result.get("result", {})
        print(f"Messages sent successfully!")
        print(f"Result: {json.dumps(content, indent=2)[:500]}")
        return True
        
    except Exception as e:
        print(f"Failed to send: {e}", file=sys.stderr)
        return False

if __name__ == "__main__":
    success = send_spam_messages()
    sys.exit(0 if success else 1)
