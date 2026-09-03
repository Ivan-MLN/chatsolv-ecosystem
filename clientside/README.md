# ChatSolv Clientside (NextGen UI)

Frontend Next.js 16 (React 19, Turbopack, Framer Motion, Tailwind CSS v4) yang menyajikan antarmuka modern single-page berkinerja tinggi, visual presentasi elegan tema Sage Green, serta demo percakapan interaktif Customer Service AI.

---

## 🎨 Design System & Visual Language

- **Palette Utama (Sage Green)**:
  - Canvas: `#d6ebd8`
  - Dark Accent / Text: `#0e1c10`, `#162b19`, `#1a381d`
  - Primary Sage: `#618264`
  - Vibrant Soft / Highlight: `#79AC78`, `#B0D9B1`, `#D0E7D2`
- **Gaya Visual**: Fluid animated ambient gradient blur mesh (GPU accelerated), minimal crisp typography tanpa bayangan kotor, dan 3D glass tactile oval buttons.
- **Micro-Interactions**: Word blur writer reveal stagger per kata, responsive spring damping transitions.

---

## 📑 Halaman & Struktur Navigasi (Slide Views)

Sistem menggunakan pinned single-viewport carousel berbasis gesture scroll / mouse wheel / swipe touch / keyboard arrow keys:

1. **Slide 01 — Beranda / Welcome**:
   - Headline 2-baris: *"Pelanggan Tidak Menghilang Tiba-Tiba. Mereka Berhenti Menunggu."*
   - CTA buttons langsung menuju Interactive Demo & Coming Soon.
2. **Slide 02 — Demo Interaktif**:
   - Tampilan live WhatsApp Conversation UI berbasis glassmorphism.
   - Smart dynamic frontend simulator bot yang merespons pertanyaan produk, harga, setup, dan WhatsApp channel secara instan.
   - Fitur preset prompt chips untuk uji coba interaktif cepat.
   - Tombol reset percakapan.
3. **Slide 03 — Coming Soon**:
   - Informasi rilis fitur generasi lanjutan ChatSolv dengan visual terpadu seamless.

---

## 🚀 Perintah Development (DX)

```bash
# Instalasi dependensi
npm install

# Menjalankan local development server (port 3333)
npm run dev -- -p 3333

# Type checking & linting
npm run lint

# Production build dengan Next.js Turbopack
npm run build

# Menjalankan production server di port 3333
npm run start -- -p 3333
```

---

## 📂 Struktur Direktori

```text
clientside/
├── app/
│   ├── globals.css         # Tailwind v4, CSS variable tokens & fluid animation keyframes
│   ├── layout.tsx          # Root HTML shell & metadata SEO
│   └── page.tsx            # Root Next.js page controller
├── components/
│   ├── GradientBlurBackground.tsx # GPU-accelerated animated sage mesh blobs
│   ├── HeroScrollScrubber.tsx     # State controller (3-Slide Pinned Scrubber & Interactive Chat)
│   └── LandingPage.tsx            # Secondary standalone landing components
├── public/                 # Static assets (logo transparent, icons)
├── package.json
└── tsconfig.json
```
