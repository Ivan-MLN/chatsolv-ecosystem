import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:3000';

export const options = {
  vus: __ENV.VUS ? parseInt(__ENV.VUS, 10) : 10,
  duration: __ENV.DURATION || '30s',
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<200'],
  },
};

export default function () {
  // Test endpoint /health
  const resHealth = http.get(`${BASE_URL}/health`);
  check(resHealth, {
    'health status is 200': (r) => r.status === 200,
    'health status ok': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.status === 'ok';
      } catch (_) {
        return false;
      }
    },
  });

  // Test endpoint /ready
  const resReady = http.get(`${BASE_URL}/ready`);
  check(resReady, {
    'ready status is 200': (r) => r.status === 200,
    'ready status ready': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.status === 'ready';
      } catch (_) {
        return false;
      }
    },
  });

  sleep(1);
}
