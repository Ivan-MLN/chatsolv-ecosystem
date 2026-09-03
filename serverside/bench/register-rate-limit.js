import http from 'k6/http';
import { check } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:3000';
const RATE_LIMIT_MAX = __ENV.RATE_LIMIT_MAX ? parseInt(__ENV.RATE_LIMIT_MAX, 10) : 10;

if (!Number.isInteger(RATE_LIMIT_MAX) || RATE_LIMIT_MAX < 1) {
  throw new Error('RATE_LIMIT_MAX must be a positive integer');
}

export const options = {
  scenarios: {
    register_rate_limit: {
      executor: 'shared-iterations',
      vus: 1,
      iterations: RATE_LIMIT_MAX + 1,
      maxDuration: '30s',
    },
  },
  thresholds: {
    checks: ['rate==1'],
    http_req_failed: ['rate==0'],
  },
};

const params = {
  headers: {
    'Content-Type': 'application/json',
  },
  responseCallback: http.expectedStatuses(201, 429),
};

export function setup() {
  const probe = http.get(`${BASE_URL}/ready`);
  if (probe.status !== 200) {
    throw new Error(`API is not ready: GET /ready returned ${probe.status}`);
  }
}

export default function () {
  const requestNumber = __ITER + 1;
  const uniqueId = `${Date.now()}_${__VU}_${__ITER}_${Math.floor(Math.random() * 100000)}`;
  const payload = JSON.stringify({
    name: `Rate Limit User ${uniqueId}`,
    email: `bench_rate_limit_${uniqueId}@example.com`,
    password: 'Password123!',
  });

  const res = http.post(`${BASE_URL}/api/v1/auth/register`, payload, params);
  const expectedStatus = requestNumber <= RATE_LIMIT_MAX ? 201 : 429;

  check(res, {
    [`request ${requestNumber} returns ${expectedStatus}`]: (r) => r.status === expectedStatus,
    'response has expected envelope': (r) => {
      try {
        const body = JSON.parse(r.body);
        if (expectedStatus === 201) {
          return body.success === true && Boolean(body.data?.id);
        }
        return body.success === false && body.error?.code === 'RATE_LIMITED';
      } catch (_) {
        return false;
      }
    },
  });
}
