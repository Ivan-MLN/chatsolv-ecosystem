# Claude Code Model ID Format Fix via 9router

## Problem

Claude Code sends model IDs with incorrect formatting when connecting through 9router:
- Uses `kiro/` prefix instead of `kr/`
- Uses underscore in version numbers: `claude-sonnet-4_5` instead of `claude-sonnet-4.5`
- Example: `kiro/claude-sonnet-4_5` → upstream Kiro API rejects with HTTP 400 `Invalid model ID`

## Root Cause

**Upstream validation happens before 9router alias resolution.** The sequence:
1. Claude Code → 9router MITM (port 443) → upstream Kiro API
2. Upstream Kiro API validates model ID format FIRST
3. If invalid format, returns HTTP 400 immediately
4. 9router never gets a chance to map aliases (combo resolution happens after upstream accepts request)

**Database aliases don't work** because they resolve AFTER upstream validation:
```sql
-- This won't help:
INSERT INTO combos (id, name, models, ...) 
VALUES (..., 'kiro/claude-sonnet-4_5', '["kr/claude-sonnet-4.5"]', ...);
```

## Solution Options

### Option 1: Fix Client Configuration (Recommended)

**File:** `/root/.claude/settings.json`

Edit the config to use correct model ID:
```json
{
  "env": {
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "kr/claude-sonnet-4.5"
  }
}
```

After editing, restart Claude Code to reload config.

**Verification:**
```bash
# Test correct format works
curl -sS "http://127.0.0.1:20128/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -d '{"model":"kr/claude-sonnet-4.5","messages":[{"role":"user","content":"test"}],"max_tokens":10,"stream":false}'
```

**Config requirements:**
- Use `kr/claude-sonnet-4.5` (dash, not underscore)
- Use `kr/` prefix, not `kiro/`

### Option 2: Request Normalization Proxy

Intercept and rewrite model IDs before they reach 9router MITM server (port 443).

**Architecture:**
```
Claude Code → Normalization Proxy (port 8443) → 9router MITM (port 443) → Upstream
```

**Implementation files** (created in `/root/.9router/runtime/mitm/`):

1. `model-normalizer.js` — Core normalization logic:
```javascript
function normalizeModelId(modelId) {
  if (!modelId || typeof modelId !== 'string') return modelId;
  
  let normalized = modelId;
  
  // kiro/* → kr/*
  if (normalized.startsWith('kiro/')) {
    normalized = 'kr/' + normalized.slice(5);
  }
  
  // underscore → dash in version numbers (4_5 → 4.5, not 4-5)
  normalized = normalized.replace(/(\d)_(\d)/g, '$1.$2');
  
  return normalized;
}

module.exports = { normalizeModelId };
```

2. `request-patcher.js` — JSON body patcher:
```javascript
const { normalizeModelId } = require('./model-normalizer');

function patchRequestBody(bodyBuffer) {
  if (!bodyBuffer || bodyBuffer.length === 0) return bodyBuffer;
  
  try {
    const bodyStr = bodyBuffer.toString('utf8');
    const body = JSON.parse(bodyStr);
    
    if (body.model) {
      const original = body.model;
      body.model = normalizeModelId(body.model);
      
      if (body.model !== original) {
        console.log(`[9router-patch] Model: ${original} → ${body.model}`);
      }
    }
    
    return Buffer.from(JSON.stringify(body), 'utf8');
  } catch (e) {
    return bodyBuffer;
  }
}

module.exports = { patchRequestBody };
```

3. `proxy-wrapper.js` — HTTP proxy server:
```javascript
#!/usr/bin/env node
const http = require('http');
const { normalizeModelId } = require('./model-normalizer');

const MITM_PORT = 443;
const WRAPPER_PORT = 8443;

function patchBody(bodyBuffer) {
  if (!bodyBuffer || bodyBuffer.length === 0) return bodyBuffer;
  
  try {
    const body = JSON.parse(bodyBuffer.toString('utf8'));
    
    if (body.model) {
      const original = body.model;
      body.model = normalizeModelId(body.model);
      
      if (body.model !== original) {
        console.log(`[wrapper] Model: ${original} → ${body.model}`);
      }
    }
    
    return Buffer.from(JSON.stringify(body), 'utf8');
  } catch (e) {
    return bodyBuffer;
  }
}

const server = http.createServer((req, res) => {
  let bodyChunks = [];
  
  req.on('data', chunk => bodyChunks.push(chunk));
  req.on('end', () => {
    const bodyBuffer = Buffer.concat(bodyChunks);
    const patchedBody = patchBody(bodyBuffer);
    
    const options = {
      hostname: '127.0.0.1',
      port: MITM_PORT,
      path: req.url,
      method: req.method,
      headers: {
        ...req.headers,
        'content-length': patchedBody.length
      }
    };
    
    const proxyReq = http.request(options, proxyRes => {
      res.writeHead(proxyRes.statusCode, proxyRes.headers);
      proxyRes.pipe(res);
    });
    
    proxyReq.on('error', err => {
      console.error(`[wrapper] Error: ${err.message}`);
      res.writeHead(502);
      res.end('Bad Gateway');
    });
    
    if (patchedBody.length > 0) {
      proxyReq.write(patchedBody);
    }
    proxyReq.end();
  });
});

server.listen(WRAPPER_PORT, () => {
  console.log(`🔧 Wrapper proxy: :${WRAPPER_PORT} → :${MITM_PORT}`);
});

server.on('error', err => {
  console.error(`[wrapper] Server error: ${err.message}`);
  process.exit(1);
});
```

**Usage:**
```bash
# Start proxy wrapper
node /root/.9router/runtime/mitm/proxy-wrapper.js

# Configure Claude Code to use port 8443 instead of 443
```

## Verification

Tested normalization logic:
```
✓ kiro/claude-sonnet-4_5 → kr/claude-sonnet-4.5
✓ kiro/claude-sonnet-4.5 → kr/claude-sonnet-4.5
✓ kr/claude-sonnet-4_5 → kr/claude-sonnet-4.5
✓ kr/claude-sonnet-4.5 → kr/claude-sonnet-4.5
✓ claude-sonnet-4_5 → claude-sonnet-4.5
```

## 9router Logs

When model ID format is wrong, 9router logs show:
```
[AUTH] Account X locked modelLock_claude-sonnet-4_5 for 120s [403]
❌ kiro [403]: [403]: HTTP 403
[FALLBACK] ⇄ ACC:Account X UNAVAILABLE (403) → NEXT ACCOUNT
[AUTH] kiro | all 5 accounts locked for claude-sonnet-4_5 (reset after 16s)
  | lastError=[400]: {"message":"Invalid model ID. Please select a different model..."}
```

The `[400]` inside `lastError` confirms upstream API rejected the format before 9router could process it.

## Key Insight

**Database alias mapping happens too late.** Upstream API validates model ID format in the HTTP request path/body before 9router's routing logic runs. To fix format issues, you must normalize the request BEFORE it reaches 9router's upstream forwarding logic — either at the client or via an intercepting proxy.
