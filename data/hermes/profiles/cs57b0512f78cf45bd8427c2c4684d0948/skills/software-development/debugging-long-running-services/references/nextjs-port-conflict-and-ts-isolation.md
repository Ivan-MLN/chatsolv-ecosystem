# Next.js Multi-App Port Conflict & TypeScript Root Pollution Case

## Scenario
Spinning up a new Next.js project on a VPS running multiple PM2 services.

### Issues Encountered
1. **Next.js TypeScript Parent Types Leakage**:
   When building a sub-project located under `/home/nldt/chatsolv-nextgen`, `next build` failed with:
   ```text
   error TS2688: Cannot find type definition file for 'bcryptjs'.
   ```
   Root cause: TypeScript searches upwards and found declarations in `/home/nldt/node_modules` or adjacent projects.
   **Solution**: In `tsconfig.json`, set `"types": []` inside `compilerOptions`.

2. **Port 3000 EADDRINUSE Conflict**:
   Port 3000 was already bound by PM2 frontend service.
   **Solution**:
   - Run on custom port via PM2: `pm2 start "npm run start -- -p 3333" --name "<app-name>" --cwd "/path/to/app"`
   - Open firewall port for external traffic: `sudo ufw allow 3333/tcp`
   - Test accessibility via `curl -4 -s ifconfig.me` and verify `http://<IP>:3333`.
