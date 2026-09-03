# Tailscale Mesh VPN & OpenSSH Remote Control Setup

Use this procedure when setting up remote device control for Hermes / VPS agents to execute commands on remote laptops/desktops (Windows or Linux) securely without public IP or port forwarding.

## 1. Tailscale Installation & Authentication

### On Debian/Ubuntu VPS:
If `apt-get update` fails due to unverified third-party repos (e.g. cloudflared/speedtest GPG key errors during `curl -fsSL https://tailscale.com/install.sh | sh`):
```bash
apt-get update --allow-insecure-repositories && apt-get install -y tailscale
```
Authenticate the VPS:
```bash
tailscale up
# Copy the login URL (https://login.tailscale.com/a/...) and complete authentication in browser
```

### On Windows Remote Laptop:
1. Download installer from `https://tailscale.com/download` and complete GUI installation.
2. Sign in with the **same** Tailscale account as the VPS.
3. Check Tailscale IPv4 address: `tailscale ip -4` or via System Tray.

---

## 2. Windows OpenSSH Server Setup (PowerShell Admin)

Execute the following commands in PowerShell (Run as Administrator) on the target Windows machine:

```powershell
# 1. Install & Start OpenSSH Server
Add-WindowsCapability -Online -Name OpenSSH.Server~~~~0.0.1.0
Start-Service sshd
Set-Service -Name sshd -StartupType 'Automatic'

# 2. Add Firewall Rule
if (!(Get-NetFirewallRule -Name "OpenSSH-Server-In-TCP" -ErrorAction SilentlyContinue)) {
    New-NetFirewallRule -Name 'OpenSSH-Server-In-TCP' -DisplayName 'OpenSSH Server (sshd)' -Enabled True -Direction Inbound -Protocol TCP -Action Allow -LocalPort 22
}
```

---

## 3. SSH Key Authorization on Windows Target

On the remote Windows machine (PowerShell user session):

```powershell
# Create .ssh directory and write authorized_keys
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.ssh"
Set-Content -Path "$env:USERPROFILE\.ssh\authorized_keys" -Value "<PUBLIC_KEY_FROM_VPS>"

# Reset permissions so sshd accepts the file
icacls "$env:USERPROFILE\.ssh\authorized_keys" /inheritance:r /grant "$($env:USERNAME):F"
```

---

## 4. VPS Helper / Wrapper Script Setup

On the VPS, generate SSH key pair and create a wrapper command (e.g., `/usr/local/bin/thinkpad`):

```bash
# Generate key if not present
ssh-keygen -t ed25519 -f ~/.ssh/thinkpad_key -N "" -q

# Create wrapper script
cat << 'EOF' > /usr/local/bin/thinkpad
#!/bin/bash
ssh -i ~/.ssh/thinkpad_key -o StrictHostKeyChecking=no USERNAME@100.X.Y.Z "$@"
EOF

chmod +x /usr/local/bin/thinkpad
```

Testing connection from VPS:
```bash
thinkpad "whoami && hostname"
```
