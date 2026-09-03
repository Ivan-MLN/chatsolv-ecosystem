import { sendRequest, uid } from './client.js';

console.log(`\n============================================================`);
console.log(`            2. AUTHENTICATION FLOW REQUEST TESTS            `);
console.log(`============================================================`);

async function run() {
  const email = `${uid('auth')}@example.com`;
  const password = 'Password123!';

  // 1. Register
  await sendRequest('2.1 Register New User', 'POST', '/api/v1/auth/register', {
    body: {
      name: 'Budi Santoso',
      email,
      password,
    },
  });

  // 2. Login
  const loginRes = await sendRequest('2.2 Login User', 'POST', '/api/v1/auth/login', {
    body: {
      email,
      password,
    },
  });

  const refreshToken = loginRes.data?.data?.refresh_token;

  // 3. Refresh Token
  if (refreshToken) {
    await sendRequest('2.3 Rotate Access Token via Refresh Token', 'POST', '/api/v1/auth/refresh', {
      body: {
        refresh_token: refreshToken,
      },
    });
  }

  // 4. Forgot Password
  await sendRequest('2.4 Request Password Reset', 'POST', '/api/v1/auth/forgot-password', {
    body: {
      email,
    },
  });
}

run();
