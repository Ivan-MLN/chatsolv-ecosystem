import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:3000';

export const options = {
  vus: __ENV.VUS ? parseInt(__ENV.VUS, 10) : 5,
  duration: __ENV.DURATION || '30s',
  thresholds: {
    checks: ['rate==1'],
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<1500'], // Argon2 hashing membutuhkan komputasi CPU lebih tinggi
  },
};

const params = {
  headers: {
    'Content-Type': 'application/json',
  },
  responseCallback: http.expectedStatuses(201),
};

export default function () {
  // Generate random email unik per iterasi agar tidak duplicate conflict (409)
  const uniqueId = `${Date.now()}_${__VU}_${__ITER}_${Math.floor(Math.random() * 100000)}`;
  const payload = JSON.stringify({
    name: `User ${uniqueId}`,
    email: `bench_${uniqueId}@example.com`,
    password: 'Password123!',
  });

  const res = http.post(`${BASE_URL}/api/v1/auth/register`, payload, params);

  check(res, {
    'register status is 201': (r) => r.status === 201,
    'register success response structure': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.success === true && Boolean(body.data?.id) && body.data?.email === `bench_${uniqueId}@example.com`;
      } catch (_) {
        return false;
      }
    },
  });

  sleep(1);
}
