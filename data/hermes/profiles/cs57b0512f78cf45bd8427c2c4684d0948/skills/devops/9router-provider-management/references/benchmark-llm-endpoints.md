# LLM Endpoint Benchmarking

Python script for comparing OpenAI-compatible API endpoints (latency, TTFT, TPS).

## Script: `benchmark_llm.py`

```python
#!/usr/bin/env python3
"""
Benchmark LLM endpoints: latency, TPS, TTFT
"""
import time
import json
import requests
from statistics import mean

CONFIGS = [
    {
        "name": "9router Local",
        "base_url": "http://127.0.0.1:20128/v1",
        "api_key": "dummy",
        "model": "kr/claude-sonnet-4.5"
    },
    {
        "name": "External Provider",
        "base_url": "https://api.example.com/v1",
        "api_key": "your-key-here",
        "model": "claude-sonnet-4.5"
    }
]

TEST_PROMPT = "Explain quantum computing in exactly 100 words."
ITERATIONS = 5

def benchmark_endpoint(config):
    results = {
        "name": config["name"],
        "latencies": [],
        "ttfts": [],
        "tokens": [],
        "errors": 0
    }
    
    print(f"\n{'='*60}")
    print(f"Testing: {config['name']}")
    print(f"Model: {config['model']}")
    print(f"{'='*60}")
    
    for i in range(ITERATIONS):
        print(f"  Run {i+1}/{ITERATIONS}...", end=" ", flush=True)
        
        try:
            start = time.time()
            ttft = None
            output_tokens = 0
            
            response = requests.post(
                f"{config['base_url']}/chat/completions",
                headers={
                    "Authorization": f"Bearer {config['api_key']}",
                    "Content-Type": "application/json"
                },
                json={
                    "model": config["model"],
                    "messages": [{"role": "user", "content": TEST_PROMPT}],
                    "stream": True,
                    "max_tokens": 150
                },
                stream=True,
                timeout=60
            )
            
            if response.status_code != 200:
                print(f"❌ HTTP {response.status_code}")
                results["errors"] += 1
                continue
            
            for line in response.iter_lines():
                if not line:
                    continue
                    
                line_str = line.decode('utf-8')
                if not line_str.startswith('data: '):
                    continue
                    
                data_str = line_str[6:]
                if data_str == '[DONE]':
                    break
                
                try:
                    chunk = json.loads(data_str)
                    if ttft is None:
                        ttft = time.time() - start
                    
                    if 'choices' in chunk and len(chunk['choices']) > 0:
                        delta = chunk['choices'][0].get('delta', {})
                        if 'content' in delta:
                            output_tokens += 1
                except json.JSONDecodeError:
                    continue
            
            end = time.time()
            latency = end - start
            
            results["latencies"].append(latency)
            results["ttfts"].append(ttft if ttft else latency)
            results["tokens"].append(output_tokens)
            
            tps = output_tokens / latency if latency > 0 else 0
            print(f"✓ {latency:.2f}s | TTFT: {ttft:.3f}s | {output_tokens} tok | {tps:.1f} tok/s")
            
        except Exception as e:
            print(f"❌ {str(e)[:50]}")
            results["errors"] += 1
        
        time.sleep(1)
    
    return results

def print_summary(all_results):
    print(f"\n{'='*80}")
    print("BENCHMARK SUMMARY")
    print(f"{'='*80}\n")
    
    for results in all_results:
        if len(results["latencies"]) == 0:
            continue
            
        name = results["name"]
        avg_latency = mean(results["latencies"])
        avg_ttft = mean(results["ttfts"])
        avg_tokens = mean(results["tokens"])
        avg_tps = avg_tokens / avg_latency if avg_latency > 0 else 0
        
        print(f"\n{name}:")
        print(f"  Avg Total Latency: {avg_latency:.2f}s")
        print(f"  Avg TTFT: {avg_ttft:.3f}s")
        print(f"  Avg Tokens: {avg_tokens:.0f}")
        print(f"  Avg TPS: {avg_tps:.1f} tok/s")
        print(f"  Errors: {results['errors']}/{ITERATIONS}")
    
    print(f"\n{'='*80}")

if __name__ == "__main__":
    all_results = []
    
    for config in CONFIGS:
        result = benchmark_endpoint(config)
        all_results.append(result)
    
    print_summary(all_results)
```

## Usage

```bash
chmod +x benchmark_llm.py
python3 benchmark_llm.py
```

## Example Results (9router vs External)

```
9router Local:
  Avg Total Latency: 3.99s
  Avg TTFT: 3.989s
  Avg Tokens: 44
  Avg TPS: 11.1 tok/s
  Errors: 0/5

LakapVIP External:
  Avg Total Latency: 5.70s
  Avg TTFT: 5.684s
  Avg Tokens: 49
  Avg TPS: 8.6 tok/s
  Errors: 0/5

🏆 Winner: 9router Local (30.1% faster)
```

## Metrics

- **Latency**: Total time from request start to completion
- **TTFT** (Time To First Token): Time until first streaming chunk arrives
- **TPS** (Tokens Per Second): Output throughput (tokens / latency)

## Notes

- Always use correct model naming (e.g., `kr/claude-sonnet-4.5` for 9router)
- HTTP 404 errors indicate wrong model name or provider not configured
- Local 9router typically 25-35% faster than external routers (zero network overhead)
