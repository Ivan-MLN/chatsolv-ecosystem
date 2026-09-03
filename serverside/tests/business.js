import { sendRequest, getAuthContext } from './client.js';

console.log(`\n============================================================`);
console.log(`      6. BUSINESS SETTINGS & POLICIES REQUEST TESTS         `);
console.log(`============================================================`);

async function run() {
  const { token, workspaceID } = await getAuthContext();

  // 1. Canonical Update Business Settings
  await sendRequest('6.1 Canonical Update Business Settings', 'PATCH', `/api/v1/business?workspace_id=${workspaceID}`, {
    token,
    body: {
      business_name: 'ChatSolv Official Store',
      industry: 'Software & Technology',
      business_description: 'Platform AI perpesanan pintar.',
      website: 'https://chatsolv.com',
      timezone: 'Asia/Jakarta',
    },
  });

  // 2. Canonical Get Business Settings
  await sendRequest('6.2 Canonical Get Business Settings', 'GET', `/api/v1/business?workspace_id=${workspaceID}`, {
    token,
  });

  // 3. Update Business Policies
  await sendRequest('6.3 Update Workspace Business Policies', 'PATCH', `/api/v1/settings/workspaces/${workspaceID}/policies`, {
    token,
    body: {
      shipping_policy: 'Pengiriman setiap hari kerja via SiCepat dan JNE.',
      refund_policy: 'Refund 100% jika produk mengalami kerusakan.',
      return_policy: 'Retur maksimal 7 hari setelah barang diterima.',
      warranty_policy: 'Garansi resmi 1 tahun.',
      payment_policy: 'Menerima pembayaran QRIS, Transfer Bank, dan Kartu Kredit.',
      complaint_policy: 'Layanan keluhan pelanggan aktif 24 jam.',
    },
  });

  // 4. Get Business Policies
  await sendRequest('6.4 Get Workspace Business Policies', 'GET', `/api/v1/settings/workspaces/${workspaceID}/policies`, {
    token,
  });
}

run();
