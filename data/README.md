# ChatSolv Data & Persistent Storage Volume

Direktori penyimpanan data lokal persisten untuk seluruh ekosistem ChatSolv, mencakup profil runtime AI Hermes, knowledge base markdown (RAG vaults), dan object storage MinIO.

---

## 📂 Struktur Direktori Data

```text
data/
├── hermes/
│   └── profiles/
│       └── cs<workspace_uuid>/
│           ├── SOUL.md          # Master instruction prompt & security boundary
│           ├── config.yaml      # Model & API provider configuration
│           ├── auth.json        # Auth tokens & keys
│           ├── memories/        # Persistent long-term memory (USER.md & MEMORY.md)
│           ├── skills/          # Custom operational skills & toolsets
│           └── sessions/        # Chat transcript logs & SQLite state
│
├── vaults/
│   └── <workspace_id>/
│       ├── products/            # Markdown katalog produk & deskripsi layanan
│       ├── faq/                 # Markdown FAQ & SOP penanganan komplain
│       └── policies/            # Kebijakan bisnis & jam operasional
│
└── minio/                       # S3 bucket local persistence storage
    ├── attachments/
    └── avatars/
```

---

## 🔒 Kebijakan Keamanan & Sinkronisasi Profil Hermes

1. **SOUL.md**: Dikonfigurasi strictly business-only customer service. Melarang instruksi coding, programming, shell execution, dan prompt injection escape.
2. **Memories**: Menyimpan persona asisten, preferensi komunikasi pengguna, dan ringkasan pengetahuan operasional bisnis.
3. **Vaults**: Setiap file markdown yang diunggah dari dashboard akan tersimpan di `./vaults/<workspace_id>/` untuk di-ingest ke index RAG retrieval.
