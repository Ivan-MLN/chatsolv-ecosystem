# Cypher Endpoint Attack Log + VPS Report (Recurring)

User meminta "kirim lagi laporan grafik vps dan log endpoint serangan cypher" —
dua laporan dikirim ke grup WhatsApp RecoVerse Team
(`120363186235853203@g.us`). Resep di bawah validasi berulang.

## Laporan 1 — VPS Dashboard (instant via MCP)

Jangan buat HTML sendiri; pakai tool `mcp__baileys__wa_sys_monitor` dengan
`jid` grup. Tool ini render dashboard live (CPU, RAM, disk, load, top proc,
uptime, IP publik) dan langsung kirim gambarnya ke chat. Output JSON-nya juga
bisa dipakai buat narasi teks pendamping. Verifikasi balasan `success: true`
dan cek field `diskPct` (user peduli disk — 80%+ = catatan penting).

**Share visual VPS di luar grup (chat Hermes / chat lain):** `wa_sys_monitor`
render in-memory — gak nyimpen file PNG-nya, jadi gak bisa di-`MEDIA:`-in
langsung. Buat ulang lokal: extract `chartHtml` dari
`/root/nael-ai/dist/ai/mcp-baileys.js` (new Function pattern, lihat SKILL.md
utama), isi cards/bars/sections dari response JSON tool, render via
`render_png.mjs` → `/root/vps_monitor.png`. Struktur call yang bener ada di
`templates/vps_monitor_chart.mjs` (tinggal ganti angkanya sama data live).

## Laporan 2 — Endpoint Attack Log (build sendiri)

Sumber: `/root/ENKRIPSI_20_OKT/databases/feature-access-log.json`
shape: `{"logs": [ {feature, email, token, authType, outcome, endpoint, method,
query, ip, userAgent, timestamp, timestampWIB}, ... ]}`.

### Analisis (python, batch dalam satu execute_code)

- total, `Counter(outcome)` (allowed/denied), `Counter(authType)`
  (user/internal/stranger/registered), `Counter(feature)`, timespan.
- **Filtering rentang waktu khusus (e.g. "hari ini jam 12:00 WIB - sekarang"):**
  Parse ISO string dengan timezone WIB (`UTC+7`), filter `dt >= start_wib`.
  Timeline chart untuk rentang 1 hari sebaiknya di-bin per jam (`12:00 WIB`, `13:00 WIB`, ...) bukan per tanggal.
- **Serangan = `authType == 'stranger'` ATAU `outcome == 'denied'`**.
- **Pisah loopback:** `::1`, `127.0.0.1`, `::ffff:127.0.0.1` = probe internal
  server sendiri, BUKAN ancaman asing. Hitung terpisah dan tag `(loopback)`.
- Eksternal: `Counter(ip)`, `Counter(feature)`, `Counter(endpoint)`,
  `Counter(method)`, timeline per jam/hari WIB (parse `timestamp` → tz +7),
  `userAgent` attacker, token/email unik yang kebocor.
- Pattern khas: 40 IP unik hampir semua 1 hit = scan luas, bukan serangan
  terfokus — sebutkan di narasi.

### Dashboard HTML (dark premium style, lihat SKILL.md utama)

- 4 stat cards: Total Log Akses (953) / Ancaman Eksternal (40 IP unik) /
  Ditolak total (76, sebut berapa loopback) / Paling Diserang (feature+endpoint).
- Panel bar: **feature paling diserang** (warn gradient utk top 2) dan
  **timeline per hari WIB** (bar = hit, warn di spike terbesar).
- Tabel: endpoint yang dicolak (endpoint + hit + kategori feature) dan
  **5 ancaman terbaru** (waktu WIB, IP dipersingkat `…` di tengah, feature,
  badge red `denied`). Bullet insight di bawah tabel (scan pattern, token
  bocor, loopback bukan ancaman).
- Footer: path sumber + rentang data + jumlah entri + timestamp render.

### Generator script (reusable)

`scripts/gen_endpoint_report.py` sudah otomatis bikin HTML-nya (analisis +
dashboard + insight) → `/root/laporan_serangan_cypher.html`. Tinggal:

```bash
python3 scripts/gen_endpoint_report.py
node scripts/render_png.mjs /root/laporan_serangan_cypher.html /root/laporan_serangan_cypher.png
```

Pitfall yang sudah difix di generator: label feature pake class `.wlabel`
(width 150px) — tanpa ini `data-enrichment` kepotong jadi "data- enrichmen...".
Footer rentang masih hardcode T15:08Z/T08:09Z — kalau data berubah drastis,
cek footer nyambung.

### Kirim

1. Render: `node scripts/render_png.mjs <in.html> <out.png>` (landscape,
   viewport 1560, fullPage).
2. Verify: `vision_analyze` pada PNG — cek bars aligned, no truncation, angka
   rata kanan, tidak ada overlap.
3. Kirim via `mcp__baileys__wa_send_file`:
   `{ "jid": "<target>", "path": "/abs/out.png", "caption": "..." }` —
   param `path`, bukan `file_path`.
   **Caption harus DETAIL & informatif** (user pernah minta "caption nya harus
   detail dan informatif") — bukan 1-2 baris. Isi: total entri, jumlah ancaman
   eksternal + IP unik, target utama (feature + hit), kredensial bocor kalau
   ada, spike, dan status akhir (semua denied). Contoh di riwayat session
   06 Agu 2026.

### Siapa yang nerima (grup vs pribadi)

- Default (user rule): kirim ke grup RecoVerse `120363186235853203@g.us`.
- Tapi kalau user bilang **"kirim visual / ke chat ini / ke nomer saya"** →
  artinya kirim ke WA PRIBADI user (`6283893964069@s.whatsapp.net`), JANGAN ke
  grup. Lupa JID? Baca `/root/nael-ai/src/config.ts` → `OWNER_JIDS` /
  `ELIGIBLE_JIDS` (dari `.env`, digits-only → `${digits}@s.whatsapp.net`;
  jangan tertukar sama `PAIR_NUMBER` = nomer bot sendiri).
- **"kirim visual" / "kirim ke chat ini"**: jika user meminta "pakai visual" atau "kirim hasilnya ke chat ini", selain menampilkan `MEDIA:<path>` di chat lokal, gunakan `mcp__baileys__wa_send_file` untuk mengirim file gambar PNG hasil render ke JID WhatsApp user (`6283893964069@s.whatsapp.net`).
- `MEDIA:<path>` di chat Hermes cuma preview lokal — gak menggantikan kirim
  WA. Kalau user minta "kirim", commit via `wa_send_file` ke jid yang diminta.

## Pitfalls yang sudah kena

- `wa_send_file` ditolak sekali karena pakai `file_path` — schema-nya `path`.
- Timeline python error `NameError: idx` — nama fungsi harus konsisten
  (definisikan `daykey(x)` sekali, jangan campur `day`/`idx`).
- Jangan hitung `2a02:4780:59:4050::1` sebagai attacker — itu IP internal
  (331 hit allowed, axios INTERNAL), bukan serangan.
- Loopback 36 denied hampir separuh dari total denied — kalau tidak dipisah,
  top attacker slot salah.
