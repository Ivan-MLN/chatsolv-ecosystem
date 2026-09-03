import { sendRequest, getAuthContext, getHmacHeaders, uid } from './client.js';

console.log(`\n============================================================`);
console.log(`       13. INTERNAL MICROSERVICE HMAC REQUEST TESTS         `);
console.log(`============================================================`);

async function run() {
  const { token, workspaceID } = await getAuthContext();

  // 1. Connect WhatsApp channel to get valid Channel ID
  const connectRes = await sendRequest('Setup: Connect WA Channel', 'POST', `/api/v1/channels/whatsapp/connect?workspace_id=${workspaceID}`, {
    token,
    body: {
      display_name: 'Customer Support WA',
    },
  });
  const channelID = connectRes.data?.data?.channel_id;

  // 2. Get Agent ID
  const agentRes = await sendRequest('Setup: Get Agent ID', 'GET', `/api/v1/agent?workspace_id=${workspaceID}`, {
    token,
  });
  const agentID = agentRes.data?.data?.id;

  if (channelID) {
    // 3. Channel Status Callback
    const statusPayload = {
      channel_id: channelID,
      status: 'connected',
      phone_number: '6283893962591',
      session_id: `wa_sess_${channelID}`,
    };
    await sendRequest('13.1 Update Channel Status (whatsmeow bot -> backend)', 'POST', '/internal/v1/channels/status', {
      headers: getHmacHeaders(statusPayload),
      body: statusPayload,
    });

    // 4. Channel Event Callback
    const eventPayload = {
      channel_id: channelID,
      event: 'connected',
      payload: { battery: 98, device: 'Android' },
    };
    await sendRequest('13.2 Push Channel Event (whatsmeow bot -> backend)', 'POST', '/internal/v1/channels/events', {
      headers: getHmacHeaders(eventPayload),
      body: eventPayload,
    });

    // 5. Incoming Message Callback
    const incomingPayload = {
      channel_id: channelID,
      external_user_id: '628123456789@s.whatsapp.net',
      sender_name: 'Customer Service Tester',
      message_id: uid('wam'),
      message_type: 'text',
      content: { text: 'Berapa harga langganan bulanan?' },
      raw_payload: {},
    };
    await sendRequest('13.3 Ingest Incoming WhatsApp Message', 'POST', '/internal/v1/messages/incoming', {
      headers: getHmacHeaders(incomingPayload),
      body: incomingPayload,
    });
  }

  if (agentID) {
    // 6. Agent Health Check
    await sendRequest('13.4 Internal Agent Health Check', 'GET', `/internal/v1/agents/${agentID}/health`, {
      headers: getHmacHeaders(''),
    });

    // 7. Internal Agent Respond
    const respondPayload = {
      conversation_id: '00000000-0000-0000-0000-000000000001',
      message: 'Halo, saya ingin menanyakan produk.',
      visitor_id: 'visitor_123',
    };
    await sendRequest('13.5 Internal Agent Respond Trigger', 'POST', `/internal/v1/agents/${agentID}/respond`, {
      headers: getHmacHeaders(respondPayload),
      body: respondPayload,
    });
  }
}

run();
