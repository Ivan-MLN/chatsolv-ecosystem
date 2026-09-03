#!/usr/bin/env python3
"""
Benchmark LLM endpoints: latency, TPS, TTFT
Compares multiple OpenAI-compatible endpoints for performance testing.
"""
import time
import json
import requests
from statistics import mean

# Test configs - modify as needed
CONFIGS = [
    {
        "name": "9router Local",
        "base_url": "http://127.0.0.1:20128/v1",
        "api_key": "dummy",  # 9router doesn't need real key
        "model": "kr/claude-sonnet-4.5"
    },
    {
        "name": "External Provider",
        "base_url": "https://your-provider.com/api/v1",
        "api_key": "your-api-key-here",
        "model": "your-model-name"
    }
]

TEST_PROMPT = "Explain quantum computing in exactly 100 words."
ITERATIONS = 5

def benchmark_endpoint(config):
    """Run benchmark on single endpoint"""
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
        
        # Small delay between requests
        time.sleep(1)
    
    return results

def print_summary(all_results):
    """Print comparison summary"""
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
        
        print(f"{name}:")
        print(f"  Avg Total Latency: {avg_latency:.2f}s")
        print(f"  Avg TTFT: {avg_ttft:.3f}s")
        print(f"  Avg Tokens: {avg_tokens:.0f}")
        print(f"  Avg TPS: {avg_tps:.1f} tok/s")
        print(f"  Errors: {results['errors']}/{ITERATIONS}\n")
    
    print(f"{'='*80}")
    
    # Winner calculation
    if len(all_results) >= 2:
        valid_results = [r for r in all_results if len(r["latencies"]) > 0]
        if len(valid_results) >= 2:
            sorted_by_latency = sorted(valid_results, key=lambda r: mean(r["latencies"]))
            winner = sorted_by_latency[0]
            runner_up = sorted_by_latency[1]
            
            winner_avg = mean(winner["latencies"])
            runner_avg = mean(runner_up["latencies"])
            diff_pct = ((runner_avg - winner_avg) / runner_avg) * 100
            
            print(f"🏆 Winner: {winner['name']} ({diff_pct:.1f}% faster)")
    
    print(f"{'='*80}\n")

if __name__ == "__main__":
    all_results = []
    
    for config in CONFIGS:
        result = benchmark_endpoint(config)
        all_results.append(result)
    
    print_summary(all_results)
