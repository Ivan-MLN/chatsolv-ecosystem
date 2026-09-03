# Cloudflare Tunnel & Docker Obsidian Vault Fix Notes

## Issue Summary
1. **Cloudflare Tunnel CNAME Route Missing**:
   - `cloudflared` process running `obsi` mapping to `http://0.0.0.0:8443`.
   - Domain `obsi.r3d4ct3d.mom` had no DNS CNAME route configured, resulting in `NXDOMAIN` / `ERR_NAME_NOT_RESOLVED` on clients.
   - **Fix**: Executed `cloudflared tunnel route dns obsi obsi.r3d4ct3d.mom` under the tunnel user.

2. **Docker Obsidian Default Vault Overwrite**:
   - Container `linuxserver/docker-obsidian` mounted `/root/ObsidianVault` to `/vault`.
   - Container default configuration at `/config/.config/obsidian/obsidian.json` auto-generated a vault at `/config/Obsidian Vault` with `Selamat datang.md`.
   - **Fix**: Stopped container (`docker stop obsidian-web`), updated `/config/.config/obsidian/obsidian.json` to point `"path"` to `/vault`, and restarted container (`docker start obsidian-web`).
