# VPS Maintenance Case Study: OOM Crash & Disk Cleanup

## Incident Context
VPS experienced unexpected reboots caused by Linux Kernel Out-Of-Memory (OOM) killer during high RAM usage (e.g. attempting to run heavy local LLM models in Ollama on a 16GB RAM instance without Swap).

## Analysis Findings
1. **Swap Memory**: `free -h` showed `Swap: 0B`. Any RAM allocation exceeding total physical memory resulted in instant OOM kill / panic.
2. **PM2 Logs**: `/root/.pm2/logs` accumulated over 13GB of unhandled log output over time.
3. **Cache & Archives**: `~/.npm/_cacache` held ~12GB, and outdated root zip archives (`30_JULI_ROOT_BACKUP.zip`, `cr03juli.zip`) held ~12.7GB.

## Remediation Steps Taken
1. Created 4GB swapfile at `/swapfile`, set `chmod 600`, activated via `swapon`, and appended to `/etc/fstab`.
2. Executed `pm2 flush` to wipe accumulated PM2 logs.
3. Cleaned `npm` cache and deleted old root backup archives, freeing ~21GB disk space total (disk usage dropped from 72% / 136GB to 62% / 116GB).
5. Clarified Bot IPC vs HTTP architecture: `nael-ai` communicates with `hermes acp` using `child_process.spawn` and JSON-RPC over stdio streams, while LLM HTTP API requests are dispatched internally by the Hermes process to `127.0.0.1:20128/v1`.
