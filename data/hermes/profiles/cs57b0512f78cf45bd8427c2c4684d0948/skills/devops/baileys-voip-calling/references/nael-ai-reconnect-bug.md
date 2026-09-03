# nael-ai Reconnection Handler Bug (code 500)

## Symptom
Bot disconnects every few hours with:
```
ERROR: Session invalid, re-pairing required {"code":500}
ERROR: Delete "sessions" and pair again {"code":500}
```

Session is actually still valid — restarting pm2 reconnects without re-pairing.

## Root Cause
`src/handlers/connectionHandler.ts` treats **all non-FATAL_CODES as potentially fatal** when `attempts >= maxReconnectAttempts`.

Code 500 (internal server error) is a **transient server-side error**, not a session corruption. It should trigger normal reconnect, not session deletion.

The original logic:
```typescript
if (FATAL_CODES.has(code)) {
  logger.error({ code }, "Session invalid, re-pairing required");
  onFatal(code);
  return;
}

if (attempts >= config.maxReconnectAttempts) {
  logger.error({ attempts }, "Reconnect attempts exhausted");
  onFatal(code);  // ❌ treats 500 as fatal after retries exhausted
  return;
}
```

## Fix
Add explicit handling for 5xx server errors before the exhaustion check:

```typescript
const code = statusCodeOf(lastDisconnect?.error);

// Only treat truly fatal disconnect reasons as unrecoverable.
// 500 (internal server error) should reconnect, not delete session.
if (FATAL_CODES.has(code)) {
  logger.error({ code }, "Session invalid, re-pairing required");
  onFatal(code);
  return;
}

// Ignore transient errors (5xx server errors) and let reconnect handle them
if (code >= 500 && code < 600) {
  logger.warn({ code }, "Server error, will reconnect");
}

if (attempts >= config.maxReconnectAttempts) {
  logger.error({ attempts }, "Reconnect attempts exhausted");
  onFatal(code);
  return;
}
```

Now 5xx errors are logged as warnings and will keep reconnecting indefinitely instead of exhausting retries and treating them as fatal.

## Verified Fix
- Applied to `/root/nael-ai/src/handlers/connectionHandler.ts`
- Build passed
- Bot restarted and connected successfully
- Awaiting multi-hour stability test

## True Fatal Codes (require re-pairing)
```typescript
const FATAL_CODES: ReadonlySet<number> = new Set([
  DisconnectReason.loggedOut,        // 401
  DisconnectReason.forbidden,        // 403
  DisconnectReason.multideviceMismatch,
  DisconnectReason.badSession,       // 440
]);
```

Code 500 is NOT in this list and should never be.
