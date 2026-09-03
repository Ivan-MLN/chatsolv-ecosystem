import { sendRequest, getAuthContext, uid } from './client.js';

console.log(`\n============================================================`);
console.log(`          4. WORKSPACE MANAGEMENT REQUEST TESTS             `);
console.log(`============================================================`);

async function run() {
  const { token, workspaceID } = await getAuthContext();

  // 1. Create Workspace
  const slug = uid('ws');
  const createWsRes = await sendRequest('4.1 Create New Workspace', 'POST', '/api/v1/workspaces', {
    token,
    body: {
      name: 'ChatSolv Jakarta Office',
      slug,
      timezone: 'Asia/Jakarta',
    },
  });
  const wsID = createWsRes.data?.data?.workspace?.id || workspaceID;

  // 2. Get Workspace by ID
  await sendRequest('4.2 Get Workspace Detail by ID', 'GET', `/api/v1/workspaces/${wsID}`, {
    token,
  });

  // 3. Patch Workspace by ID
  await sendRequest('4.3 Update Workspace by ID', 'PATCH', `/api/v1/workspaces/${wsID}`, {
    token,
    body: {
      name: 'ChatSolv Jakarta Office Updated',
      timezone: 'Asia/Jakarta',
    },
  });

  // 4. Get Subscription
  await sendRequest('4.4 Get Workspace Subscription & Entitlements', 'GET', `/api/v1/workspaces/${wsID}/subscription`, {
    token,
  });

  // 5. Canonical Get Workspace
  await sendRequest('4.5 Canonical Get Workspace', 'GET', `/api/v1/workspace?workspace_id=${wsID}`, {
    token,
  });

  // 6. Canonical Update Workspace
  await sendRequest('4.6 Canonical Update Workspace', 'PATCH', `/api/v1/workspace?workspace_id=${wsID}`, {
    token,
    body: {
      name: 'ChatSolv Jakarta Canonical',
    },
  });
}

run();
