import { createServer } from 'node:http';
import { createHmac } from 'node:crypto';

const port = Number(process.env.PORT || 4010);
const backend = process.env.BACKEND_INTERNAL_URL || 'http://localhost:3000/internal/v1';
const secret = process.env.INTERNAL_SERVICE_SECRET || 'replace-with-at-least-32-random-bytes';
const sessions = new Map();

function json(res, status, value) {
  const body = JSON.stringify(value);
  res.writeHead(status, {'content-type': 'application/json', 'content-length': Buffer.byteLength(body)});
  res.end(body);
}

async function body(req) {
  const chunks = [];
  for await (const chunk of req) chunks.push(chunk);
  return Buffer.concat(chunks).toString('utf8');
}

async function signedPost(path, payload) {
  const raw = JSON.stringify(payload);
  const timestamp = new Date().toISOString();
  const signature = createHmac('sha256', secret).update(`${timestamp}.${raw}`).digest('hex');
  return fetch(`${backend}${path}`, {method: 'POST', headers: {'content-type': 'application/json', 'x-chatsolv-timestamp': timestamp, 'x-chatsolv-signature': signature}, body: raw});
}

createServer(async (req, res) => {
  if (req.method === 'GET' && req.url === '/health') return json(res, 200, {status: 'ok'});
  if (req.method === 'POST' && req.url === '/internal/v1/channels/connect') {
    const input = JSON.parse(await body(req));
    const sessionId = `mock_${crypto.randomUUID()}`;
    sessions.set(sessionId, input.channel_id);
    return json(res, 201, {data: {session_id: sessionId, status: 'waiting_pairing', qr: `mock://pair/${sessionId}`}});
  }
  if (req.method === 'POST' && req.url === '/mock/pair') {
    const input = JSON.parse(await body(req));
    const channelId = sessions.get(input.session_id);
    if (!channelId) return json(res, 404, {error: 'session_not_found'});
    await signedPost('/channels/status', {channel_id: channelId, status: 'connected', phone_number: input.phone_number || '628000000000'});
    return json(res, 200, {status: 'connected'});
  }
  if (req.method === 'POST' && req.url === '/mock/incoming') {
    const input = JSON.parse(await body(req));
    const response = await signedPost('/messages/incoming', input);
    return json(res, response.status, await response.json());
  }
  return json(res, 404, {error: 'not_found'});
}).listen(port, '127.0.0.1', () => console.log(`mock WhatsApp service listening on http://127.0.0.1:${port}`));
