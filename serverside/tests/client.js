import crypto from 'node:crypto';

export const BASE_URL = process.env.BASE_URL || 'http://127.0.0.1:3000';
export const SECRET_KEY = process.env.INTERNAL_SERVICE_SECRET || 'replace-with-at-least-32-random-bytes';

/**
 * Pretty logger for API test requests and responses
 */
export async function sendRequest(label, method, path, options = {}) {
  const url = `${BASE_URL}${path}`;
  const headers = {
    'Content-Type': 'application/json',
    ...(options.token ? { Authorization: `Bearer ${options.token}` } : {}),
    ...(options.headers || {}),
  };

  console.log(`\n------------------------------------------------------------`);
  console.log(`▶ ${label}`);
  console.log(`  ${method.toUpperCase()} ${url}`);
  if (options.token) {
    console.log(`  Authorization: Bearer ${options.token.slice(0, 16)}...`);
  }
  if (options.body) {
    console.log(`  Request Body:`);
    console.log(`  ${JSON.stringify(options.body, null, 2).replace(/\n/g, '\n  ')}`);
  }

  const config = {
    method: method.toUpperCase(),
    headers,
  };

  if (options.body && method.toUpperCase() !== 'GET' && method.toUpperCase() !== 'HEAD') {
    config.body = typeof options.body === 'string' ? options.body : JSON.stringify(options.body);
  }

  const startTime = Date.now();
  try {
    const response = await fetch(url, config);
    const duration = Date.now() - startTime;

    let data = null;
    const contentType = response.headers.get('content-type') || '';
    if (contentType.includes('application/json')) {
      data = await response.json().catch(() => null);
    } else {
      data = await response.text().catch(() => '');
    }

    const statusBadge = response.status >= 200 && response.status < 300 ? `✅ HTTP ${response.status}` : `⚠️ HTTP ${response.status}`;
    console.log(`\n  Response [${statusBadge}] (${duration}ms):`);
    if (typeof data === 'object' && data !== null) {
      console.log(`  ${JSON.stringify(data, null, 2).replace(/\n/g, '\n  ')}`);
    } else {
      console.log(`  ${data}`);
    }

    return {
      status: response.status,
      headers: response.headers,
      data,
    };
  } catch (err) {
    console.log(`\n  ❌ Request Failed: ${err.message}`);
    return {
      status: 0,
      error: err.message,
    };
  }
}

/**
 * Generate HMAC SHA-256 headers for internal microservice routes
 */
export function getHmacHeaders(bodyObj = {}) {
  const timestamp = new Date().toISOString();
  const bodyStr = typeof bodyObj === 'string' ? bodyObj : JSON.stringify(bodyObj);
  const message = `${timestamp}.${bodyStr}`;
  const signature = crypto.createHmac('sha256', SECRET_KEY).update(message).digest('hex');

  return {
    'Content-Type': 'application/json',
    'X-ChatSolv-Timestamp': timestamp,
    'X-ChatSolv-Signature': signature,
  };
}

/**
 * Unique ID generator
 */
export function uid(prefix = 'test') {
  return `${prefix}-${Date.now()}-${Math.floor(Math.random() * 10000)}`.toLowerCase().replace(/_/g, '-');
}

/**
 * Helper to authenticate and get ready token & workspaceID
 */
export async function getAuthContext() {
  const email = `${uid('user')}@example.com`;
  const password = 'Password123!';

  // 1. Register
  await sendRequest('Setup: Register Account', 'POST', '/api/v1/auth/register', {
    body: { name: 'Auto Test User', email, password },
  });

  // 2. Login
  const loginRes = await sendRequest('Setup: Login', 'POST', '/api/v1/auth/login', {
    body: { email, password },
  });
  const token = loginRes.data?.data?.access_token || '';
  const userID = loginRes.data?.data?.user?.id || '';

  // 3. Create Workspace
  const slug = uid('ws');
  const wsRes = await sendRequest('Setup: Create Workspace', 'POST', '/api/v1/workspaces', {
    token,
    body: { name: 'Node Fetch Workspace', slug, timezone: 'Asia/Jakarta' },
  });
  const workspaceID = wsRes.data?.data?.workspace?.id || '';

  return { token, userID, workspaceID, email, password };
}
