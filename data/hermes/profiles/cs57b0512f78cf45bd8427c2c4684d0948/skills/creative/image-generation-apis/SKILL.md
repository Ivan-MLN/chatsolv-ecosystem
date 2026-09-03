---
name: image-generation-apis
description: Use when generating images via cloud APIs without local GPU.
version: 1.0.0
author: curator
platforms: [linux, macos, windows]
prerequisites:
  commands: ["curl"]
metadata:
  hermes:
    tags: [image-generation, api, gemini, imagen, stability-ai, replicate, fal-ai, cloud]
    category: creative
---

# Image Generation APIs

Generate images via cloud APIs (Gemini Imagen, Stability AI, Replicate, fal.ai, ComfyUI Cloud) when local GPU compute isn't available or practical. Covers setup, pricing, quota limits, and provider selection.

## When to Use

- User wants image generation but has no capable GPU
- Prototyping / low-volume use cases
- Already using a provider's ecosystem (e.g., Gemini for text → add image generation)
- Scaling beyond local hardware capacity
- Need models not available locally (proprietary, cutting-edge)

## Provider Comparison

| Provider | Setup | Free Tier | Paid Pricing | Quota/Rate | Models | Advanced Features |
|----------|-------|-----------|--------------|------------|--------|-------------------|
| **Gemini Imagen** | API key | Very limited daily quota | Unknown | Low (free), unknown (paid) | Imagen 3 | Prompt only |
| **Stability AI** | API key | 25 credits (~3 imgs) | $0.002–0.04/img | 500 req/min (paid) | SDXL, SD3, Ultra | img2img, upscale, edit |
| **Replicate** | API key | Pay-per-use only | $0.001–0.05/img | Model-dependent | 100+ (community) | Full model zoo |
| **fal.ai** | API key | Pay-per-use only | $0.001–0.03/img | Fast, pay-as-go | Flux, SDXL, SD3 | WebSocket, fast inference |
| **ComfyUI Cloud** | API key | 1 job (read-only free) | ~$0.00025/img | 1/3/5 concurrent | SD1.5, SDXL, Flux | Full workflow control |

**For local generation (zero API cost, unlimited), see the `comfyui` skill.**

## Gemini Image Generation (Imagen 3 & Antigravity 9router)

Google's Gemini API supports native image generation and image-to-image editing. On 9router / local proxy setups with Antigravity (`ag/`), image generation and img2img (editing existing images via multimodal base64 prompt) are available via `ag/gemini-3.1-flash-image`.

### Setup & 9router Antigravity Endpoint

```bash
# Direct endpoint via 9router OpenAI-compatible API
API_URL="http://127.0.0.1:20128/v1/chat/completions"
# Model: ag/gemini-3.1-flash-image
```

### Image-to-Image / Style Transfer via Multimodal Base64 (Python)

To modify or restyle a user's image (e.g. changing hairstyle, clothing, or scene) while preserving identity:

```python
import requests, base64, re

with open("input.jpg", "rb") as f:
    b64_img = base64.b64encode(f.read()).decode("utf-8")

payload = {
    "model": "ag/gemini-3.1-flash-image",
    "messages": [
        {
            "role": "user",
            "content": [
                {"type": "text", "text": "Based on this photo, generate a realistic photo with [desired style/haircut/changes]. Keep the facial features and identity consistent."},
                {"type": "image_url", "image_url": {"url": f"data:image/jpeg;base64,{b64_img}"}}
            ]
        }
    ],
    "max_tokens": 1000,
    "stream": False
}

headers = {
    "Content-Type": "application/json",
    "Authorization": "Bearer YOUR_9ROUTER_KEY"
}

res = requests.post("http://127.0.0.1:20128/v1/chat/completions", json=payload, headers=headers)
content = res.json()["choices"][0]["message"]["content"]

# Response returns Markdown image format: ![image](data:image/jpeg;base64,<data>)
match = re.search(r'!\[image\]\(data:image/jpeg;base64,([^)]+)\)', content)
if match:
    b64_out = match.group(1)
    # Pad base64 if needed (length % 4)
    missing_padding = len(b64_out) % 4
    if missing_padding:
        b64_out += '=' * (4 - missing_padding)
    with open("output.jpg", "wb") as f:
        f.write(base64.b64decode(b64_out))
```

