import { sendRequest, getAuthContext } from './client.js';

console.log(`\n============================================================`);
console.log(`         8. KNOWLEDGE BASE MANAGEMENT REQUEST TESTS         `);
console.log(`============================================================`);

async function run() {
  const { token, workspaceID } = await getAuthContext();

  // 1. List Knowledge Sources
  await sendRequest('8.1 List Knowledge Sources', 'GET', `/api/v1/knowledge?workspace_id=${workspaceID}`, {
    token,
  });

  // 2. Create Text Knowledge
  await sendRequest('8.2 Create Text Knowledge Source', 'POST', `/api/v1/knowledge/text?workspace_id=${workspaceID}`, {
    token,
    body: {
      title: 'Kebijakan Pengembalian Dana',
      content: 'Pengembalian dana diproses dalam 1-3 hari kerja setelah barang diterima oleh tim gudang.',
      category: 'faq',
    },
  });

  // 3. Create FAQ Knowledge
  await sendRequest('8.3 Create FAQ Knowledge Source', 'POST', `/api/v1/knowledge/faqs?workspace_id=${workspaceID}`, {
    token,
    body: {
      items: [
        {
          question: 'Bagaimana cara melacak pesanan saya?',
          answer: 'Anda dapat melacak pesanan melalui nomor resi yang dikirimkan ke email atau WhatsApp.',
        },
        {
          question: 'Apakah bisa bayar di tempat (COD)?',
          answer: 'Ya, layanan COD tersedia untuk wilayah Jabodetabek.',
        },
      ],
    },
  });
}

run();
