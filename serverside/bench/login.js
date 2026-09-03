import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:3000';

export const options = {
  vus: __ENV.VUS ? parseInt(__ENV.VUS, 10) : 5,
  duration: __ENV.DURATION || '30s',
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<1500'], // Argon2 verification
  },
};

const params = {
  headers: {
    'Content-Type': 'application/json',
  },
};

// Setup data test user terlebih dahulu sebelum load test dijalankan
export function setup() {
  const uniqueId = `setup_${Date.now()}`;
  const email = `bench_login_${uniqueId}@example.com`;
  const password = 'Password123!';

  const registerPayload = JSON.stringify({
    name: 'Bench Login User',
    email: email,
    password: password,
  });

  const res = http.post(`${BASE_URL}/api/v1/auth/register`, registerPayload, params);
  check(res, {
    'setup user registered successfully': (r) => r.status === 201,
  });

  return { email, password };
}

export default function (data) {
  const payload = JSON.stringify({
    email: data.email,
    password: data.password,
  });

  const res = http.post(`${BASE_URL}/api/v1/auth/login`, payload, params);

  check(res, {
    'login status is 200 or 429': (r) => r.status === 200 || r.status === 429,
    'login tokens received': (r) => {
      if (r.status === 200) {
        try {
          const body = JSON.parse(r.body);
          return body.success === true &&
                 Boolean(body.data?.access_token) &&
                 Boolean(body.data?.refresh_token) &&
                 body.data?.token_type === 'Bearer';
        } catch (_) {
          return false;
        }
      }
      return true;
    },
  });

  sleep(1);
}
