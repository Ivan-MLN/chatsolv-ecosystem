---
name: temporary-email-automation
description: "Use when generating temp mail & polling OTPs via qemail."
version: 1.0.0
platforms: [linux]
metadata:
  hermes:
    tags: [email, tempmail, qemail, otp, verification, polling, whatsapp]
---

# Temporary Email Automation (qemail.web.id API)

Use when creating disposable email addresses, monitoring their inbox for incoming verification/OTP emails (e.g. SpaceXAI, X.AI, etc.), and automatically extracting OTPs or forwarding alerts (e.g. to WhatsApp).

## 1. Generate Disposable Email Address

Endpoint: `POST https://api.qemail.web.id/v1/email/generate`

### Request Headers
```json
{
  "User-Agent": "Mozilla/5.0 (Linux; Android 15; Pixel 9) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/151.0.0.0 Mobile Safari/537.36",
  "Content-Type": "application/json",
  "Referer": "https://qemail.web.id/"
}
```

### Request Body
```json
{
  "domain_id": 15,
  "username": "naelgrok<random_string>",
  "password": "<random_16_chars>",
  "is_custom": false,
  "forward_to": "aelyn4k@gmail.com"
}
```

### Response Example (Status 201)
```json
{
  "email": "naelgrokum14k4@aii.my.id",
  "session_token": "nG7LO9fllwCM8v0QPSDjNa8bzpWnMPp0q2rSWkHwLTNm9Lhy",
  "token": "eyJhbG...",
  "expires_at": "2027-02-17T00:54:53.131Z"
}
```

---

## 2. Inbox Polling & OTP Extraction

Endpoint: `GET https://api.qemail.web.id/v1/email/inbox/{session_token}?page=1&limit=20`

### Polling Pattern (Python 5-second loop)
- Poll every **5 seconds** for fast real-time OTP delivery.
- Execute HTTP requests using `urllib.request`.
- Parse `body_html` or `subject` for OTP codes using regex.

```python
import urllib.request
import json
import re
import time

TARGET_JID = "120363186235853203@g.us"
BAILEYS_URL = "http://127.0.0.1:5778/send-message"

def send_wa(msg):
    try:
        payload = json.dumps({
            "jsonrpc": "2.0",
            "id": int(time.time() * 1000),
            "method": "tools/call",
            "params": {
                "name": "wa_send_message",
                "arguments": {
                    "jid": TARGET_JID,
                    "text": msg
                }
            }
        }).encode('utf-8')
        req = urllib.request.Request(BAILEYS_URL, data=payload, headers={"Content-Type": "application/json"}, method='POST')
        with urllib.request.urlopen(req) as resp:
            pass
    except Exception as e:
        print(f"Error sending WA: {e}")

session_token = "<session_token>"
inbox_url = f"https://api.qemail.web.id/v1/email/inbox/{session_token}?page=1&limit=20"
headers = {
    "User-Agent": "Mozilla/5.0 (Linux; Android 15; Pixel 9) AppleWebKit/537.36",
    "Referer": "https://qemail.web.id/"
}

# Poll loop (5-second sleep at end of loop)
for i in range(120): # 10 minutes total max
    try:
        req = urllib.request.Request(inbox_url, headers=headers)
        with urllib.request.urlopen(req) as resp:
            inbox_data = json.loads(resp.read().decode('utf-8'))
            for msg in inbox_data.get("data", []):
                html = msg.get("body_html", "")
                # Common OTP Patterns: 215-108 or 6-digit codes
                otp_match = re.search(r'(\d{3}-\d{3})', html) or re.search(r'\b(\d{6})\b', html)
                if otp_match:
                    otp_code = otp_match.group(1)
                    send_wa(f"*OTP RECEIVED:* `{otp_code}`")
                    break
    except Exception as e:
        print("Error polling:", e)
    time.sleep(5)
```

---

## Pitfalls & Best Practices
- **Session Persistence & Cookie Export:** Puppeteer by default launches with an isolated temporary profile; all cookies and sessions are lost when `browser.close()` is called. To preserve the logged-in session so the user can open it directly in their desktop browser:
  - Specify `userDataDir: 'C:\\Users\\ThinkPad\\ChromeProfileXAI'` (or another dedicated persistent directory) during `puppeteer.launch()`.
  - Alternatively, extract all cookies via CDP (`const { cookies } = await page.target().createCDPSession().send('Network.getAllCookies');`) and save to `cookies.json`.
  - To auto-login the user's browser, run a Python script with Selenium: load domain, inject cookies via `driver.add_cookie()`, and refresh (`driver.get(url)`).
- **Cloudflare Protection, CDP Automation Flags & Turnstile Limitations:**
  - Standard HTTP requests to `api.qemail.web.id/v1/email/generate` require a full desktop `User-Agent` header (e.g. `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36`), otherwise Cloudflare/WAF returns `HTTP 403 Forbidden`.
  - Modern auth forms (e.g. x.ai / SpaceX AI) often use Cloudflare Turnstile protection. Standard Selenium / automated browser requests trigger Cloudflare 403 challenges, returning `Unexpected token '<', "<!DOCTYPE "... is not valid JSON` when expecting JSON. Use `puppeteer-extra-plugin-stealth` in Puppeteer (or `undetected-chromedriver` with matched browser version) to bypass initial bot detection.
  - **CDP Automation Detection Limits:** Connecting Puppeteer/Selenium to a running Chrome instance via `--remote-debugging-port=9222` or `puppeteer.connect()` still injects Chrome DevTools Protocol (CDP) bindings into the browser process. Even if the user manually clicks Cloudflare Turnstile checkboxes in an attached browser window, Turnstile detects the active CDP connection/automation flags and rejects verification with `Verification failed. Please refresh the page and try again.` Explain to the user that Cloudflare Turnstile actively detects automation protocol listeners (CDP/webdriver), so automated form submission + manual captcha clicking on Turnstile-protected pages will fail unless using official API keys or un-automated manual browser sessions.