### Direct Google API Generate Image

```bash
curl -s "https://generativelanguage.googleapis.com/v1/models/gemini-2.5-flash-image:generateContent" \
  -H "x-goog-api-key: $GEMINI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "contents": [{
      "parts": [{"text": "A majestic dragon flying over snowy mountains at sunset, ultra realistic, cinematic lighting, 8k"}]
    }],
    "generationConfig": {
      "responseModalities": ["IMAGE"]
    }
  }' | jq -r '.candidates[0].content.parts[0].inlineData.data' | base64 -d > output.png
```

### Quota Limits

**Free tier is VERY restrictive:**
- Daily quota per model (resets ~24h after first use)
- Per-minute request limits
- Input token count limits

Error 429 with `RESOURCE_EXHAUSTED` means quota exceeded. Response includes `retryDelay` (typically 3–48 seconds for rate limit, but daily quota requires waiting until next day).

**For production/monetization, paid tier required** — but pricing is not yet publicly documented.

### When to Use Gemini

- Zero-install prototyping
- Low-volume (<10 images/day on free tier)
- Already using Gemini API for text (same billing)

### When NOT to Use

- High-volume generation (quota exhausts fast)
- Need img2img, inpainting, upscaling (not supported)
- Need specific models (Flux, ControlNet, LoRA)
- Free tier requirement for production

## Stability AI

**Docs:** https://platform.stability.ai/docs

### Setup

```bash
export STABILITY_API_KEY="sk-..."
```

### Generate (SDXL)

```bash
curl -X POST "https://api.stability.ai/v2beta/stable-image/generate/sd3" \
  -H "Authorization: Bearer $STABILITY_API_KEY" \
  -H "Accept: image/*" \
  -F "prompt=A majestic dragon" \
  -F "output_format=png" \
  --output dragon.png
```

### Pricing

- **Free:** 25 credits (~3 SDXL images)
- **Paid:** $10/month (500 credits) → Professional plans for scale
- SDXL: ~7 credits/image, SD3: ~6.5 credits, Ultra: ~8 credits

### When to Use

- Need img2img, upscaling, inpainting
- Reliable SLA for production
- Stable Diffusion ecosystem familiarity

## Replicate

**Docs:** https://replicate.com/docs

Community-driven model zoo with 100+ image generation models.

### Setup

```bash
export REPLICATE_API_TOKEN="r8_..."
pip install replicate  # optional Python SDK
```

### Generate (Flux Dev)

```bash
curl -X POST "https://api.replicate.com/v1/predictions" \
  -H "Authorization: Bearer $REPLICATE_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "version": "black-forest-labs/flux-dev",
    "input": {"prompt": "A majestic dragon"}
  }'
# Returns prediction URL; poll GET /v1/predictions/{id} for status
```

### Pricing

Pay-per-use only, varies by model:
- Flux Dev: ~$0.003/image
- SDXL: ~$0.001/image
- Custom fine-tunes available

### When to Use

- Need cutting-edge / community models
- Experimenting with multiple models
- Python SDK integration preferred

## fal.ai

**Docs:** https://fal.ai/docs

Fast inference platform focused on speed.

### Setup

```bash
export FAL_KEY="..."
pip install fal-client  # Python SDK
```

### Generate (Flux Schnell)

```bash
curl -X POST "https://fal.run/fal-ai/flux/schnell" \
  -H "Authorization: Key $FAL_KEY" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "A majestic dragon"}'
```

### Pricing

