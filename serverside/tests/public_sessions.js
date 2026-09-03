import { sendRequest, getAuthContext } from './client.js';

console.log(`\n============================================================`);
console.log(`       12. PUBLIC WEBSITE AGENT API REQUEST TESTS           `);
console.log(`============================================================`);

async function run() {
  const { token, workspaceID } = await getAuthContext();

  // 1. Create Secret API Key to authenticate public sessions
  const keyRes = await sendRequest('Setup: Create Developer API Key', 'POST', `/api/v1/api-keys?workspace_id=${workspaceID}`, {
    token,
    body: {
      name: 'Public Website Key',
      role: 'member',
      scopes: ['conversation:write'],
    },
  });

  const rawApiKey = keyRes.data?.data?.raw_key;

  if (rawApiKey) {
    // 2. Create Public Agent Session
    const sessionRes = await sendRequest('12.1 Create Public Agent Session (Visitor)', 'POST', '/api/v1/agent-sessions', {
      headers: {
        'X-API-Key': rawApiKey,
      },
      body: {
        visitor_id: 'visitor_web_7788',
        metadata: {
          browser: 'Chrome 128',
          page: 'https://myshop.com/products/laptop',
        },
      },
    });

    const sessionID = sessionRes.data?.data?.session_id;
    const sessionToken = sessionRes.data?.data?.session_token;

    if (sessionID && sessionToken) {
      // 3. Send Message in Public Session
      await sendRequest('12.2 Send Message in Public Session', 'POST', `/api/v1/agent-sessions/${sessionID}/messages`, {
        headers: {
          Authorization: `Bearer ${sessionToken}`,
        },
        body: {
          content: 'Halo, apakah ada diskon untuk pembelian laptop hari ini?',
        },
      });

      // 4. Stream Message (SSE)
      await sendRequest('12.3 Stream AI Response (SSE)', 'POST', `/api/v1/agent-sessions/${sessionID}/messages/stream`, {
        headers: {
          Authorization: `Bearer ${sessionToken}`,
        },
        body: {
          content: 'Tolong jelaskan spesifikasinya secara singkat.',
        },
      });
    }
  }
}

run();
