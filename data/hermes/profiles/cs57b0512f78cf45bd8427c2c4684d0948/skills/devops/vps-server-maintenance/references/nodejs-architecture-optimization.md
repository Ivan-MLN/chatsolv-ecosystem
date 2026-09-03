# Node.js Architecture & Multi-Core Optimization

## Build vs Runtime: src/ vs dist/

**src/** = Source code (TypeScript, JSX, etc.)
- Human-readable code you edit during development
- Contains type annotations, modern syntax
- Cannot be executed directly by Node.js

**dist/** = Distribution/build output
- Compiled/transpiled JavaScript from src/
- Optimized for production execution
- What pm2 actually runs

**Workflow**: Edit src/ → `npm run build` compiles to dist/ → `pm2 restart` runs from dist/

### Why Not Run TypeScript Directly?

Node.js runtime only understands JavaScript. TypeScript features (type annotations, interfaces, enums) must be stripped first.

**Options**:
1. **Pre-compile (current nael-ai approach)** - Most efficient for production
   - `tsc` or `esbuild` compiles once → dist/
   - Zero runtime overhead
   - Fastest startup & lowest latency
   
2. **On-the-fly compilation** - Convenient for dev
   - `tsx` or `ts-node`: compile during startup
   - Extra memory overhead + slower cold starts
   - Good for rapid iteration

3. **Alternative runtimes**
   - `bun`: Native TypeScript support, golang-speed compilation
   - `deno`: Also native TypeScript

**Performance ranking** (startup speed & memory efficiency):
1. node + pre-compiled JS ⚡ (fastest, lowest memory)
2. bun native TS ⚡ (close second, sometimes faster)
3. tsx/ts-node 🐢 (slowest, higher memory usage)

## Single-Threaded vs Multi-Core Node.js

Node.js is **single-threaded by default** (event loop on one CPU core).

### Multi-Threading Options

**1. Worker Threads** (native Node.js)
```javascript
import { Worker } from 'worker_threads';
const worker = new Worker('./workers/audio-processor.js');
```
- Best for CPU-heavy tasks: audio/video processing, image manipulation, large parsing
- Main thread stays responsive (handles WebSocket, HTTP)
- Shared memory via `SharedArrayBuffer`

**2. PM2 Cluster Mode**
```bash
pm2 start dist/index.js -i 4  # spawn 4 instances
```
- ⚠️ **Baileys socket caveat**: Only ONE socket connection per auth session allowed
- If clustered naively, each instance tries to connect → kicks others offline
- **Solution**: Single master handles socket, workers handle processing via queue

**3. Hybrid Architecture** (recommended for WhatsApp bots)
```bash
# Main instance: socket handler only (1 core)
pm2 start dist/index.js --name bot-main -i 1

# Worker pool: job processing (3 cores)
pm2 start dist/worker.js --name bot-worker -i 3
```
- Main process: lightweight socket + message routing
- Workers: CPU-heavy tasks (media conversion, AI calls) via Redis/BullMQ

## nael-ai Process Architecture Discovery (Aug 2026)

**Key finding**: nael-ai is a **monolithic process** that includes:
- WhatsApp client (Baileys socket)
- MCP server on port 5778 (eval_code, eval_shell, wa_* tools)
- Media processing (sox, ffmpeg via child_process)
- Memory footprint: ~192MB (acceptable for combined bot+MCP)

**Related services**:
- 9router (LLM gateway): next-server on port 20128 (~244MB)
- Hermes Agent: Python inference process (~205MB)

**Port ownership**:
- `:5778` = nael-ai built-in Baileys MCP (owner: 6283893964069)
- `:5788/:5789` = Takim-AI instance under /home/naeladtya (owner: 628388024064)

**Optimization paths**:
1. Extract media processing to Worker Threads (sox/ffmpeg)
2. Lazy-load MCP tool modules (don't import all at startup)
3. Consider switching to `bun` runtime for ~30-40% memory reduction
4. Keep socket handler single-threaded (Baileys requirement)

## When NOT to Optimize

Current setup is already near-optimal if:
- Memory usage is stable (<200MB for combined bot+MCP)
- Response latency is acceptable
- No CPU bottlenecks during peak load

**Real memory hogs on dev VPS** (from ps aux --sort=-%mem):
- VSCode server instances (500MB-1.5GB total) ← close when not coding
- n8n automation (294MB)
- Multiple Next.js dev servers

Optimize those first before micro-optimizing the bot.