Pay-as-you-go, model-dependent:
- Flux Schnell: ~$0.003/image
- SDXL: ~$0.001/image
- WebSocket support for streaming progress

### When to Use

- Speed-critical applications
- Real-time generation UI
- WebSocket progress monitoring preferred

## ComfyUI Cloud

See the `comfyui` skill for full docs.

**Quick comparison:**
- **Free tier:** Read-only (cannot run workflows via API)
- **Paid:** ~$0.00025/image estimate, 1/3/5 concurrent jobs depending on plan
- **Best for:** Users already familiar with ComfyUI workflows who want hosted GPUs

## Landing Page & UI VideoGen Sequence Pipeline

For scroll-driven Canvas animations (e.g., product/smartphone 3D rotation scrub):
- **Never generate 120 separate AI images** (causes geometry mutation and visual jitter).
- **Correct Pipeline:**
  1. Generate 1 high-fidelity anchor hero image (ImageGen).
  2. Feed anchor into VideoGen (e.g., Runway/Kling/Luma/Sora) for a **6-second continuous master video** (e.g., 0-1s static, 1-2.2s push-in, 2.2-5.2s rotation, 5.2-6s settle).
  3. Extract frames with ffmpeg: `ffmpeg -i master.mp4 -vf "fps=20,scale=1920:1080" -vcodec libwebp -q:v 85 frame-%04d.webp` (exactly 120 frames @ 20fps for 6s).
  4. Feed extracted WebP frames into Canvas scroll-scrub engine with progressive loading.

## Asset Request Protocol (When Assets are Missing)

When building UI and production visuals are not yet rendered, agents must log structured requests into `asset-requests.md` with:
- **Asset ID & Section**
- **Exact resolution, aspect ratio & format** (WebP/PNG/MP4)
- **Transparency requirement**
- **Generation type** (`ImageGen` / `VideoGen` / `Manual`)
- **Master Prompt & Negative Constraints**
- **Implementation / Destination Path**

## Decision Tree

| User Says | Recommended Provider | Why |
|-----------|---------------------|-----|
| "I need to prototype quickly, no GPU" | **Gemini** (if <10 imgs) or **fal.ai** | Fastest setup |
| "I need 100+ images/day" | **ComfyUI local** (if GPU) or **Stability AI** / **Replicate** | Cost per image matters |
| "I need img2img / inpainting / upscaling" | **Stability AI** or **ComfyUI** | Gemini/Replicate basic models lack these |
| "I want the newest community models" | **Replicate** | Largest model zoo |
| "I need sub-3s generation" | **fal.ai** | Optimized for speed |
| "I already use Gemini for text" | **Gemini Imagen** | Same billing, low volume OK |
| "I want full workflow control (ControlNet, LoRA)" | **ComfyUI** (local or cloud) | Only ComfyUI supports arbitrary workflows |

## Pitfalls

1. **Free tier quota exhaustion** — Gemini free tier is NOT suitable for any production use. Test with 1–2 images, then switch to paid or local.

2. **Base64 inline data size** — Gemini returns images as base64 in JSON responses. For large images or batch generation, this inflates network transfer. Other APIs return direct image bytes or URLs.

3. **Model availability** — Gemini only supports Imagen 3 (no Stable Diffusion, Flux, etc.). If user asks for a specific model, route to Replicate or ComfyUI.

4. **Retry logic required** — All APIs can return 429 (rate limit) or 5xx (transient failure). Implement exponential backoff for production.

5. **Pricing changes** — Gemini image generation pricing is not yet public. Monitor https://ai.google.dev/pricing for updates.

6. **No local fallback** — API-only providers cannot work offline. For airgapped/offline environments, use ComfyUI local.

## Verification

- [ ] API key set in environment
- [ ] `curl` test returns image data (not 401/403)
- [ ] Quota limits understood (free tier daily caps)
- [ ] Pricing confirmed for expected volume
- [ ] Output format tested (base64 decode, binary download, URL fetch)
