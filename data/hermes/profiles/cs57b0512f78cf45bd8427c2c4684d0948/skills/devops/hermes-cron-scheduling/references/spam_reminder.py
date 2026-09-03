#!/usr/bin/env python3
"""
Spam reminder to WhatsApp group with @mentions.

Demonstrates:
- Calling Baileys MCP tools (wa_group_metadata, wa_send_message) from a no-agent cron script
- Building mentions array from participants (using @lid format, not @s.whatsapp.net)
- Loop with error handling for burst sends
"""

import json
import requests
import sys

MCP_URL = "http://127.0.0.1:5778/mcp/v1"
GROUP_JID = "120363186235853203@g.us"

def get_group_metadata():
    """Fetch group metadata to get participant list."""
    resp = requests.post(
        f"{MCP_URL}/call-tool",
        json={
            "name": "wa_group_metadata",
            "arguments": {"jid": GROUP_JID}
        },
        timeout=30
    )
    resp.raise_for_status()
    result = resp.json()
    
    if result.get("isError"):
        raise Exception(f"Error getting metadata: {result.get('content')}")
    
    content = result.get("content", [{}])[0].get("text", "{}")
    return json.loads(content)

def send_spam_messages():
    """Send 10 spam messages with @all mentions."""
    metadata = get_group_metadata()
    participants = metadata.get("participants", [])
    
    # Build mentions list (format: @lid, not @s.whatsapp.net)
    mentions = [p["id"] for p in participants]
    
    # Build message text with all mentions
    mention_text = " ".join([f"@{m.split('@')[0]}" for m in mentions])
    message = f"{mention_text}\n\nsudah saatnya"
    
    print(f"Sending 10 spam messages to group with {len(mentions)} mentions...")
    
    for i in range(10):
        resp = requests.post(
            f"{MCP_URL}/call-tool",
            json={
                "name": "wa_send_message",
                "arguments": {
                    "jid": GROUP_JID,
                    "text": message,
                    "mentions": mentions
                }
            },
            timeout=30
        )
        resp.raise_for_status()
        result = resp.json()
        
        if result.get("isError"):
            print(f"Message {i+1}/10 FAILED: {result.get('content')}")
        else:
            print(f"Message {i+1}/10 sent ✓")
    
    print("Done!")

if __name__ == "__main__":
    try:
        send_spam_messages()
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)
