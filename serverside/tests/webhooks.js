import { sendRequest, getAuthContext } from './client.js';

console.log(`\n============================================================`);
console.log(`           11. DEVELOPER WEBHOOKS REQUEST TESTS             `);
console.log(`============================================================`);

async function run() {
  const { token, workspaceID } = await getAuthContext();

  // 1. List Webhooks
  await sendRequest('11.1 List Workspace Webhook Subscriptions', 'GET', `/api/v1/webhooks?workspace_id=${workspaceID}`, {
    token,
  });

  // 2. Create Webhook Subscription
  const createWhRes = await sendRequest('11.2 Create Webhook Endpoint', 'POST', `/api/v1/webhooks?workspace_id=${workspaceID}`, {
    token,
    body: {
      url: 'https://api.mycrm.com/chatsolv/events',
      events: ['conversation.created', 'message.received', 'handoff.requested'],
      description: 'Internal CRM sync endpoint',
    },
  });

  const webhookID = createWhRes.data?.data?.webhook?.id;

  if (webhookID) {
    // 3. Update Webhook
    await sendRequest('11.3 Update Webhook Endpoint URL and Events', 'PATCH', `/api/v1/webhooks/${webhookID}`, {
      token,
      body: {
        url: 'https://api.mycrm.com/chatsolv/v2/events',
        events: ['conversation.created', 'message.created', 'agent.error'],
        status: 'active',
      },
    });

    // 4. Delete Webhook
    await sendRequest('11.4 Delete Webhook Subscription', 'DELETE', `/api/v1/webhooks/${webhookID}`, {
      token,
    });
  }
}

run();
