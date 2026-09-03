# ChatGPT-Web Proxy Integration (OpenAI-Web-Luna)

Local proxy running at `/root/proxy-chatgptweb/index.mjs` on port `8745`.

## Topology & 9router Config
- **Proxy Process**: `node /root/proxy-chatgptweb/index.mjs` (listening on `127.0.0.1:8745`)
- **Directory**: `/root/proxy-chatgptweb/` (includes helper scripts: `setup_combos.py`, `register_9router.py`, `update_prefix.py`, etc.)
- **providerNodes Entry**: `id: openai-compatible-chat-luna`, `prefix: cgweb`, `baseUrl: http://127.0.0.1:8745/v1`
- **providerConnections Entry**: `id: bc28c121-9f0a-4e48-94f3-8cee3edf40e7`, `name: OpenAI-Web-Luna`
- **Upstream Model ID**: `gpt-5.6-luna`
- **Exposed 9router Routes**: `cgweb/gpt-5.6-luna`, `gpt-5.6-luna`

## Debugging & Verification Commands
- Check active process & port listener:
  ```bash
  lsof -i :8745
  ps aux | grep proxy-chatgptweb
  ```
- Query proxy direct models endpoint:
  ```bash
  curl -s http://127.0.0.1:8745/v1/models
  ```
- Test completion via 9router:
  ```bash
  curl -s http://127.0.0.1:20128/v1/chat/completions \
    -H "Content-Type: application/json" \
    -d '{"model": "cgweb/gpt-5.6-luna", "messages": [{"role": "user", "content": "hi"}], "stream": false}'
  ```
