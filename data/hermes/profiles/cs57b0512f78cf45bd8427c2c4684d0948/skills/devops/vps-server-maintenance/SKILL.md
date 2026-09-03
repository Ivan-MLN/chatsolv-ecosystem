---
name: vps-server-maintenance
description: "VPS memory/swap setup, disk reclamation, and OOM fix."
version: 1.0.0
author: Hermes Agent
license: MIT
platforms: [linux]
metadata:
  hermes:
    tags: [vps, devops, swap, oom, disk-cleanup, memory, pm2, linux]
    related_skills: [debugging-long-running-services, systematic-debugging]
---

# VPS Server Maintenance & OOM Prevention

## When to use

Use when:
- VPS experiences OOM (Out Of Memory) crashes, kernel process kills, or unexpected reboots.
- RAM usage is high with zero swap space configured.
- Disk space is low or partitions are running out of storage space.
- Cleaning up accumulated logs (`pm2`, system logs), package caches (`npm`, `pip`, `docker`), or old backup archives.

## Core Procedures

### 1. Emergency Swap Creation (OOM Prevention)

If `free -h` shows `Swap: 0B`, create a persistent swapfile immediately to prevent Linux kernel OOM panics when RAM spikes:

```bash
# Create 4GB swapfile
fallocate -l 4G /swapfile || dd if=/dev/zero of=/swapfile bs=1M count=4096

# Secure permissions
chmod 600 /swapfile

# Set up swap space
mkswap /swapfile
swapon /swapfile

# Make swap persistent across reboots
echo '/swapfile none swap sw 0 0' >> /etc/fstab

# Verify
free -h
```

### 2. Disk Space Inspection & Reclamation

Find top storage consumers:

```bash
# Check filesystem usage
df -h /

# Find largest directories in /root or home
du -h --max-depth=2 /root 2>/dev/null | sort -rh | head -20
```

Common cleanup targets:
- **PM2 Log Accumulation**: `pm2 flush` (reclaims gigabytes of stale log files in `~/.pm2/logs`).
- **NPM Cache**: `rm -rf ~/.npm/_cacache`
- **Old Zip / Root Backups**: Check for stale `.zip`, `.tar.gz`, or `.mp4` recordings in `/root` or home directories.
- **Docker Prune** (if docker is used): `docker system prune -f`

### 3. Memory & High-RAM Process Triage

Identify high-memory processes before they trigger OOM:

```bash
ps aux --sort=-%mem | head -15
```

Key memory hogs on dev/bot VPS servers:
- Local LLM runners (Ollama / llama.cpp) attempting to load models exceeding available physical RAM.
- Stale daemon workers / memory leaks (e.g. Bun worker daemons, orphaned VSCode server sub-processes).
- Uncapped Node.js / Next.js app instances.

### 4. Clearing Linux Kernel RAM Page Cache (Drop Caches)

When requested to clear RAM cache (PageCache, dentries, and inodes) without clearing disk storage:

```bash
# Flush dirty filesystem buffers to disk first, then drop caches
sync && echo 3 > /proc/sys/vm/drop_caches
```

Values for `/proc/sys/vm/drop_caches`:
- `1`: Free PageCache only
- `2`: Free dentries and inodes
- `3`: Free PageCache, dentries, and inodes

### 4. Clearing Linux Kernel RAM Page Cache (Drop Caches)

When requested to clear RAM cache (PageCache, dentries, and inodes) without clearing disk storage:

```bash
# Flush dirty filesystem buffers to disk first, then drop caches
sync && echo 3 > /proc/sys/vm/drop_caches
```

Values for `/proc/sys/vm/drop_caches`:
- `1`: Free PageCache only
- `2`: Free dentries and inodes
- `3`: Free PageCache, dentries, and inodes

### 5. IPC Architecture & Hermes Bot Integration Diagnostics

When inspecting Node.js bot / AI agent integrations (e.g. `nael-ai` or Hermes ACP clients):
- **Communication Mechanism**: Bot clients launch `hermes acp` via `child_process.spawn` (long-running process with stdio pipe), **not** `child_process.exec` (which waits for process exit and buffers output).
- **Protocol**: JSON-RPC 2.0 frames over stdout/stdin using newline-delimited stream readers (`FrameReader`).
- **LLM HTTP Requests**: HTTP requests to LLM endpoints/gateways (`127.0.0.1:20128/v1`) are handled inside the Hermes binary engine (`httpx`/`urllib` in Python), not directly via Axios/node-fetch inside the Node.js wrapper.

### 6. Cloudflare Tunnel DNS Routing & Web Container Default Vault Troubleshooting

#### Cloudflare Tunnel DNS Routing
When a service port (e.g. Obsidian Web on port 8443) is tunnelled via `cloudflared` but the subdomain returns `NXDOMAIN` / connection refused:
- Verify active tunnel process (`cloudflared tunnel run --url http://0.0.0.0:<port> <tunnel-name>`).
- Add missing CNAME DNS route:
  ```bash
  cloudflared tunnel route dns <tunnel-name> <subdomain>
  ```
