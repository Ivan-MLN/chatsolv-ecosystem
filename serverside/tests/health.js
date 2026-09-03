import { sendRequest } from './client.js';

console.log(`\n============================================================`);
console.log(`        1. SYSTEM HEALTH & READINESS REQUEST TESTS          `);
console.log(`============================================================`);

async function run() {
  await sendRequest('1.1 Fast Liveness Check', 'GET', '/health');
  await sendRequest('1.2 Full Dependency Readiness Check', 'GET', '/ready');
  await sendRequest('1.3 Health Alias /health/live', 'GET', '/health/live');
  await sendRequest('1.4 Health Alias /health/ready', 'GET', '/health/ready');
}

run();
