import { sendRequest, getAuthContext } from './client.js';

console.log(`\n============================================================`);
console.log(`         10. DEVELOPER API KEYS REQUEST TESTS               `);
console.log(`============================================================`);

async function run() {
  const { token, workspaceID } = await getAuthContext();

  // 1. List API Keys (Initially Empty)
  await sendRequest('10.1 List Developer API Keys', 'GET', `/api/v1/api-keys?workspace_id=${workspaceID}`, {
    token,
  });

  // 2. Create API Key
  const createKeyRes = await sendRequest('10.2 Create New Secret API Key', 'POST', `/api/v1/api-keys?workspace_id=${workspaceID}`, {
    token,
    body: {
      name: 'Production Website Widget Key',
      role: 'member',
      scopes: ['conversation:read', 'conversation:write'],
    },
  });

  const keyID = createKeyRes.data?.data?.key?.id;

  // 3. Delete / Revoke API Key
  if (keyID) {
    await sendRequest('10.3 Revoke API Key by ID', 'DELETE', `/api/v1/api-keys/${keyID}`, {
      token,
    });
  }
}

run();
