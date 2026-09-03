# Isolated Multi-User Instance Setup & Port Mapping

## Context
When cloning or creating an isolated secondary instance of a Node.js / Baileys / AI bot (such as `nael-ai` -> `Takim-AI`) for another system user (e.g. `naeladtya`):

### Key Considerations & Steps
1. **Directory & Ownership**:
   - Copy source code without `node_modules`, `dist`, session data, `.voip-auth`, or `.git` to the target user's home directory (e.g., `/home/naeladtya/Takim-AI`).
   - Run `chown -R <username>:<username> <target-dir>` to ensure full permission isolation.

2. **Persona & Descriptor Customization**:
   - Update persona name definitions in `src/ai/hermes.ts` (`const PERSONA = "Novi"`), handler prompts in `src/handlers/messageHandler.ts`, tests in `src/ai/hermes.check.ts`, and project descriptions in `package.json`.

3. **Port Allocation & Avoidance**:
   - Primary instance ports: Baileys MCP (`5778`), Sticker MCP (`5779`), 9router (`20128`).
   - Secondary instance ports: Shift to a distinct non-conflicting range like Baileys MCP (`5788`), Sticker MCP (`5789`).
   - Update `process.env` configuration (e.g., `.env`), `index.ts` default port parameters, and `sticker.ts` MCP URL.

4. **VoIP / WebRTC Pre-connection**:
   - For fresh instances or secondary accounts, disable heavy startup pre-connections (such as `initVoipClient()`) in `src/ai/mcp-baileys.ts` to prevent stale WebRTC state or socket collisions on startup.

5. **Environment & Shared Gateway Isolation**:
   - Set isolated session store paths (`NAEL_SESSION_STORE=/home/<user>/.takim_sessions.json`).
   - Configure user-specific `ELIGIBLE_JIDS` and `OWNER_JIDS` in the isolated `.env`.
   - **Shared 9router Gateway**: When sharing root's 9router LLM gateway (e.g. port `20128`), update `WORKDIR` in `src/ai/acp.ts` (e.g. `/home/<user>/Takim-AI`) so `hermes acp` spawns in the user's workspace rather than `/root/nael-ai`.
   - **Pairing & Number Change Procedure**: If changing `PAIR_NUMBER` or encountering 401 connection failure on initial launch, stop PM2, purge the `sessions/` directory (`rm -rf sessions`), update `PAIR_NUMBER=<target_number>` in `.env`, restart PM2, and tail logs to capture the fresh pairing code.

6. **PM2 & Build Setup**:
   - Run `npm install` and `npm run build` under the target user (`su - <user> -c "..."`).
   - Configure a dedicated `ecosystem.config.cjs` using `node --enable-source-maps dist/index.js` under the target user's PM2 daemon space.