- **Retrieving Remote Headless Browser Screenshots over SSH (Windows to Linux VPS):** When capturing screenshots from Puppeteer running on a remote Windows machine (e.g. ThinkPad) over SSH, direct `scp thinkpad:...` may fail due to hostname resolution. Use base64 encoding over `ssh`:
  ```python
  import subprocess, base64
  cmd = ['ssh', '-i', '/root/.ssh/thinkpad_key', 'thinkpad@100.77.36.29', 'powershell -Command "[Convert]::ToBase64String([IO.File]::ReadAllBytes(\'C:\\\\Users\\\\ThinkPad\\\\puppeteer_live.png\'))"']
  out = subprocess.check_output(cmd).decode().strip()
  with open('/tmp/puppeteer_live.png', 'wb') as f:
      f.write(base64.b64decode(out))
  ```
- **Live Screenshot Streaming for Headless Execution:** When the user requests to monitor headless browser execution live on their desktop without opening an invasive GUI browser window, periodically capture full-page/viewport screenshots to a fixed file on disk (e.g. `C:\Users\ThinkPad\puppeteer_live.png`) and append step timestamps to a log file (`puppeteer_live.log`). This enables non-intrusive live visual monitoring.
- **Exact OTP Regex Extraction:** Always extract the exact real-time OTP code (including dashed formats like `839-556` -> `839556`) from `subject`, `body_html`, or `body_text`. Never inject mock fallback OTPs (e.g. `333333` or `888888`), as entering invalid OTPs causes verification failure on the site.
- **Node.js `https.request` Timeout & Fallback Handling:** Always specify `timeout: 5000` on `https.request` / `https.get` calls in Node.js and attach a `req.on('timeout', ...)` handler that calls `req.destroy()` and resolves a fallback or error. Without explicit timeout destruction, pending network calls to temp mail APIs will hang the Node.js event loop indefinitely.
- **Windows PowerShell 5.1 CLI Syntax Pitfalls:**
  - PowerShell 5.1 does not support `&&` for chaining commands (`cd C:\dir && node script.js` fails with `ParserError: The token '&&' is not a valid statement separator`). Use `;` instead (`cd C:\dir; node script.js`).
  - Executing commands with flag arguments starting with `--` (e.g. `chrome.exe --remote-debugging-port=9222`) in PowerShell throws unexpected token errors unless preceded by the call operator `&` (`& "C:\...\chrome.exe" --remote-debugging-port=9222`) or run via `Start-Process`.
- **Chrome Remote Debugging Port (9222) Port Locks:** When connecting Puppeteer to an existing Chrome via `puppeteer.connect({ browserURL: 'http://127.0.0.1:9222' })`, if a prior script crashed or hung without calling `browser.disconnect()`, HTTP requests to `127.0.0.1:9222/json/version` will hang indefinitely. Clear stale Chrome/Node processes (`Stop-Process -Name chrome -Force` or kill the PID) before re-launching debug Chrome.
- **Multi-Step Profile Completion & OTP Form Input:** Check if the form requires single-digit input array (`input[data-input-otp="true"]` or 6 separate inputs) vs single input box (`input[name="code"]`). Post-OTP verification, forms usually proceed to profile creation requiring randomized `firstName`, `lastName`, and `password` before final submission.
- **Interactive Windows Sessions vs Session 0 (Headless vs Visible GUI):** Executing GUI browser scripts via SSH on Windows runs in Session 0 (hidden background service session), so no browser window appears on the user's monitor. When the user asks for a visible window on ThinkPad, call the remote listener endpoint (`curl "http://100.77.36.29:19999/open?url=<target_url>"`) to open Chrome in the user's active GUI session, or explain Session 0 isolation and send real-time progress screenshots to WhatsApp via base64 retrieval over SSH.
- **Node.js ES Modules & Shell Quotation:** When deploying `.mjs` scripts or inline scripts via SSH wrappers, inline quotes and character limits can truncate command execution (`The command line is too long.`). Split payloads into base64 chunks or write the `.mjs` file cleanly to disk first before executing `node <file>.mjs`.
- **Baileys MCP Protocol:** Port 5778 runs a JSON-RPC HTTP server (`tools/call`), NOT a plain REST endpoint (`/send-message`). Sending raw REST JSON payloads directly to `http://127.0.0.1:5778/send-message` or without JSON-RPC structure (`{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "wa_send_message", "arguments": {"jid": "...", "text": "..."}}}`) will fail silently or be ignored.
- **Spam / Real-time Status Updates:** User prefers immediate feedback on every polling check (e.g. sending `[Check ke-N]` status updates per 5 seconds) until OTP is found, confirming the watcher script is actively running and fetching.
- **Fast Polling Interval:** Default to 5 seconds (`time.sleep(5)`) placed at the *end* of the loop, so initial checks and notifications fire immediately without artificial initial delays.
- **No external dependencies required:** Use `urllib.request` and `json` from standard library to ensure reliability in sandbox execution.
- **Session Token:** Always store `session_token` returned from email generation; it is required for querying inbox.
- **Background Execution:** For long-running inbox polling (e.g. up to 10 minutes), write the Python script to a file (e.g., `/tmp/temp_mail_watcher.py`) and launch with `terminal(background=True, notify_on_complete=True)`.
