import http from 'k6/http';
import { check, sleep, group } from 'k6';
import crypto from 'k6/crypto';

// =====================================================================
// ChatSolv K6 Comprehensive Benchmark Suite - All API Routes
// =====================================================================

const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:3000';
const SECRET_KEY = __ENV.INTERNAL_SERVICE_SECRET || 'replace-with-at-least-32-random-bytes';

export const options = {
  scenarios: {
    all_routes: {
      executor: 'shared-iterations',
      vus: parseInt(__ENV.VUS || '5'),
      iterations: parseInt(__ENV.ITERATIONS || '25'),
      maxDuration: '1m',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'], // < 1% error rate on expected requests
    http_req_duration: ['p(95)<50', 'p(99)<150'], // p95 < 50ms, p99 < 150ms
  },
};

// Global Setup: Register a dedicated test user, authenticate, and prepare workspace
export function setup() {
  const timestamp = Date.now();
  const rand = Math.floor(Math.random() * 100000);
  const email = `k6_bench_${timestamp}_${rand}@example.com`;
  const password = 'SecurePassword123!';

  // 1. Register
  const regPayload = JSON.stringify({
    name: 'K6 Bench User',
    email: email,
    password: password,
  });
  http.post(`${BASE_URL}/api/v1/auth/register`, regPayload, {
    headers: { 'Content-Type': 'application/json' },
  });

  // 2. Login
  const loginPayload = JSON.stringify({ email: email, password: password });
  const loginRes = http.post(`${BASE_URL}/api/v1/auth/login`, loginPayload, {
    headers: { 'Content-Type': 'application/json' },
  });

  let token = '';
  let userID = '';
  if (loginRes.status === 200) {
    const resData = loginRes.json();
    if (resData && resData.data) {
      token = resData.data.access_token || '';
    }
  }

  // 3. Create Workspace
  const slug = `k6-ws-${timestamp}-${rand}`;
  const wsPayload = JSON.stringify({
    name: 'K6 Benchmark Workspace',
    slug: slug,
    timezone: 'Asia/Jakarta',
  });
  const wsRes = http.post(`${BASE_URL}/api/v1/workspaces`, wsPayload, {
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
  });

  let workspaceID = '';
  if (wsRes.status === 202) {
    const wsData = wsRes.json();
    if (wsData && wsData.data && wsData.data.workspace) {
      workspaceID = wsData.data.workspace.id;
    }
  }

  // Fallback to fetch workspace from /api/v1/me if not directly returned
  if (!workspaceID && token) {
    for (let retry = 0; retry < 5 && !workspaceID; retry++) {
      const meRes = http.get(`${BASE_URL}/api/v1/me`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (meRes.status === 200) {
        const meData = meRes.json();
        if (meData.data && meData.data.workspaces && meData.data.workspaces.length > 0) {
          workspaceID = meData.data.workspaces[0].workspace_id;
        }
        if (meData.data && meData.data.user) {
          userID = meData.data.user.id;
        }
      }
      if (!workspaceID) sleep(0.1);
    }
  }

  return {
    token: token,
    userID: userID,
    workspaceID: workspaceID,
  };
}

export default function (data) {
  const token = data.token;
  const workspaceID = data.workspaceID;
  const authHeaders = {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${token}`,
  };

  // Helper for HMAC header generation in K6
  function getHmacHeaders(bodyStr) {
    const ts = new Date().toISOString();
    const message = `${ts}.${bodyStr}`;
    const sig = crypto.hmac('sha256', SECRET_KEY, message, 'hex');
    return {
      'Content-Type': 'application/json',
      'X-ChatSolv-Timestamp': ts,
      'X-ChatSolv-Signature': sig,
    };
  }

  // -------------------------------------------------------------------
  // 1. Health & Readiness Endpoints
  // -------------------------------------------------------------------
  group('1. System Health & Readiness', () => {
    const resHealth = http.get(`${BASE_URL}/health`);
    check(resHealth, { 'GET /health is 200': (r) => r.status === 200 });

    const resReady = http.get(`${BASE_URL}/ready`);
    check(resReady, { 'GET /ready is 200': (r) => r.status === 200 });

    const resLive = http.get(`${BASE_URL}/health/live`);
    check(resLive, { 'GET /health/live is 200': (r) => r.status === 200 });

    const resHealthReady = http.get(`${BASE_URL}/health/ready`);
    check(resHealthReady, { 'GET /health/ready is 200': (r) => r.status === 200 });
  });

  // -------------------------------------------------------------------
  // 2. Current User & Dashboard
  // -------------------------------------------------------------------
  group('2. Current User & Dashboard', () => {
    const resMe = http.get(`${BASE_URL}/api/v1/me`, { headers: authHeaders });
    check(resMe, { 'GET /api/v1/me is 200': (r) => r.status === 200 });

    if (workspaceID) {
      const resDash = http.get(`${BASE_URL}/api/v1/dashboard?workspace_id=${workspaceID}`, {
        headers: authHeaders,
      });
      check(resDash, { 'GET /api/v1/dashboard is 200': (r) => r.status === 200 });
    }
  });

  // -------------------------------------------------------------------
  // 3. Workspace Management (Canonical & Scoped)
  // -------------------------------------------------------------------
  group('3. Workspaces', () => {
    if (workspaceID) {
      const resWsCanonical = http.get(`${BASE_URL}/api/v1/workspace?workspace_id=${workspaceID}`, {
        headers: authHeaders,
      });
      check(resWsCanonical, { 'GET /api/v1/workspace is 200': (r) => r.status === 200 });

      const patchWsPayload = JSON.stringify({
        name: 'K6 Bench Workspace Updated',
        timezone: 'Asia/Jakarta',
      });
      const resPatchWs = http.patch(
        `${BASE_URL}/api/v1/workspace?workspace_id=${workspaceID}`,
        patchWsPayload,
        { headers: authHeaders }
      );
      check(resPatchWs, { 'PATCH /api/v1/workspace is 200': (r) => r.status === 200 });

      const resSub = http.get(`${BASE_URL}/api/v1/workspaces/${workspaceID}/subscription`, {
        headers: authHeaders,
      });
      check(resSub, { 'GET /workspaces/:id/subscription is 200': (r) => r.status === 200 });
    }
  });

  // -------------------------------------------------------------------
  // 4. Agent Configuration & Personality
  // -------------------------------------------------------------------
  group('4. Agent Configuration', () => {
    if (workspaceID) {
      const resAgent = http.get(`${BASE_URL}/api/v1/agent?workspace_id=${workspaceID}`, {
        headers: authHeaders,
      });
      check(resAgent, { 'GET /api/v1/agent is 200': (r) => r.status === 200 });

      const patchProfile = JSON.stringify({
        display_name: 'K6 Bench Assistant',
        language: 'id',
        description: 'Virtual Support AI',
        greeting_message: 'Halo, ada yang bisa dibantu?',
        away_message: 'Sedang offline.',
        fallback_message: 'Menghubungkan ke admin.',
      });
      const resProfile = http.patch(
        `${BASE_URL}/api/v1/agent/profile?workspace_id=${workspaceID}`,
        patchProfile,
        { headers: authHeaders }
      );
      check(resProfile, { 'PATCH /api/v1/agent/profile is 200': (r) => r.status === 200 });

      const resGetProfile = http.get(
        `${BASE_URL}/api/v1/agent/profile?workspace_id=${workspaceID}`,
        { headers: authHeaders }
      );
      check(resGetProfile, { 'GET /api/v1/agent/profile is 200': (r) => r.status === 200 });

      const patchPersonality = JSON.stringify({
        bot_name: 'ChatSolv Bot',
        role: 'Customer Support',
        tone: 'friendly',
        communication_style: 'casual_professional',
        primary_language: 'id',
        response_length: 'short',
        emoji_usage: 'minimal',
        greeting_style: 'Halo!',
        closing_style: 'Terima kasih!',
        custom_instructions: 'Selalu ramah dan cepat.',
      });
      const resPersonality = http.patch(
        `${BASE_URL}/api/v1/agent/personality?workspace_id=${workspaceID}`,
        patchPersonality,
        { headers: authHeaders }
      );
      check(resPersonality, { 'PATCH /api/v1/agent/personality is 200': (r) => r.status === 200 });

      const resGetPersonality = http.get(
        `${BASE_URL}/api/v1/agent/personality?workspace_id=${workspaceID}`,
        { headers: authHeaders }
      );
      check(resGetPersonality, { 'GET /api/v1/agent/personality is 200': (r) => r.status === 200 });
    }
  });

  // -------------------------------------------------------------------
  // 5. Business Settings & Policies
  // -------------------------------------------------------------------
  group('5. Business Settings & Policies', () => {
    if (workspaceID) {
      const patchBusiness = JSON.stringify({
        business_name: 'ChatSolv ID',
        industry: 'Software',
        business_description: 'AI platform',
        timezone: 'Asia/Jakarta',
        website: 'https://chatsolv.com',
      });
      const resBusiness = http.patch(
        `${BASE_URL}/api/v1/business?workspace_id=${workspaceID}`,
        patchBusiness,
        { headers: authHeaders }
      );
      check(resBusiness, { 'PATCH /api/v1/business is 200': (r) => r.status === 200 });

      const resGetBusiness = http.get(
        `${BASE_URL}/api/v1/business?workspace_id=${workspaceID}`,
        { headers: authHeaders }
      );
      check(resGetBusiness, { 'GET /api/v1/business is 200': (r) => r.status === 200 });

      const patchPolicies = JSON.stringify({
        shipping_policy: 'Pengiriman reguler 2 hari',
        refund_policy: 'Refund 100%',
        return_policy: 'Retur 7 hari',
        warranty_policy: 'Garansi 1 tahun',
        payment_policy: 'QRIS & Transfer Bank',
        complaint_policy: 'Layanan 24 jam',
      });
      const resPolicies = http.patch(
        `${BASE_URL}/api/v1/settings/workspaces/${workspaceID}/policies`,
        patchPolicies,
        { headers: authHeaders }
      );
      check(resPolicies, { 'PATCH /settings/workspaces/:id/policies is 200': (r) => r.status === 200 });

      const resGetPolicies = http.get(
        `${BASE_URL}/api/v1/settings/workspaces/${workspaceID}/policies`,
        { headers: authHeaders }
      );
      check(resGetPolicies, { 'GET /settings/workspaces/:id/policies is 200': (r) => r.status === 200 });
    }
  });

  // -------------------------------------------------------------------
  // 6. Channels Management
  // -------------------------------------------------------------------
  group('6. Channels', () => {
    if (workspaceID) {
      const resChannels = http.get(`${BASE_URL}/api/v1/channels?workspace_id=${workspaceID}`, {
        headers: authHeaders,
      });
      check(resChannels, { 'GET /api/v1/channels is 200': (r) => r.status === 200 });
    }
  });

  // -------------------------------------------------------------------
  // 7. Knowledge Base
  // -------------------------------------------------------------------
  group('7. Knowledge Base', () => {
    if (workspaceID) {
      const resKnowledge = http.get(`${BASE_URL}/api/v1/knowledge?workspace_id=${workspaceID}`, {
        headers: authHeaders,
      });
      check(resKnowledge, { 'GET /api/v1/knowledge is 200': (r) => r.status === 200 });
    }
  });

  // -------------------------------------------------------------------
  // 8. Conversations & Handoff
  // -------------------------------------------------------------------
  group('8. Conversations', () => {
    if (workspaceID) {
      const resConversations = http.get(
        `${BASE_URL}/api/v1/conversations?workspace_id=${workspaceID}`,
        { headers: authHeaders }
      );
      check(resConversations, { 'GET /api/v1/conversations is 200': (r) => r.status === 200 });
    }
  });

  // -------------------------------------------------------------------
  // 9. Developer API Keys (Create & Revoke Lifecycle)
  // -------------------------------------------------------------------
  group('9. Developer API Keys', () => {
    if (workspaceID) {
      const resApiKeys = http.get(`${BASE_URL}/api/v1/api-keys?workspace_id=${workspaceID}`, {
        headers: authHeaders,
      });
      check(resApiKeys, { 'GET /api/v1/api-keys is 200': (r) => r.status === 200 });

      const keyPayload = JSON.stringify({
        name: `K6 Bench Key ${Date.now()}_${Math.floor(Math.random() * 1000)}`,
        scopes: ['agent:invoke', 'knowledge:read'],
      });
      const resCreateKey = http.post(
        `${BASE_URL}/api/v1/api-keys?workspace_id=${workspaceID}`,
        keyPayload,
        { headers: authHeaders }
      );
      check(resCreateKey, { 'POST /api/v1/api-keys is 201': (r) => r.status === 201 });

      if (resCreateKey.status === 201) {
        const keyData = resCreateKey.json();
        if (keyData.data && keyData.data.api_key) {
          const keyID = keyData.data.api_key.id;
          const resDelKey = http.del(`${BASE_URL}/api/v1/api-keys/${keyID}`, null, {
            headers: authHeaders,
          });
          check(resDelKey, { 'DELETE /api/v1/api-keys/:id is 200': (r) => r.status === 200 });
        }
      }
    }
  });

  // -------------------------------------------------------------------
  // 10. Developer Webhooks
  // -------------------------------------------------------------------
  group('10. Developer Webhooks', () => {
    if (workspaceID) {
      const resWebhooks = http.get(`${BASE_URL}/api/v1/webhooks?workspace_id=${workspaceID}`, {
        headers: authHeaders,
      });
      check(resWebhooks, { 'GET /api/v1/webhooks is 200': (r) => r.status === 200 });
    }
  });

  sleep(0.05);
}
