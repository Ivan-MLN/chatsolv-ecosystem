import http from 'k6/http';
import { check, group, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:3000';

export const options = {
  vus: __ENV.VUS ? parseInt(__ENV.VUS, 10) : 5,
  duration: __ENV.DURATION || '30s',
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<1500'],
  },
};

const params = {
  headers: {
    'Content-Type': 'application/json',
  },
};

export default function () {
  const uniqueId = `${Date.now()}_${__VU}_${__ITER}_${Math.floor(Math.random() * 100000)}`;
  const email = `all_${uniqueId}@example.com`;
  const password = 'Password123!';

  group('Health & Ready Check', () => {
    const resHealth = http.get(`${BASE_URL}/health`);
    check(resHealth, { 'health status 200': (r) => r.status === 200 });

    const resReady = http.get(`${BASE_URL}/ready`);
    check(resReady, { 'ready status 200': (r) => r.status === 200 });
  });

  group('Auth Register Flow', () => {
    const resReg = http.post(
      `${BASE_URL}/api/v1/auth/register`,
      JSON.stringify({ name: `User ${uniqueId}`, email, password }),
      params
    );
    check(resReg, {
      'register 201 or 429': (r) => r.status === 201 || r.status === 429,
    });
  });

  group('Auth Login Flow', () => {
    const resLogin = http.post(
      `${BASE_URL}/api/v1/auth/login`,
      JSON.stringify({ email, password }),
      params
    );
    check(resLogin, {
      'login 200 or 429': (r) => r.status === 200 || r.status === 429,
    });
  });

  group('Forgot Password Flow', () => {
    const resForgot = http.post(
      `${BASE_URL}/api/v1/auth/forgot-password`,
      JSON.stringify({ email }),
      params
    );
    check(resForgot, {
      'forgot password 200 or 429': (r) => r.status === 200 || r.status === 429,
    });
  });

  group('Reset Password Flow (Invalid Token Check)', () => {
    const resReset = http.post(
      `${BASE_URL}/api/v1/auth/reset-password`,
      JSON.stringify({ token: 'invalid_token_sample', new_password: 'NewPassword123!' }),
      params
    );
    check(resReset, {
      'reset password 400 or 429': (r) => r.status === 400 || r.status === 429,
    });
  });

  sleep(1);
}
