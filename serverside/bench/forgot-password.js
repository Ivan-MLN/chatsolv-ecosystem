import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:3000';

export const options = {
  vus: __ENV.VUS ? parseInt(__ENV.VUS, 10) : 5,
  duration: __ENV.DURATION || '30s',
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<500'],
  },
};

const params = {
  headers: {
    'Content-Type': 'application/json',
  },
};

export default function () {
  const payload = JSON.stringify({
    email: 'bench_forgot@example.com',
  });

  const res = http.post(`${BASE_URL}/api/v1/auth/forgot-password`, payload, params);

  check(res, {
    'forgot-password status is 200 or 429': (r) => r.status === 200 || r.status === 429,
    'forgot-password success response': (r) => {
      if (r.status === 200) {
        try {
          const body = JSON.parse(r.body);
          return body.success === true &&
                 body.message === 'If the account exists, password reset instructions have been sent';
        } catch (_) {
          return false;
        }
      }
      return true;
    },
  });

  sleep(1);
}
