import http from 'k6/http';
import { check, sleep } from 'k6';

// =====================================================================
// ChatSolv K6 Stress / Load Test (Ramp-up -> Steady -> Ramp-down)
// =====================================================================

const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:3000';

export const options = {
  stages: [
    { duration: '5s', target: 10 },  // Ramp up to 10 VUs
    { duration: '10s', target: 25 }, // Steady load at 25 VUs
    { duration: '5s', target: 0 },   // Ramp down to 0
  ],
  thresholds: {
    http_req_failed: ['rate<0.01'], // 99% requests must succeed
    http_req_duration: ['p(95)<50', 'p(99)<100'], // p95 < 50ms, p99 < 100ms
  },
};

export default function () {
  // Liveness & Readiness checks under load
  const resHealth = http.get(`${BASE_URL}/health`);
  check(resHealth, { 'GET /health is 200': (r) => r.status === 200 });

  const resReady = http.get(`${BASE_URL}/ready`);
  check(resReady, { 'GET /ready is 200': (r) => r.status === 200 });

  const resLive = http.get(`${BASE_URL}/health/live`);
  check(resLive, { 'GET /health/live is 200': (r) => r.status === 200 });

  sleep(0.05);
}
