import crypto from 'node:crypto';

export const BASE_URL = process.env.BASE_URL || 'http://127.0.0.1:3000';
export const SECRET_KEY = process.env.INTERNAL_SERVICE_SECRET || 'replace-with-at-least-32-random-bytes';

/**
 * Perform an HTTP request using node:fetch
 */
export async function apiRequest(method, path, options = {}) {
  const url = `${BASE_URL}${path}`;
  const headers = {
    'Content-Type': 'application/json',
    ...(options.token ? { Authorization: `Bearer ${options.token}` } : {}),
    ...(options.headers || {}),
  };

  const config = {
    method: method.toUpperCase(),
    headers,
  };

  if (options.body && method.toUpperCase() !== 'GET' && method.toUpperCase() !== 'HEAD') {
    config.body = typeof options.body === 'string' ? options.body : JSON.stringify(options.body);
  }

  const response = await fetch(url, config);
  let data = null;
  const contentType = response.headers.get('content-type') || '';
  if (contentType.includes('application/json')) {
    data = await response.json().catch(() => null);
  } else {
    data = await response.text().catch(() => '');
  }

  return {
    status: response.status,
    headers: response.headers,
    data,
  };
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
 * Generate unique random string for isolation
 */
export function randomId(prefix = 'test') {
  return `${prefix}_${Date.now()}_${Math.floor(Math.random() * 100000)}`;
}
