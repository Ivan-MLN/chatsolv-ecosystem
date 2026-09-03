# Cypher Executive Multi-Page HTML to PDF Build Workflow

Prosedur pembaruan dan render dokumen laporan multi-halaman Cypher (`laporan_build` -> `laporan_new.pdf`).

## Struktur File & Modul

Lokasi dokumen di `/home/nldt/laporan_build/`:
- `head.html`: `@page { size: A4; margin: 0; }`, variabel warna CSS, tipografi print-first.
- `body1.html`: Halaman 1 (Cover) + Halaman 2 (Income Realtime) + Halaman 3 (Rincian Transaksi Approved).
- `body2.html`: Halaman 4 (Rekomendasi Marketing) + Halaman 5 (Intisari Non-Teknis).
- `body3.html`: Halaman 6 (Ranking Fitur Terfavorit) + Halaman 7 (Status Keaktifan Fitur & Backup API).
- `body4.html`: Halaman 8 (Kredensial Operasional) + Halaman 9 (Spesifikasi Server & Infrastruktur).
- `body5.html`: Halaman 10–12 (Dokumentasi Visual / Screenshot) + Halaman 13 (Penutup & Signoff).
- `laporan.html`: File gabungan (`head.html` + `body1..5.html` + `</body></html>`).
- `laporan_new.pdf`: Output PDF siap cetak / distribusi.

## Sumber Data Realtime (Cypher Production DB)

1. **Pendapatan & Transaksi Harian**: `/root/ENKRIPSI_20_OKT/databases/payments.json`
   - Filter tanggal kemarin berdasarkan zona WIB (`UTC+7` dari `createdAt`).
   - Ekstrak nominal approved (`amount` / `price`), session ID, email pembeli, dan token cypher.
2. **Statistik Penggunaan Fitur (3 Hari Terakhir)**: `/root/ENKRIPSI_20_OKT/databases/feature-access-log.json` dan `balance-log.json`.
   - Hitung total request, sukses (`allowed`/`granted`), dan denied per fitur.

## Alur Update & Kompilasi

1. **Extract & Calculate Data**:
   Hitung agregat transaksi (total pengajuan, approved, tingkat keberhasilan, rata-rata top-up) dan tabel rincian transaksi approved.
2. **Update Sub-HTML Parts (`body1..body5.html`)**:
   - Perbarui tanggal pengecekan terakhir pada Cover & Penutup (misal `03 September 2026 · 08:00 WIB`).
   - Perbarui tabel KPI dan baris transaksi di `body1.html`.
   - Perbarui ranking agregat fitur 3 hari di `body3.html`.
3. **Kompilasi ke `laporan.html`**:
   ```python
   with open('head.html') as f: head = f.read()
   bodies = "".join(open(f'body{i}.html').read() for i in range(1, 6))
   open('laporan.html', 'w').write(head + bodies + '</body></html>')
   ```
4. **Ekspor PDF Menggunakan Headless Chrome**:
   ```bash
   google-chrome --headless --disable-gpu --no-sandbox --print-to-pdf=/home/nldt/laporan_build/laporan_new.pdf file:///home/nldt/laporan_build/laporan.html
   ```
5. **Verifikasi Kualitas & Integritas**:
   - Ekstrak halaman ke PNG menggunakan `pdftoppm -png -r 150 laporan_new.pdf qa/pg_check`.
   - Periksa visual halaman transaksi (Hal 2 dan Hal 3) menggunakan `vision_analyze` atau `browser_vision` untuk memastikan tidak ada pemotongan teks (*offside*), tata letak rapi, dan nominal akurat.
