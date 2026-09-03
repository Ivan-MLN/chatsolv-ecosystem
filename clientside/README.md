# ChatSolv Clientside (NextGen UI Engine) 🌿
*High-Performance Single-Page Presentation & Interactive AI Demo Interface*

---

## 🌟 1. Pengenalan Clientside

**Clientside** adalah aplikasi frontend resmi untuk ekosistem ChatSolv Next-Generation. Didesain secara khusus sebagai web presentasi interaktif single-page yang menggabungkan kecepatan kilat, estetika visual bertema **Sage Green**, animasi transisi tingkat lanjut yang halus (60-120 FPS), dan showcase Customer Service AI yang dapat langsung dicoba secara live oleh calon pengguna.

Dibangun di atas **Next.js 16 App Router** terbaru dengan compiler **Turbopack**, **React 19**, **Framer Motion**, dan **Tailwind CSS v4**, aplikasi ini memberikan standar pengalaman pengembang (Developer Experience / DX) yang modern dan terstruktur.

---

## 🎨 2. Design System & Spesifikasi Visual

### 🟢 Palette Warna Resmi (Sage Green Tokens)
Frontend menggunakan spektrum palet Sage Green bertingkat untuk menciptakan nuansa tenang, profesional, dan segar:
- **Canvas Base**: `#d6ebd8` (Background lembut yang menyatu dengan fluid mesh).
- **Text & High Contrast**: 
  - `#0e1c10` (Deepest Forest Green - Judul Utama / Headline baris 1).
  - `#162b19` (Solid Body Text - Teks isi dan label aktif).
  - `#1a381d` (Sub-headline / Headline baris 2).
  - `#253e28` (Supporting description text).
- **Primary & Interactive Elements**:
  - `#618264` (Primary Emerald Sage - Tombol CTA, pill aktif, bubble chat user, border penegas).
  - `#79AC78` (Accent Mid Sage - Gradient highlight).
  - `#B0D9B1` (Soft Sage - Border halus & subtle divider).
  - `#D0E7D2` (Light Sage - Latar belakang kartu & ambient glow).

### ✨ Visual Effects & Motion Language
- **GPU-Accelerated Gradient Blur Mesh**: Komponen background yang merender 5 blob animated gradient (`@keyframes float-1` s/d `float-4` dan `float-center`) dengan filter Gaussian blur `70px - 85px` dan hardware-accelerated `translate3d` untuk memastikan nol stuttering / jank.
- **Blur Writer Typography**: Animasi entrance teks kata-per-kata yang menggunakan stagger spring physics (`wordBlurVariant`), bergeser dari blur `10px` + scale `0.96` menjadi tajam dan jernih.
- **Glassmorphism Panels**: Panel semi-transparan `bg-white/70 backdrop-blur-2xl border-white/90` yang memberikan kedalaman tactile modern.

---

## 📑 3. Navigasi & Alur Carousel (Pinned Scrubber)

Aplikasi mengadopsi pola arsitektur **Pinned Full-Viewport Scrubber**, di mana seluruh konten berada dalam viewport yang terpusat secara optis tanpa scrollbar halaman browser yang berantakan:

1. **Slide 0 — Beranda (Hero Welcome)**:
   - Headline 2 baris terstruktur: *"Pelanggan Tidak Menghilang Tiba-Tiba."* & *"Mereka Berhenti Menunggu."*
   - Kalimat pendukung terukur: *"ChatSolv membantu bisnis merespons lebih cepat, tetap konsisten, dan menjaga setiap pelanggan tetap terlayani."*
   - Tombol CTA 3D Glass langsung mengarahkan pengguna ke slide Demo Interaktif dan Coming Soon.
2. **Slide 1 — Demo Interaktif (Live AI Conversation)**:
   - Antarmuka chat WhatsApp modern real-time.
   - Status indikator aktif, avatar bot AI dan user.
   - Quick-reply preset prompt chips untuk pertanyaan umum (penjelasan ChatSolv, harga langganan, integrasi WhatsApp, cara kerja).
   - Input bar dengan simulator auto-reply pintar dan animasi typing indicator 3 dots.
   - Fitur reset percakapan untuk memulai ulang sesi simulasi.
   - Deteksi area scroll khusus (`.chat-scroll-area`) sehingga pengguna bebas membaca chat tanpa memicu pergantian slide.
3. **Slide 2 — Coming Soon**:
   - Informasi rilis fitur generasi berikutnya dengan visual minimalis dan tipografi senada.

Transisi antar slide dapat dilakukan melalui:
- **Mouse Wheel / Trackpad Gesture** (dengan debounce transisi `500ms`).
- **Touch Swipe Mobile** (deteksi threshold `40px` vertikal).
- **Keyboard Navigation** (`ArrowDown`, `ArrowUp`, `PageDown`, `PageUp`).
- **Navbar Buttons** di bagian atas layar.

---

## 📂 4. Penjelasan Detail Setiap File & Fungsinya

```text
clientside/
├── app/
│   ├── globals.css                     # Definisi token Tailwind v4, variabel warna CSS, & keyframes animasi fluid mesh
│   ├── layout.tsx                      # Root shell HTML, konfigurasi font Google (Plus Jakarta Sans & Geist), metadata SEO & OpenGraph
│   ├── page.tsx                        # Controller utama halaman yang memuat komponen HeroScrollScrubber
│   └── favicon.ico                     # Icon favicon browser
├── components/
│   ├── GradientBlurBackground.tsx      # Komponen murni background SVG/CSS mesh blobs dengan rotasi dan pergerakan GPU multi-axis
│   ├── HeroScrollScrubber.tsx          # State engine utama: menangani gesture event listener, transisi slide, dan live chat simulator
│   └── LandingPage.tsx                 # Koleksi komponen sekunder landing page cadangan
├── lib/
│   └── utils.ts                        # Utility helper fungsi class merging (clsx + tailwind-merge)
├── public/
│   ├── chatsolv-logo-transparent.png   # Aset logo ChatSolv resmi berlatar transparan
│   └── logo-dark.png                   # Aset logo versi gelap
├── next.config.ts                      # Konfigurasi Next.js 16 (Turbopack bundler options & optimizations)
├── postcss.config.mjs                  # Konfigurasi PostCSS compiler untuk Tailwind CSS v4
├── eslint.config.mjs                   # Konfigurasi ESLint 9 untuk aturan React & TypeScript purity
├── tsconfig.json                       # Konfigurasi TypeScript compiler (strict mode, path alias @/*)
└── package.json                        # Definisi dependensi (framer-motion, lucide-react, tailwindcss v4, next, react 19)
```

---

## 🛠️ 5. Panduan Menjalankan & Build (DX)

### Prasyarat
- Node.js versi 18.18.0 atau lebih baru.
- Port default: `3333` (dihubungkan ke Cloudflare tunnel `cs.naeladtya.my.id`).

### Perintah Utama
```bash
# 1. Menginstall dependensi
npm install

# 2. Menjalankan development server pada port 3333
npm run dev -- -p 3333

# 3. Menjalankan pemeriksaan linter
npm run lint

# 4. Membuat production build teroptimasi dengan Turbopack
npm run build

# 5. Menjalankan production server
npm run start -- -p 3333
```
