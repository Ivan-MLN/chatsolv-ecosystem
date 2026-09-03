# Baileys Store and Auth State Gotchas (@whiskeysockets/baileys)

## 1. `makeInMemoryStore` Removal in Modern Baileys (v6.7+)
- In newer versions of `@whiskeysockets/baileys` (v6.7.x+), `makeInMemoryStore` has been completely removed from top-level exports due to RAM inefficiency and memory leaks.
- Attempting `import { makeInMemoryStore } from '@whiskeysockets/baileys'` or `require('@whiskeysockets/baileys').makeInMemoryStore` returns `undefined`.

## 2. Outdated Official Documentation / README
- As of 2026, the official GitHub README for `WhiskeySockets/Baileys` (`master` branch) is outdated and still shows example code importing `makeInMemoryStore`.
- Do not rely on README code snippets for store instantiation.

## 3. Recommended Auth & Store Implementations

### Session / Auth State Persistence
Use `useMultiFileAuthState`:
```javascript
const { default: makeWASocket, useMultiFileAuthState } = require('@whiskeysockets/baileys');

async function start() {
    const { state, saveCreds } = await useMultiFileAuthState('auth_info');
    const sock = makeWASocket({
        auth: state,
        printQRInTerminal: true
    });
    sock.ev.on('creds.update', saveCreds);
}
```

### In-Memory Message Caching for `getMessage`
For message retries, quotes, and poll votes, implement a custom `Map`-backed store:
```typescript
const messageStore = new Map<string, proto.IMessage>();

const sock = makeWASocket({
  auth: state,
  getMessage: async (key) => {
    const id = `${key.remoteJid}:${key.id}`;
    return messageStore.get(id);
  },
});

sock.ev.on('messages.upsert', ({ messages }) => {
  for (const msg of messages) {
    if (msg.key.id && msg.message) {
      const id = `${msg.key.remoteJid}:${msg.key.id}`;
      messageStore.set(id, msg.message);
    }
  }
});
```
