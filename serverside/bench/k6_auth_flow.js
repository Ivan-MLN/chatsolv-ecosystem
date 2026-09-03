import http from 'k6/http';
import { check, sleep } from 'k6';

// =====================================================================
// ChatSolv K6 Auth Lifecycle Benchmark (Register -> Login -> Me -> Refresh)
// =====================================================================

const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:3000';

export const options = {
  scenarios: {
    auth_lifecycle: {
      executor: 'shared-iterations',
      vus: 1,
      iterations: parseInt(__ENV.ITERATIONS || '5'),
      maxDuration: '30s',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<400'], // Argon2id password hashing is CPU-bound
  },
};

export default function () {
  const ts = Date.now();
  const rand = Math.floor(Math.random() * 100000);
  const email = `k6_auth_${ts}_${rand}@example.com`;
  const password = 'SecurePassword123!';

  // 1. Register
  const regPayload = JSON.stringify({
    name: 'K6 Auth Tester',
    email: email,
    password: password,
  });
  const regRes = http.post(`${BASE_URL}/api/v1/auth/register`, regPayload, {
    headers: { 'Content-Type': 'application/json' },
  });
  check(regRes, { 'POST /api/v1/auth/register is 201': (r) => r.status === 201 });

  // 2. Login
  const loginPayload = JSON.stringify({ email: email, password: password });
  const loginRes = http.post(`${BASE_URL}/api/v1/auth/login`, loginPayload, {
    headers: { 'Content-Type': 'application/json' },
  });
  check(loginRes, { 'POST /api/v1/auth/login is 200': (r) => r.status === 200 });

  let accessToken = '';
  let refreshToken = '';
  if (loginRes.status === 200) {
    const data = loginRes.json();
    if (data && data.data) {
      accessToken = data.data.access_token;
      refreshToken = data.data.refresh_token;
    }
  }

  // 3. Current User (/api/v1/me)
  if (accessToken) {
    const meRes = http.get(`${BASE_URL}/api/v1/me`, {
      headers: { Authorization: `Bearer ${accessToken}` },
    });
    check(meRes, { 'GET /api/v1/me is 200': (r) => r.status === 200 });
  }

  // 4. Token Refresh
  if (refreshToken) {
    const refreshPayload = JSON.stringify({ refresh_token: refreshToken });
    const refreshRes = http.post(`${BASE_URL}/api/v1/auth/refresh`, refreshPayload, {
      headers: { 'Content-Type': 'application/json' },
    });
    check(refreshRes, { 'POST /api/v1/auth/refresh is 200': (r) => r.status === 200 });
  }

  sleep(0.5);
}
