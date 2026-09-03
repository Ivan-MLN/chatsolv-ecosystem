# 9router Internal Systemd & Custom Server Architecture

This reference outlines the internal script structure and systemd orchestration used by 9router on this server.

## 1. Systemd Service Definition (`/etc/systemd/system/9router.service`)

Runs the standalone Next.js server directly with custom memory and DNS optimizations, bypassing CLI overhead:

```ini
[Unit]
Description=9Router AI Gateway
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=/usr/local/lib/node_modules/9router/app
ExecStart=/usr/local/bin/node --dns-result-order=ipv4first --max-old-space-size=6144 /usr/local/lib/node_modules/9router/app/custom-server.js
Restart=always
RestartSec=5

Environment=HOME=/root
Environment=NODE_ENV=production
Environment=PORT=20128
Environment=HOSTNAME=0.0.0.0
Environment=NODE_PATH=/root/.9router/runtime/node_modules:/usr/local/lib/node_modules/9router/app/node_modules

StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

## 2. HTTP Server Wrapper (`/usr/local/lib/node_modules/9router/app/custom-server.js`)

Core capabilities in the wrapper:
- **IP Sanitization & Header Protection**: Derives client IP from the TCP socket directly to prevent client spoofing via `x-forwarded-for`, and attaches an internal `x-9r-peer-token`.
- **Background Token Refresh**: Automatically initializes `startBackgroundTokenRefresh()` from `src/sse/services/backgroundTokenRefresh.js` when the HTTP server starts listening.
- **HTTP/2 Cleartext (h2c) Downgrade**: Catches h2c upgrade attempts and proxies them gracefully through HTTP/1.1 handlers.
- **Next.js Standalone Boot**: Direct entry point loading `/usr/local/lib/node_modules/9router/app/server.js`.
