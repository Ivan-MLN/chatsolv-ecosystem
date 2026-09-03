import { sendRequest, getAuthContext } from './client.js';

console.log(`\n============================================================`);
console.log(`      5. AGENT CONFIGURATION & PERSONALITY REQUEST TESTS    `);
console.log(`============================================================`);

async function run() {
  const { token, workspaceID } = await getAuthContext();

  // 1. Canonical Get Agent
  const agentRes = await sendRequest('5.1 Canonical Get Agent Details', 'GET', `/api/v1/agent?workspace_id=${workspaceID}`, {
    token,
  });
  const agentID = agentRes.data?.data?.id;

  // 2. Canonical Update Agent Profile
  await sendRequest('5.2 Canonical Update Agent Profile', 'PATCH', `/api/v1/agent/profile?workspace_id=${workspaceID}`, {
    token,
    body: {
      display_name: 'ChatSolv AI Assistant',
      language: 'id',
      description: 'Asisten AI Resmi ChatSolv',
      greeting_message: 'Halo! Ada yang bisa kami bantu?',
      away_message: 'Kami sedang offline saat ini.',
      fallback_message: 'Mohon tunggu sebentar, saya teruskan ke tim CS kami.',
    },
  });

  // 3. Canonical Get Agent Profile
  await sendRequest('5.3 Canonical Get Agent Profile', 'GET', `/api/v1/agent/profile?workspace_id=${workspaceID}`, {
    token,
  });

  // 4. Canonical Update Agent Personality
  await sendRequest('5.4 Canonical Update Agent Personality', 'PATCH', `/api/v1/agent/personality?workspace_id=${workspaceID}`, {
    token,
    body: {
      bot_name: 'ChatSolv Bot',
      role: 'Customer Support',
      tone: 'friendly',
      communication_style: 'casual_professional',
      primary_language: 'id',
      response_length: 'short',
      emoji_usage: 'minimal',
      greeting_style: 'Halo! Selamat datang.',
      closing_style: 'Terima kasih telah menghubungi kami.',
      custom_instructions: 'Selalu gunakan bahasa yang sopan dan jelas.',
    },
  });

  // 5. Canonical Get Agent Personality
  await sendRequest('5.5 Canonical Get Agent Personality', 'GET', `/api/v1/agent/personality?workspace_id=${workspaceID}`, {
    token,
  });

  // 6. Direct Scoped Agent Profile & Personality (if agentID exists)
  if (agentID) {
    await sendRequest('5.6 Scoped Get Agent Profile', 'GET', `/api/v1/agents/${agentID}/profile`, {
      token,
    });
    await sendRequest('5.7 Scoped Get Agent Personality', 'GET', `/api/v1/agents/${agentID}/personality`, {
      token,
    });
  }

  // 7. Agent Playground Test
  await sendRequest('5.8 Agent Playground Test', 'POST', `/api/v1/agent/test?workspace_id=${workspaceID}`, {
    token,
    body: {
      message: 'Halo, apakah ada promo hari ini?',
    },
  });
}

run();
