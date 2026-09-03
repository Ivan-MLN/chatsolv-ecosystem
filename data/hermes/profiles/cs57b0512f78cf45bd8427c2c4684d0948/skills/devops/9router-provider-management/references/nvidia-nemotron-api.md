# NVIDIA Build API & Nemotron Model Testing

Reference notes for testing and validating NVIDIA Nemotron models via the OpenAI-compatible NVIDIA Integrate API endpoint, and registering them into 9router.

## Endpoint & Auth

- Base URL: `https://integrate.api.nvidia.com/v1`
- Auth Header: `Authorization: Bearer nvapi-...`

## Model Discovery & Testing Script

To list and test active Nemotron models using Python:

```python
import urllib.request, json

api_key = "nvapi-..."
models_to_test = [
    "nvidia/nemotron-3-ultra-550b-a55b",
    "nvidia/nemotron-3-super-120b-a12b",
    "nvidia/nemotron-3-nano-omni-30b-a3b-reasoning",
    "nvidia/llama-3.1-nemotron-70b-instruct"
]

url = "https://integrate.api.nvidia.com/v1/chat/completions"

for model in models_to_test:
    payload = {
        "model": model,
        "messages": [{"role": "user", "content": "Hello, reply with 1 short sentence."}],
        "max_tokens": 50
    }
    req = urllib.request.Request(
        url,
        data=json.dumps(payload).encode("utf-8"),
        headers={
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json"
        }
    )
    try:
        with urllib.request.urlopen(req, timeout=15) as resp:
            data = json.loads(resp.read().decode("utf-8"))
            msg = data["choices"][0]["message"]["content"].strip()
            print(f"✅ [SUCCESS] {model}: {msg}")
    except Exception as e:
        print(f"❌ [FAILED] {model} -> {e}")
```

## Tested Nemotron Models Status

- **Flagship / Top Tier:** `nvidia/nemotron-3-ultra-550b-a55b` (550B total / 55B active, LatentMoE, 1M context, matches/nears Claude 3.5/3.7 Sonnet & Opus on agentic SWE-Bench/Terminal Bench).
- **Mid Tier:** `nvidia/nemotron-3-super-120b-a12b` (120B total / 12B active, fast & solid reasoning).
- **Nano Reasoning:** `nvidia/nemotron-3-nano-omni-30b-a3b-reasoning` (30B total / 3B active).
- **Deprecated / Unregistered Endpoints:** Older IDs like `nvidia/llama-3.1-nemotron-70b-instruct` or `nvidia/nemotron-4-340b-instruct` return `404 Not Found` on the API integration endpoint.

## 9router Web UI Provider Registration Pattern

For standard OpenAI-compatible endpoints (like NVIDIA Integrate API) to be recognized and rendered properly in 9router's Web UI:

1. `providerConnections.provider` MUST use the format `openai-compatible-chat-<uuid>`. Custom strings like `nvidia-<uuid>` will not render in the Web UI list.
2. `providerSpecificData.prefix` MUST also be set to `openai-compatible-chat-<uuid>`.
3. Model calls via 9router should use: `openai-compatible-chat-<uuid>/nvidia/nemotron-3-ultra-550b-a55b`.
