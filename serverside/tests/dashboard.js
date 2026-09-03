import { sendRequest, getAuthContext } from './client.js';

console.log(`\n============================================================`);
console.log(`         3. USER & DASHBOARD OVERVIEW REQUEST TESTS         `);
console.log(`============================================================`);

async function run() {
  const { token, workspaceID } = await getAuthContext();

  // 1. Current User
  await sendRequest('3.1 Get Current User & Workspaces', 'GET', '/api/v1/me', {
    token,
  });

  // 2. Dashboard Overview
  await sendRequest('3.2 Get Dashboard Workspace Overview', 'GET', `/api/v1/dashboard?workspace_id=${workspaceID}`, {
    token,
  });
}

run();