- Test resolution & HTTP connection: `curl -Iv --resolve <subdomain>:443:<cf-ip> https://<subdomain>` or directly `curl -Iv https://<subdomain>`.

#### Docker Web Apps / Obsidian-Web Vault Path Reset
When deploying `linuxserver/docker-obsidian` or similar containerized web app interfaces:
- **Issue**: Image default configs (e.g. `/config/.config/obsidian/obsidian.json`) often create a fresh empty vault under `/config/Obsidian Vault` with a "Selamat datang" note, bypassing mounted data at `/vault`.
- **Fix**: Stop the container, overwrite `/config/.config/obsidian/obsidian.json` to point `"path"` to the mounted vault (e.g., `"/vault"`), and start the container:
  ```json
  {"vaults":{"a1b2c3d4e5f67890":{"path":"/vault","ts":1786415860564,"open":true}},"language":"id"}
  ```

### 7. Isolated Secondary Bot Instance Setup (Multi-User Isolation)

When setting up a separate bot instance for another VPS user (e.g., cloning `nael-ai` to `/home/<user>/Takim-AI`):
- Refer to `references/isolated-multi-user-instance-setup.md` for step-by-step port remapping (e.g., Baileys MCP `5788`, Sticker MCP `5789`), VoIP pre-connect disabling, permissions chown, and isolated PM2 setup.

### 8. Node.js Process Architecture & Multi-Core Optimization

When troubleshooting or optimizing Node.js bots (nael-ai, Takim-AI) or services:
- Refer to `references/nodejs-architecture-optimization.md` for:
  - src/ vs dist/ build workflow
  - TypeScript runtime options (pre-compile vs on-the-fly)
  - Multi-threading strategies (Worker Threads, PM2 cluster, hybrid architecture)
  - Baileys socket clustering caveats (single connection per auth)
  - nael-ai process architecture (monolithic bot+MCP on port 5778)
  - Performance ranking and optimization decision trees

### 9. Tailscale & OpenSSH Remote Control Setup (VPS to Remote Laptop/Desktop)

When connecting VPS agents to control remote machines (Windows/Linux laptops/desktops) securely over Tailscale mesh VPN:
- Refer to `references/tailscale-remote-host-control.md` for:
  - Apt error workarounds during Tailscale install on Debian (`--allow-insecure-repositories`)
  - OpenSSH server installation, firewall rule configuration, and service auto-start on Windows (PowerShell)
  - Key-based passwordless authorization and ACL permissions on Windows (`authorized_keys` + `icacls`)
  - VPS CLI wrapper script setup for execution commands.


### 10. Large Directory / Root VPS Archiving & Remote Transfer

When creating full backups of `/root` or large VPS directories to zip archives and transferring them offsite:

#### Exclusion Rules
- **Node Modules & IDE Caches**: `*/node_modules/*`, `*/.vscode-server/*`.
- **Active Logs & Sockets**: `*.log`, `*/.pm2/logs/*`, `*.sock`, socket locks (`SingletonLock`, `SingletonCookie`).
- **Selective Image/Screenshot Filters**: Avoid blanket exclusions like `*.jpg`/`*.png` (which drops legitimate static media assets); use wildcard pattern matching like `*screensho*`, `*screenshot*`, `*Screensho*`, `*Screenshot*` to specifically exclude target screenshot assets.

#### Execution & Performance Strategy
- **Background Execution**: For archives >5GB, running `zip` in foreground will hit tool timeouts (10-minute cap). Always run via `terminal(background=True, notify_on_complete=true)` and monitor via `process(action='poll'/'wait')`.
- **Fast Archiving (`zip -0`)**: Use store-only mode (`zip -0`) if disk I/O or CPU compression overhead causes timeouts on huge file trees (~20GB+).
- **Remote SSH/SCP Monitoring**: Monitor background SCP transfers to remote hosts (e.g., Windows laptop via Tailscale) using remote PowerShell size checks (`Get-ChildItem <file> | Select-Object Length, LastWriteTime`) rather than assuming immediate completion.

## Pitfalls

- **Do NOT create swap inside temporary filesystems** (`/tmp` or `tmpfs`). Always place `/swapfile` on the root filesystem.
- **Avoid blanket media extension exclusions (`*.jpg`)**: When users ask to ignore "screenshots", do not exclude all images; filter specifically by filename pattern (`*screensho*`).
- **Socket and Lock File Traps**: Always exclude socket files (`*.sock`, `SingletonLock`) when zipping active root directories, as `zip` can hang or log warnings on special IPC files.
- **Check `/etc/fstab` before appending**: avoid duplicate swap entries if an inactive swapfile exists.
- **Never `rm -rf ~/.pm2/logs` while PM2 is actively writing** if file handles might get corrupted; prefer `pm2 flush` or truncate log files.

## Verification

Confirm system stability:
1. `free -h` confirms active swap space (e.g., `Swap: 4.0Gi`).
2. `df -h /` confirms reclaimed disk space and healthy free space margin.
3. `/etc/fstab` contains `/swapfile none swap sw 0 0`.
