# 9router Daemon Autostart & Background Persistence

Guide to ensuring 9router runs persistently in the background across reboot/restart across platforms.

## 1. Systemd Service (Linux / Ubuntu / Debian)

Create `/etc/systemd/system/9router.service`:

```ini
[Unit]
Description=9router OpenAI Compatible API Proxy Service
After=network.target network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/9router
Restart=always
RestartSec=5
KillMode=process
LimitNOFILE=65536
Environment=PORT=20128

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable 9router
sudo systemctl restart 9router
sudo systemctl status 9router
```

## 2. PM2 (Multi-Platform / Node.js)

```bash
pm2 start "9router" --name 9router
pm2 save
pm2 startup
```
Execute the printed `sudo env PATH=...` command to register the startup hook with init/systemd.

## 3. Windows (NSSM / Task Scheduler)

Using NSSM (Non-Sucking Service Manager):
```cmd
nssm install 9router "C:\path\to\9router.exe"
nssm set 9router AppRestartDelay 5000
nssm start 9router
```
