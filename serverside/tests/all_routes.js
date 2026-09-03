import { execSync } from 'node:child_process';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const scripts = [
  'health.js',
  'auth.js',
  'dashboard.js',
  'workspaces.js',
  'agents.js',
  'business.js',
  'channels.js',
  'knowledge.js',
  'conversations.js',
  'api_keys.js',
  'webhooks.js',
  'public_sessions.js',
  'internal_hmac.js',
];

console.log(`\n============================================================`);
console.log(`🚀 RUNNING ALL CHATSOLV ROUTE REQUEST TESTS (node:fetch)   `);
console.log(`============================================================\n`);

let passed = 0;
let failed = 0;

for (const script of scripts) {
  const scriptPath = path.join(__dirname, script);
  try {
    execSync(`node "${scriptPath}"`, { stdio: 'inherit' });
    passed++;
  } catch (err) {
    console.error(`\n❌ Error running ${script}:`, err.message);
    failed++;
  }
}

console.log(`\n============================================================`);
console.log(`                  ALL TESTS SUMMARY                         `);
console.log(`============================================================`);
console.log(`  Total Modules : ${scripts.length}`);
console.log(`  Passed        : ${passed}`);
console.log(`  Failed        : ${failed}`);
console.log(`============================================================\n`);
