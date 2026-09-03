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
  // Menguji validasi & respons endpoint reset password dengan token dummy (ekspektasi 400 INVALID_RESET_TOKEN atau 429 jika rate limited)
  const payload = JSON.stringify({
    token: 'dummy_or_invalid_reset_token_benchmark',
    new_password: 'NewSecurePassword123!',
  });

  const res = http.post(`${BASE_URL}/api/v1/auth/reset-password`, payload, params);

  check(res, {
    'reset-password status is 400 or 429': (r) => r.status === 400 || r.status === 429,
    'reset-password invalid token error code': (r) => {
      if (r.status === 400) {
        try {
          const body = JSON.parse(r.body);
          return body.success === false && body.error?.code === 'INVALID_RESET_TOKEN';
        } catch (_) {
          return false;
        }
      }
      return true;
    },
  });

  sleep(1);
}
