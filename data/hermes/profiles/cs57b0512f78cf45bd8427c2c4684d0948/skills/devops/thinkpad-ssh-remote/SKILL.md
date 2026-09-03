---
name: thinkpad-ssh-remote
description: Access ThinkPad E490 via Tailscale SSH wrapper.
---

# ThinkPad E490 SSH & 9Router Automation Access

## Target Information
- **Device**: Lenovo ThinkPad E490 (Intel Core i5 8th Gen)
- **OS**: Windows 11 Pro 64-bit
- **Tailscale IP**: `100.77.36.29`
- **Username**: `thinkpad` (Domain/Host: `nael\thinkpad`)
- **SSH Key**: `~/.ssh/thinkpad_key`
- **Wrapper Executable**: `/usr/local/bin/thinkpad`

## 9Router Commands & Navigation Rules
- **9Router Port**: `20128` (`http://76.13.193.152:20128`)
- **PIN**: `147000`
- **Cek Usage 9router**: Buka/navigasi ke `/dashboard/usage`
- **Cek Limit 9router**: Buka/navigasi ke `/dashboard/quota`
- **Headless Screenshot Automation**: Gunakan script puppeteer headless di laptop untuk screenshot bersih dengan CSS penuh.

