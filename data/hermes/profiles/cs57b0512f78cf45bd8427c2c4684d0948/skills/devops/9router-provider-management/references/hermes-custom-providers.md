# Adding Custom Providers to Hermes (Not 9router)

Complete workflow for registering external OpenAI-compatible endpoints as **Hermes providers** with fallback support. This is the correct approach when you want to add an external router/proxy as a Hermes fallback without modifying 9router's codebase.

## Use Case

- You have an external OpenAI-compatible endpoint (custom router, proxy gateway, third-party service)
- You want it available as a Hermes fallback when primary models fail
- You don't need it exposed through 9router's `/v1/models` endpoint

## Workflow

### 1. Register Provider via `hermes config set`

**Never hand-edit `config.yaml` for provider registration** — use the CLI:

```bash
# Set provider type (openai for OpenAI-compatible endpoints)
hermes config set providers.<name>.type openai

# Set base URL (without /chat/completions suffix)
hermes config set providers.<name>.base_url https://your-endpoint.com/api/v1

# Set API key environment variable name
hermes config set providers.<name>.api_key_env YOUR_PROVIDER_API_KEY
```

**Example:**
```bash
hermes config set providers.lapakvip.type openai
hermes config set providers.lapakvip.base_url https://router.lapakvip.com/api/v1
hermes config set providers.lapakvip.api_key_env LAPAKVIP_API_KEY
```

### 2. Set API Key Environment Variable

```bash
# Temporary (current session)
export YOUR_PROVIDER_API_KEY="your-api-key-here"

# Persistent (add to shell rc file)
echo 'export YOUR_PROVIDER_API_KEY="your-api-key-here"' >> ~/.bashrc
source ~/.bashrc
```

**Security:** API keys belong in environment variables or `~/.hermes/.env`, NEVER in `config.yaml`.

### 3. Add to Fallback Chain

Fallback arrays require manual `config.yaml` editing (sed or text editor) because `hermes config set` doesn't support array append:

```bash
# Option A: sed (insert before "provider: custom" line in fallbacks section)
sed -i '/^  fallbacks:/,/^  provider: custom/ {
  /^  provider: custom/i\    - model: your-model-name\n      provider: your-provider-name
}' ~/.hermes/config.yaml

# Option B: manual edit (insert at correct indentation)
# Open ~/.hermes/config.yaml and add under model.fallbacks:
#   - model: claude-sonnet-4.5-testing
#     provider: lapakvip
```

### 4. Verify Configuration

```bash
# Check provider registered
hermes config get providers.<name>

# Check fallback chain
hermes config get model.fallbacks

# Test endpoint directly
curl -X POST https://your-endpoint.com/api/v1/chat/completions \
  -H "Authorization: Bearer $YOUR_PROVIDER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "model-name",
    "messages": [{"role": "user", "content": "hi"}],
    "max_tokens": 10
  }'
```

### 5. Triggering Fallback

Hermes fallback activates on:
- 429 (rate limit)
- 503 (service error)
- 529 (overload)
- Connection failures

Check `~/.hermes/logs/gateway.log` for fallback routing decisions.

## Complete Example: LakapVIP Router

```bash
# 1. Register
hermes config set providers.lapakvip.type openai
hermes config set providers.lapakvip.base_url https://router.lapakvip.com/api/v1
hermes config set providers.lapakvip.api_key_env LAPAKVIP_API_KEY

# 2. Set key
export LAPAKVIP_API_KEY="lv-__your_key_here"
echo 'export LAPAKVIP_API_KEY="lv-__your_key_here"' >> ~/.bashrc

# 3. Add to fallback (sed approach)
sed -i '/^  fallbacks:/,/^  provider: custom/ {
  /^  provider: custom/i\    - model: claude-sonnet-4.5-testing\n      provider: lapakvip
}' ~/.hermes/config.yaml

# 4. Verify
hermes config get providers.lapakvip
hermes config get model.fallbacks
curl -X POST https://router.lapakvip.com/api/v1/chat/completions \
  -H "Authorization: Bearer $LAPAKVIP_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model": "claude-sonnet-4.5-testing", "messages": [{"role":"user","content":"hi"}], "max_tokens": 10}'
```

## Pitfalls

- **Model name prefix confusion**: If the provider uses prefixes (e.g. `lv/model-name`), check whether the prefix is required at the provider API level. Test with `curl` first. Do NOT include provider prefixes in the `model:` field inside `config.yaml` fallbacks unless the upstream API requires it.
- **base_url trailing slash**: Some endpoints require no trailing slash (`/api/v1` not `/api/v1/`). Test with `curl` if 404 errors occur.
- **Fallback not triggering**: Normal errors (invalid model, bad request) don't trigger fallback. Only rate limits, service errors, and connection failures do.
- **Model availability**: Not all models from external providers auto-populate Hermes' model catalog. Fallback models must be explicitly named in `config.yaml`.

## Benchmarking

To compare latency/TPS between custom providers and 9router local models, see `scripts/benchmark_llm.py` in this skill directory.

## Related

- Main skill: `9router-provider-management` — SQLite provider injection (different approach for 9router itself)
- `references/benchmark-llm-endpoints.md` — Python benchmark script reference
