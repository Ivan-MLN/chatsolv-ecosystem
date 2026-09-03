import { sendRequest, getAuthContext } from './client.js';

console.log(`\n============================================================`);
console.log(`         7. CHANNELS & WHATSAPP REQUEST TESTS               `);
console.log(`============================================================`);

async function run() {
  const { token, workspaceID } = await getAuthContext();

  // 1. List Channels
  const listRes = await sendRequest('7.1 List Workspace Channels', 'GET', `/api/v1/channels?workspace_id=${workspaceID}`, {
    token,
  });

  // 2. Connect WhatsApp (Pairing Request)
  const connectRes = await sendRequest('7.2 Initiate WhatsApp Pairing', 'POST', `/api/v1/channels/whatsapp/connect?workspace_id=${workspaceID}`, {
    token,
    body: {
      display_name: 'Customer Support WA',
    },
  });

  const channelID = connectRes.data?.data?.channel_id;

  // 3. Delete / Disconnect Channel
  if (channelID) {
    await sendRequest('7.3 Delete Channel Connection', 'DELETE', `/api/v1/channels/${channelID}`, {
      token,
    });
  }
}

run();
