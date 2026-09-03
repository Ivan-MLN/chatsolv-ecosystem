# Lumosel Proxy Reference (403 Bypass via Header Injection)

Lumosel API (`https://api.lumosel.vip/v1`) returns `403 Forbidden` if missing the `User-Agent: claude-code/1.0.53` header.

## Proxy script (`/root/proxy-lumosel/index.js` or PM2 process `proxy-lumosel`)

Runs on `http://127.0.0.1:8742`. Intercepts `/v1/*` requests and injects the header before forwarding to `https://api.lumosel.vip`.

## Test proxy direct & Key Quota Verification

```bash
# 1. Test model listing (Header verification)
curl -s -H "Authorization: Bearer lumo_live_..." http://127.0.0.1:8742/v1/models

# 2. Test chat completion (Quota verification — GET /v1/models succeeds even when quota is exhausted!)
curl -s -H "Authorization: Bearer lumo_live_..." \
  -H "Content-Type: application/json" \
  -d '{"model":"claude-sonnet-5","messages":[{"role":"user","content":"hi"}]}' \
  http://127.0.0.1:8742/v1/chat/completions
```

Expected output for completion: 200 OK with choices. If trial/quota is exhausted, it returns 429/400 `Your free trial has ended. Add balance or upgrade to continue.` despite `/v1/models` returning HTTP 200.

## 9router integration

1. Set `baseUrl` in provider data to `http://127.0.0.1:8742/v1`
2. Set `isActive = 1`
3. Clear `errorCode` in provider JSON
