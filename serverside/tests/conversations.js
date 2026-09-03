import { sendRequest, getAuthContext, getHmacHeaders, uid } from './client.js';

console.log(`\n============================================================`);
console.log(`       9. CONVERSATIONS & HUMAN TAKEOVER REQUEST TESTS      `);
console.log(`============================================================`);

async function run() {
  const { token, workspaceID } = await getAuthContext();

  // 1. Create a live WhatsApp channel & simulate an incoming message to spawn a conversation
  const connectRes = await sendRequest('Setup: Connect WA Channel', 'POST', `/api/v1/channels/whatsapp/connect?workspace_id=${workspaceID}`, {
    token,
    body: {
      display_name: 'Customer Support WA',
    },
  });
  const channelID = connectRes.data?.data?.channel_id;

  if (channelID) {
    const externalUser = `628123456${Math.floor(Math.random() * 10000)}`;
    const incomingPayload = {
      channel_id: channelID,
      external_user_id: externalUser,
      sender_name: 'Customer Budi',
      message_id: uid('wam'),
      message_type: 'text',
      content: { text: 'Halo CS, apakah paket saya sudah dikirim?' },
      raw_payload: {},
    };

    await sendRequest('Setup: Simulate Incoming WA Message', 'POST', '/internal/v1/messages/incoming', {
      headers: getHmacHeaders(incomingPayload),
      body: incomingPayload,
    });
  }

  // 2. List Conversations
  const convListRes = await sendRequest('9.1 List Workspace Conversations', 'GET', `/api/v1/conversations?workspace_id=${workspaceID}`, {
    token,
  });

  const conversationID = convListRes.data?.data?.items?.[0]?.id;

  if (conversationID) {
    // 3. Get Conversation Details
    await sendRequest('9.2 Get Conversation Detail by ID', 'GET', `/api/v1/conversations/${conversationID}`, {
      token,
    });

    // 4. Get Conversation Messages
    await sendRequest('9.3 Get Conversation Message History', 'GET', `/api/v1/conversations/${conversationID}/messages`, {
      token,
    });

    // 5. Takeover Conversation to Human Mode
    await sendRequest('9.4 Takeover Mode to Human (CS Agent)', 'PATCH', `/api/v1/conversations/${conversationID}/mode`, {
      token,
      body: {
        mode: 'human',
      },
    });

    // 6. Resume Conversation to Agent Mode (AI Bot)
    await sendRequest('9.5 Resume Mode to AI Agent', 'PATCH', `/api/v1/conversations/${conversationID}/mode`, {
      token,
      body: {
        mode: 'agent',
      },
    });
  }
}

run();
