"use client";

import React, { useState, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { ArrowRight, ChevronRight, Check, Sparkles } from "lucide-react";

export default function LandingPage() {
  // Hero conversation loop
  const [heroIndex, setHeroIndex] = useState(0);
  const heroChats = [
    {
      user: "Halo, ini masih tersedia?",
      bot: "Tentu, saya bantu cek.",
      tag: "Produk Tersedia • 2 Unit"
    },
    {
      user: "Bisa kirim hari ini ke Surabaya?",
      bot: "Bisa kak, order sebelum 15.00 langsung dikirim.",
      tag: "J&T Express • Ready"
    },
    {
      user: "Ada promo untuk pembelian partai?",
      bot: "Ada kak, saya sambungkan ke tim kami ya.",
      tag: "Diserahkan ke Handler"
    }
  ];

  useEffect(() => {
    const interval = setInterval(() => {
      setHeroIndex((prev) => (prev + 1) % heroChats.length);
    }, 4200);
    return () => clearInterval(interval);
  }, [heroChats.length]);

  // Section 3 Horizontal Light Flow animation active step
  const [flowStep, setFlowStep] = useState(0);
  useEffect(() => {
    const flowInterval = setInterval(() => {
      setFlowStep((prev) => (prev + 1) % 5);
    }, 1800);
    return () => clearInterval(flowInterval);
  }, []);

  // Section 5 Result Flow simulation loop
  const [resultStep, setResultStep] = useState(0);
  useEffect(() => {
    const resultInterval = setInterval(() => {
      setResultStep((prev) => (prev + 1) % 3);
    }, 3800);
    return () => clearInterval(resultInterval);
  }, []);

  return (
    <div className="relative min-h-screen bg-[#070c08] text-[#f1f5f2] selection:bg-[#10b981] selection:text-black overflow-x-hidden">
      {/* Background Subtle Grain & Ambient Emerald Glow */}
      <div className="fixed inset-0 bg-grain pointer-events-none z-0 opacity-80" />
      
      {/* Ambient Glow Orbs */}
      <div className="fixed top-0 left-1/2 -translate-x-1/2 w-[800px] h-[500px] bg-gradient-to-b from-[#10b981]/10 via-[#059669]/5 to-transparent rounded-full blur-[140px] pointer-events-none -z-10" />
      <div className="fixed top-1/2 -right-40 w-[600px] h-[600px] bg-gradient-to-bl from-[#34d399]/8 via-transparent to-transparent rounded-full blur-[150px] pointer-events-none -z-10" />
      <div className="fixed bottom-0 -left-40 w-[600px] h-[600px] bg-gradient-to-tr from-[#059669]/8 via-transparent to-transparent rounded-full blur-[150px] pointer-events-none -z-10" />

      {/* Persistent Minimal Navbar */}
      <header className="fixed top-0 left-0 right-0 z-50 py-5 px-6 md:px-12 max-w-6xl mx-auto flex items-center justify-between pointer-events-auto">
        <a href="#hero" className="flex items-center gap-2.5 group">
          <div className="w-7 h-7 rounded-lg bg-[#10b981]/20 border border-[#10b981]/40 flex items-center justify-center text-[#34d399] font-bold text-sm">
            C
          </div>
          <span className="font-semibold text-lg tracking-tight text-white group-hover:text-[#34d399] transition-colors">
            ChatSolv
          </span>
        </a>

        <nav className="hidden md:flex items-center gap-8 text-xs text-[#94a397] font-medium tracking-wide">
          <a href="#masalah" className="hover:text-white transition-colors">Masalah</a>
          <a href="#cara-kerja" className="hover:text-white transition-colors">Cara Kerja</a>
          <a href="#kendali" className="hover:text-white transition-colors">Kendali</a>
          <a href="#mulai" className="px-4 py-2 rounded-full bg-white/5 border border-white/10 text-white hover:bg-white/10 transition-colors">
            Mulai dengan ChatSolv
          </a>
        </nav>
      </header>

      {/* ========================================================================= */}
      {/* 01 — HERO                                                                 */}
      {/* ========================================================================= */}
      <section id="hero" className="relative min-h-screen flex items-center justify-center pt-28 pb-20 px-6 max-w-6xl mx-auto z-10">
        <div className="w-full grid grid-cols-1 lg:grid-cols-12 gap-12 lg:gap-8 items-center">
          
          {/* Left Hero Narrative (Text 100% Verbatim) */}
          <div className="lg:col-span-7 flex flex-col gap-6">
            <div className="inline-flex items-center gap-2 self-start px-3 py-1 rounded-full bg-[#10b981]/10 border border-[#10b981]/20 text-xs font-mono text-[#34d399]">
              <span className="w-1.5 h-1.5 rounded-full bg-[#34d399] animate-pulse" />
              CHATSOLV
            </div>

            <h1 className="text-4xl sm:text-5xl md:text-6xl font-extrabold text-white tracking-tight leading-[1.12]">
              Pelanggan Tidak Menghilang Tiba-Tiba. <br />
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-[#34d399] via-[#10b981] to-[#6ee7b7]">
                Mereka Berhenti Menunggu.
              </span>
            </h1>

            <div className="space-y-3 text-base sm:text-lg text-[#94a397] leading-relaxed max-w-xl">
              <p>Saat seseorang menghubungi bisnis Anda, mereka sedang punya niat.</p>
              <p>
                Mereka ingin bertanya.<br />
                Mereka ingin memastikan.<br />
                Mereka mungkin sedang mempertimbangkan untuk membeli.
              </p>
              <p className="text-white/90 font-medium">
                Tetapi niat tidak menunggu selamanya.
              </p>
              <p>
                ChatSolv menjaga percakapan tetap berjalan ketika tim Anda sedang sibuk, sehingga pelanggan tidak kehilangan momentum hanya karena tidak ada yang sempat membalas.
              </p>
            </div>

            <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-3 pt-2">
              <a
                href="#mulai"
                className="px-6 py-3.5 rounded-full bg-[#10b981] text-black font-semibold text-sm hover:bg-[#34d399] transition-all flex items-center justify-center gap-2 shadow-lg shadow-[#10b981]/20 cursor-pointer"
              >
                <span>Mulai dengan ChatSolv</span>
                <ArrowRight className="w-4 h-4" />
              </a>
              <a
                href="#cara-kerja"
                className="px-6 py-3.5 rounded-full bg-white/5 border border-white/10 text-white font-medium text-sm hover:bg-white/10 transition-colors flex items-center justify-center gap-2 cursor-pointer"
              >
                <span>Lihat Cara Kerjanya</span>
              </a>
            </div>
          </div>

          {/* Right Floating Ambient Conversations (Tenang, Elegan, Tanpa Mockup Gadget) */}
          <div className="lg:col-span-5 relative flex items-center justify-center min-h-[380px]">
            {/* Background Ambient Drifting Conversation Bubbles */}
            <motion.div
              animate={{ y: [0, -10, 0], opacity: [0.3, 0.45, 0.3] }}
              transition={{ duration: 6, repeat: Infinity, ease: "easeInOut" }}
              className="absolute -top-4 -left-6 px-3.5 py-2 rounded-2xl bg-white/[0.03] border border-white/[0.06] text-xs text-[#94a397] max-w-[200px]"
            >
              Order #9924 sudah diproses min?
            </motion.div>

            <motion.div
              animate={{ y: [0, 12, 0], opacity: [0.25, 0.4, 0.25] }}
              transition={{ duration: 7, repeat: Infinity, ease: "easeInOut", delay: 1 }}
              className="absolute -bottom-6 -right-4 px-3.5 py-2 rounded-2xl bg-white/[0.03] border border-white/[0.06] text-xs text-[#94a397] max-w-[190px]"
            >
              Terima kasih respon cepatnya!
            </motion.div>

            {/* Active Core Living Conversation Card */}
            <div className="relative w-full max-w-md p-6 rounded-3xl bg-[#0e1711]/80 border border-white/10 shadow-2xl backdrop-blur-xl flex flex-col gap-4">
              <div className="flex items-center justify-between text-xs text-[#94a397] border-b border-white/5 pb-3">
                <span className="font-mono text-[11px] text-[#34d399] flex items-center gap-1.5">
                  <span className="w-1.5 h-1.5 rounded-full bg-[#34d399] animate-pulse" />
                  Percakapan Aktif
                </span>
                <span className="font-mono text-[10px] text-white/40">Real-time Stream</span>
              </div>

              <AnimatePresence mode="wait">
                <motion.div
                  key={heroIndex}
                  initial={{ opacity: 0, y: 12 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -12 }}
                  transition={{ duration: 0.45, ease: "easeOut" }}
                  className="space-y-4 py-2"
                >
                  {/* Incoming user bubble */}
                  <div className="self-start max-w-[85%] p-3.5 rounded-2xl rounded-tl-sm bg-white/5 border border-white/10 text-sm text-white/90">
                    {heroChats[heroIndex].user}
                  </div>

                  {/* Flow Arrow Indicator */}
                  <div className="flex items-center gap-2 pl-2 text-xs font-mono text-[#34d399]/70">
                    <span className="text-[10px]">↓</span>
                    <span>{heroChats[heroIndex].tag}</span>
                  </div>

                  {/* Outgoing instant response */}
                  <div className="self-end ml-auto max-w-[90%] p-3.5 rounded-2xl rounded-tr-sm bg-[#10b981]/15 border border-[#10b981]/30 text-sm text-[#34d399]">
                    {heroChats[heroIndex].bot}
                  </div>
                </motion.div>
              </AnimatePresence>

              <div className="pt-2 border-t border-white/5 text-[11px] text-[#94a397] font-mono flex items-center justify-between">
                <span>Respon Kilat Otomatis</span>
                <span className="text-[#34d399]">Momentum Terjaga</span>
              </div>
            </div>

          </div>

        </div>
      </section>

      {/* ========================================================================= */}
      {/* 02 — THE HIDDEN PROBLEM                                                   */}
      {/* ========================================================================= */}
      <section id="masalah" className="relative py-28 px-6 max-w-6xl mx-auto z-10 border-t border-white/5">
        <div className="w-full grid grid-cols-1 lg:grid-cols-12 gap-12 items-center">
          
          {/* Left Narrative (Text 100% Verbatim) */}
          <div className="lg:col-span-7 flex flex-col gap-6">
            <div className="inline-flex items-center gap-2 self-start px-3 py-1 rounded-full bg-white/5 border border-white/10 text-xs font-mono text-[#94a397]">
              02 — THE HIDDEN PROBLEM
            </div>

            <h2 className="text-3xl sm:text-4xl md:text-5xl font-extrabold text-white tracking-tight leading-[1.15]">
              Masalahnya Bukan Chat yang Banyak. <br />
              <span className="text-[#94a397]">
                Masalahnya Ada di Waktu yang Hilang di Antaranya.
              </span>
            </h2>

            <div className="space-y-3 text-base text-[#94a397] leading-relaxed max-w-xl">
              <p>Setiap hari pelanggan datang melalui WhatsApp.</p>
              <p>
                Mereka bertanya.<br />
                Mereka meminta harga.<br />
                Mereka meminta kepastian.<br />
                Mereka menanyakan hal yang sebenarnya sudah pernah ditanyakan pelanggan sebelumnya.
              </p>
              <p className="text-white/90 font-medium">
                Dan sebagian besar bisnis masih menyelesaikan semuanya dengan pola yang sama:<br />
                <span className="font-mono text-sm text-[#34d399]">satu pesan → satu handler → satu jawaban.</span>
              </p>
              <p>
                Ketika chat masih sedikit, pola ini terasa normal.<br />
                Ketika jumlahnya mulai bertambah, semuanya berubah.
              </p>
              <p>
                Satu pesan menunggu.<br />
                Lalu lima.<br />
                Lalu dua belas.<br />
                Tim mulai memilih mana yang harus dijawab lebih dulu.
              </p>
              <p>
                Dan tanpa disadari, pelanggan yang tadinya ingin melanjutkan percakapan mulai kehilangan alasan untuk tetap menunggu.
              </p>
              <p className="text-white font-semibold pt-1 border-l-2 border-[#10b981] pl-3">
                Yang hilang bukan hanya satu chat.<br />
                Bisa jadi satu kesempatan yang tidak pernah kembali.
              </p>
            </div>
          </div>

          {/* Right Visual: Bottleneck Percakapan (Banyak Pesan -> Satu Jalur Sempit Handler) */}
          <div className="lg:col-span-5 flex items-center justify-center">
            <div className="w-full max-w-md p-6 rounded-3xl bg-[#0e1711]/60 border border-white/10 flex flex-col gap-5 backdrop-blur-md">
              <div className="text-xs font-mono text-[#94a397] border-b border-white/5 pb-2 flex items-center justify-between">
                <span>Aliran Pesan Menumpuk</span>
                <span className="text-rose-400 font-bold">+6 pesan baru</span>
              </div>

              {/* Message queue narrowing down to handler bottleneck */}
              <div className="space-y-2.5">
                {/* Message 1 (Fading away / Kehilangan alasan menunggu) */}
                <div className="p-3 rounded-2xl bg-white/[0.02] border border-white/5 text-xs text-white/30 flex items-center justify-between">
                  <span className="line-through">Kak, ini bisa custom warna?</span>
                  <span className="font-mono text-[10px] text-rose-500/50">8 menit • Memudar</span>
                </div>

                {/* Message 2 (Waiting) */}
                <div className="p-3 rounded-2xl bg-white/5 border border-white/10 text-xs text-white/60 flex items-center justify-between">
                  <span>Ada diskon grosir untuk toko?</span>
                  <span className="font-mono text-[10px] text-amber-400/80">Menunggu</span>
                </div>

                {/* Message 3 (Waiting) */}
                <div className="p-3 rounded-2xl bg-white/5 border border-white/10 text-xs text-white/80 flex items-center justify-between">
                  <span>Ongkir ke Jakarta Barat berapa?</span>
                  <span className="font-mono text-[10px] text-amber-400">Belum dijawab</span>
                </div>

                {/* Narrowing funnel lines */}
                <div className="py-2 flex flex-col items-center justify-center gap-1 text-[#94a397]/40 text-xs">
                  <div className="w-0.5 h-4 bg-gradient-to-b from-white/20 to-amber-500/40" />
                  <span className="font-mono text-[10px] text-amber-400/70 tracking-widest uppercase">Jalur Sempit</span>
                  <div className="w-0.5 h-4 bg-gradient-to-b from-amber-500/40 to-white/20" />
                </div>

                {/* Bottleneck Point: Handler */}
                <div className="p-3.5 rounded-2xl bg-white/10 border border-white/20 text-center flex flex-col items-center gap-1 shadow-inner">
                  <span className="text-xs font-bold text-white tracking-wide uppercase font-mono">
                    Handler
                  </span>
                  <span className="text-[10px] text-[#94a397]">
                    Kapasitas Terbatas • Antrean Menumpuk
                  </span>
                </div>
              </div>

              <div className="pt-2 border-t border-white/5 text-[11px] text-[#94a397] font-mono text-center">
                Banyak Pesan → Satu Jalur Sempit
              </div>
            </div>
          </div>

        </div>
      </section>

      {/* ========================================================================= */}
      {/* 03 — HOW CHATSOLV WORKS                                                   */}
      {/* ========================================================================= */}
      <section id="cara-kerja" className="relative py-28 px-6 max-w-6xl mx-auto z-10 border-t border-white/5">
        <div className="flex flex-col gap-12">
          
          {/* Top Narrative (Text 100% Verbatim) */}
          <div className="max-w-3xl flex flex-col gap-5">
            <div className="inline-flex items-center gap-2 self-start px-3 py-1 rounded-full bg-[#10b981]/10 border border-[#10b981]/20 text-xs font-mono text-[#34d399]">
              03 — HOW CHATSOLV WORKS
            </div>

            <h2 className="text-3xl sm:text-5xl font-extrabold text-white tracking-tight leading-[1.12]">
              Tidak Semua Percakapan <br />
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-[#34d399] to-[#10b981]">
                Membutuhkan Tim Anda.
              </span>
            </h2>

            <div className="space-y-3 text-base text-[#94a397] leading-relaxed max-w-2xl">
              <p>Pelanggan tetap mengirim pesan seperti biasa.</p>
              <p>Tetapi sekarang setiap percakapan tidak langsung menunggu seseorang dari tim Anda.</p>
              <p>
                ChatSolv memahami apa yang sedang ditanyakan.<br />
                Mencari informasi yang relevan dari bisnis Anda.<br />
                Kemudian membantu memberikan respons yang sesuai.
              </p>
              <p>
                Pertanyaan yang berulang dapat diselesaikan.<br />
                Informasi dasar dapat diberikan.<br />
                Percakapan tetap bergerak.
              </p>
              <p className="text-white/90 font-medium">
                Dan ketika memang ada situasi yang membutuhkan keputusan atau perhatian khusus—<br />
                baru handler mengambil alih.
              </p>
              <p className="text-white font-semibold">
                Bukan menggantikan tim Anda.<br />
                Tetapi membuat tim Anda masuk hanya ketika memang dibutuhkan.
              </p>
            </div>
          </div>

          {/* Simple Clean Horizontal Flow Animation (Titik Cahaya Melintasi Jalur) */}
          <div className="w-full p-6 sm:p-8 rounded-3xl bg-[#0e1711]/60 border border-white/10 backdrop-blur-md">
            <div className="text-xs font-mono text-[#94a397] mb-6 flex items-center justify-between">
              <span>Animasi Flow Percakapan</span>
              <span className="text-[#34d399]">Aliran Otomatis &amp; Hand-off</span>
            </div>

            {/* Horizontal Nodes Flow */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4 relative">
              {/* Step 1: Pesan Masuk */}
              <div className={`p-4 rounded-2xl transition-all duration-500 border ${flowStep === 0 ? "bg-[#10b981]/15 border-[#10b981] text-white shadow-lg shadow-[#10b981]/10" : "bg-white/[0.03] border-white/5 text-[#94a397]"}`}>
                <div className="text-[10px] font-mono text-[#34d399] mb-1">LANGKAH 1</div>
                <div className="text-sm font-bold text-white">Pesan Masuk</div>
                <div className="text-[11px] text-[#94a397] mt-1">Pelanggan mengirim pesan WhatsApp</div>
              </div>

              {/* Step 2: Dipahami */}
              <div className={`p-4 rounded-2xl transition-all duration-500 border ${flowStep === 1 ? "bg-[#10b981]/15 border-[#10b981] text-white shadow-lg shadow-[#10b981]/10" : "bg-white/[0.03] border-white/5 text-[#94a397]"}`}>
                <div className="text-[10px] font-mono text-[#34d399] mb-1">LANGKAH 2</div>
                <div className="text-sm font-bold text-white">Dipahami</div>
                <div className="text-[11px] text-[#94a397] mt-1">ChatSolv menganalisis maksud obrolan</div>
              </div>

              {/* Step 3: Informasi Ditemukan */}
              <div className={`p-4 rounded-2xl transition-all duration-500 border ${flowStep === 2 ? "bg-[#10b981]/15 border-[#10b981] text-white shadow-lg shadow-[#10b981]/10" : "bg-white/[0.03] border-white/5 text-[#94a397]"}`}>
                <div className="text-[10px] font-mono text-[#34d399] mb-1">LANGKAH 3</div>
                <div className="text-sm font-bold text-white">Informasi Ditemukan</div>
                <div className="text-[11px] text-[#94a397] mt-1">Mencari data bisnis &amp; SOP yang relevan</div>
              </div>

              {/* Step 4: Respons Diberikan */}
              <div className={`p-4 rounded-2xl transition-all duration-500 border ${flowStep === 3 ? "bg-[#10b981]/15 border-[#10b981] text-white shadow-lg shadow-[#10b981]/10" : "bg-white/[0.03] border-white/5 text-[#94a397]"}`}>
                <div className="text-[10px] font-mono text-[#34d399] mb-1">LANGKAH 4</div>
                <div className="text-sm font-bold text-white">Respons Diberikan</div>
                <div className="text-[11px] text-[#94a397] mt-1">Balasan instan terkirim ke pelanggan</div>
              </div>
            </div>

            {/* Branching to Handler */}
            <div className="mt-6 pt-6 border-t border-white/5 flex flex-col sm:flex-row items-center justify-between gap-4">
              <div className="flex items-center gap-3">
                <span className="w-2 h-2 rounded-full bg-[#10b981]" />
                <span className="text-sm text-white font-medium">Butuh Penanganan Khusus?</span>
              </div>
              <div className="flex items-center gap-3">
                <span className="text-xs font-mono text-[#94a397]">→</span>
                <div className={`px-5 py-2.5 rounded-full border transition-all duration-500 ${flowStep === 4 ? "bg-[#10b981] text-black font-bold border-[#10b981]" : "bg-white/5 border-white/10 text-white"}`}>
                  Handler Mengambil Alih
                </div>
              </div>
            </div>

            <div className="mt-4 text-center text-xs text-[#94a397] font-mono">
              ChatSolv menangani yang bisa ditangani. Tim masuk ketika dibutuhkan.
            </div>
          </div>

        </div>
      </section>

      {/* ========================================================================= */}
      {/* 04 — CONTROL                                                              */}
      {/* ========================================================================= */}
      <section id="kendali" className="relative py-28 px-6 max-w-6xl mx-auto z-10 border-t border-white/5">
        <div className="w-full grid grid-cols-1 lg:grid-cols-12 gap-12 items-center">
          
          {/* Left Narrative (Text 100% Verbatim) */}
          <div className="lg:col-span-7 flex flex-col gap-6">
            <div className="inline-flex items-center gap-2 self-start px-3 py-1 rounded-full bg-white/5 border border-white/10 text-xs font-mono text-[#94a397]">
              04 — CONTROL
            </div>

            <h2 className="text-3xl sm:text-5xl font-extrabold text-white tracking-tight leading-[1.12]">
              Otomatis Bukan Berarti <br />
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-[#34d399] to-[#10b981]">
                Kehilangan Kendali.
              </span>
            </h2>

            <div className="space-y-4 text-base text-[#94a397] leading-relaxed max-w-xl">
              <p>
                Sistem yang menangani pelanggan seharusnya tidak terasa seperti sesuatu yang bekerja di luar kendali Anda.
              </p>
              <p>
                Karena itu ChatSolv bekerja berdasarkan informasi dan aturan bisnis Anda.
              </p>
              <p className="text-white/90">
                Anda menentukan apa yang ChatSolv ketahui.<br />
                Anda menentukan bagaimana cara berkomunikasi.<br />
                Anda tetap dapat melihat percakapan.<br />
                Dan Anda tetap dapat mengambil alih ketika diperlukan.
              </p>

              {/* 4 Pillars Breakdown */}
              <div className="space-y-3 pt-2">
                <div>
                  <h4 className="text-sm font-bold text-white">Knowledge Bisnis</h4>
                  <p className="text-xs text-[#94a397]">
                    Berikan informasi yang memang digunakan bisnis Anda sehari-hari. Produk. Layanan. FAQ. SOP. Kebijakan. Informasi operasional lainnya.
                  </p>
                </div>

                <div>
                  <h4 className="text-sm font-bold text-white">Cara Berkomunikasi</h4>
                  <p className="text-xs text-[#94a397]">
                    Sesuaikan bagaimana ChatSolv berbicara kepada pelanggan. Tidak harus terasa kaku. Tidak harus terdengar seperti template.
                  </p>
                </div>

                <div>
                  <h4 className="text-sm font-bold text-white">Human Handoff</h4>
                  <p className="text-xs text-[#94a397]">
                    Ketika percakapan membutuhkan keputusan, kasus khusus, atau bantuan langsung—handler dapat mengambil alih.
                  </p>
                </div>

                <div>
                  <h4 className="text-sm font-bold text-white">Conversation Visibility</h4>
                  <p className="text-xs text-[#94a397]">
                    Anda tetap tahu apa yang sedang terjadi dalam percakapan pelanggan.
                  </p>
                </div>
              </div>

              <p className="text-white font-semibold pt-2 border-l-2 border-[#10b981] pl-3">
                Yang berubah bukan kendali Anda.<br />
                Yang berubah adalah berapa banyak pekerjaan repetitif yang harus dilakukan tim.
              </p>
            </div>
          </div>

          {/* Right Visual: Control Surface Minimal Panel */}
          <div className="lg:col-span-5 flex items-center justify-center">
            <div className="w-full max-w-md p-6 rounded-3xl bg-[#0e1711]/70 border border-white/10 shadow-2xl backdrop-blur-md flex flex-col gap-4">
              <div className="text-xs font-mono text-[#94a397] border-b border-white/5 pb-2.5 flex items-center justify-between">
                <span>Panel Kendali ChatSolv</span>
                <span className="text-[#34d399] font-bold">Status: Aktif</span>
              </div>

              {/* 4 Control Parameters */}
              <div className="space-y-2.5">
                <div className="p-3 rounded-2xl bg-white/[0.03] border border-white/5 flex items-center justify-between">
                  <span className="text-xs text-white/90">Knowledge</span>
                  <span className="text-xs font-mono font-semibold text-[#34d399] px-2 py-0.5 rounded bg-[#10b981]/15">Connected</span>
                </div>

                <div className="p-3 rounded-2xl bg-white/[0.03] border border-white/5 flex items-center justify-between">
                  <span className="text-xs text-white/90">Communication Style</span>
                  <span className="text-xs font-mono font-semibold text-[#34d399] px-2 py-0.5 rounded bg-[#10b981]/15">Friendly &amp; Professional</span>
                </div>

                <div className="p-3 rounded-2xl bg-white/[0.03] border border-white/5 flex items-center justify-between">
                  <span className="text-xs text-white/90">Handoff</span>
                  <span className="text-xs font-mono font-semibold text-[#34d399] px-2 py-0.5 rounded bg-[#10b981]/15">Active</span>
                </div>

                <div className="p-3 rounded-2xl bg-white/[0.03] border border-white/5 flex items-center justify-between">
                  <span className="text-xs text-white/90">Conversation Visibility</span>
                  <span className="text-xs font-mono font-semibold text-[#34d399] px-2 py-0.5 rounded bg-[#10b981]/15">On</span>
                </div>
              </div>

              {/* Live Handoff Stream Simulation */}
              <div className="mt-2 p-3.5 rounded-2xl bg-[#10b981]/10 border border-[#10b981]/30 flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="w-2 h-2 rounded-full bg-amber-400 animate-ping" />
                  <span className="text-xs text-white font-medium">Status: Needs Handler</span>
                </div>
                <span className="text-xs font-mono font-bold text-[#34d399]">→ Team</span>
              </div>

              <div className="pt-2 border-t border-white/5 text-[11px] text-[#94a397] font-mono text-center">
                “Semuanya otomatis, tetapi saya tetap memegang kendali.”
              </div>
            </div>
          </div>

        </div>
      </section>

      {/* ========================================================================= */}
      {/* 05 — RESULT / FINAL CTA                                                   */}
      {/* ========================================================================= */}
      <section id="mulai" className="relative py-28 px-6 max-w-6xl mx-auto z-10 border-t border-white/5">
        <div className="w-full grid grid-cols-1 lg:grid-cols-12 gap-12 items-center">
          
          {/* Left Narrative (Text 100% Verbatim) */}
          <div className="lg:col-span-7 flex flex-col gap-6">
            <div className="inline-flex items-center gap-2 self-start px-3 py-1 rounded-full bg-[#10b981]/10 border border-[#10b981]/20 text-xs font-mono text-[#34d399]">
              05 — RESULT
            </div>

            <h2 className="text-3xl sm:text-5xl font-extrabold text-white tracking-tight leading-[1.12]">
              Tim Anda Tidak Harus <br />
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-[#34d399] via-[#10b981] to-[#6ee7b7]">
                Menghabiskan Hari untuk Membalas Hal yang Sama.
              </span>
            </h2>

            <div className="space-y-3 text-base text-[#94a397] leading-relaxed max-w-xl">
              <p>Semakin banyak bisnis berkembang, semakin banyak percakapan yang harus ditangani.</p>
              <p>Tetapi jumlah pekerjaan repetitif tidak seharusnya ikut tumbuh dengan kecepatan yang sama.</p>
              <p>
                Biarkan ChatSolv menangani percakapan yang dapat diselesaikan tanpa perhatian langsung dari tim Anda.
              </p>
              <p>
                Sehingga tim bisa fokus pada pelanggan, keputusan, dan pekerjaan yang memang membutuhkan manusia.
              </p>
              <p className="text-white/90 font-medium">
                Lebih sedikit chat yang tertinggal.<br />
                Lebih sedikit pelanggan menunggu.<br />
                Lebih sedikit pertanyaan yang dijawab berulang kali.<br />
                Dan lebih banyak percakapan tetap bergerak ketika pelanggan masih memiliki niat.
              </p>
            </div>

            {/* Final Payoff Climax */}
            <div className="pt-4 border-t border-white/10 space-y-1">
              <p className="text-xl sm:text-2xl font-bold text-white">Pelanggan Tetap Mendapat Jawaban.</p>
              <p className="text-xl sm:text-2xl font-bold text-white">Tim Tetap Memegang Kendali.</p>
              <p className="text-xl sm:text-2xl font-bold text-[#34d399]">Percakapan Tetap Berjalan.</p>
            </div>

            <div className="pt-3">
              <button className="px-8 py-4 rounded-full bg-[#10b981] text-black font-bold text-base hover:bg-[#34d399] transition-all shadow-xl shadow-[#10b981]/25 flex items-center gap-2 cursor-pointer">
                <span>Mulai dengan ChatSolv</span>
                <ArrowRight className="w-5 h-5" />
              </button>
            </div>
          </div>

          {/* Right Visual: Payoff Hero Visual (Percakapan Teratur & Mengalir) */}
          <div className="lg:col-span-5 flex items-center justify-center">
            <div className="w-full max-w-md p-6 rounded-3xl bg-[#0e1711]/70 border border-white/10 shadow-2xl backdrop-blur-md flex flex-col gap-4">
              <div className="text-xs font-mono text-[#94a397] border-b border-white/5 pb-2.5 flex items-center justify-between">
                <span>Aliran Percakapan Teratur</span>
                <span className="text-[#34d399]">Mengalir Lancar</span>
              </div>

              {/* Dynamic Smooth Ordered Flow Stream */}
              <div className="space-y-3 my-auto">
                <AnimatePresence mode="wait">
                  {resultStep === 0 && (
                    <motion.div
                      key="res-1"
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -10 }}
                      className="space-y-2"
                    >
                      <div className="p-3 rounded-2xl bg-white/5 text-xs text-white/80">
                        1. Pesan masuk: &quot;Bisa cek stok size L?&quot;
                      </div>
                      <div className="p-3 rounded-2xl bg-[#10b981]/15 border border-[#10b981]/30 text-xs text-[#34d399]">
                        2. Respons otomatis: &quot;Ready 4 unit kak!&quot;
                      </div>
                      <div className="text-[10px] font-mono text-[#94a397] text-right">
                        ✓ Percakapan Selesai
                      </div>
                    </motion.div>
                  )}

                  {resultStep === 1 && (
                    <motion.div
                      key="res-2"
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -10 }}
                      className="space-y-2"
                    >
                      <div className="p-3 rounded-2xl bg-white/5 text-xs text-white/80">
                        1. Pesan masuk: &quot;Berapa estimasi ongkir ke Medan?&quot;
                      </div>
                      <div className="p-3 rounded-2xl bg-[#10b981]/15 border border-[#10b981]/30 text-xs text-[#34d399]">
                        2. Respons otomatis: &quot;JNE Reguler Rp 24.000 (2-3 hari).&quot;
                      </div>
                      <div className="text-[10px] font-mono text-[#94a397] text-right">
                        ✓ Percakapan Selesai
                      </div>
                    </motion.div>
                  )}

                  {resultStep === 2 && (
                    <motion.div
                      key="res-3"
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -10 }}
                      className="space-y-2"
                    >
                      <div className="p-3 rounded-2xl bg-white/5 text-xs text-white/80">
                        1. Pesan masuk: &quot;Mau negosiasi tender 200 unit.&quot;
                      </div>
                      <div className="p-3 rounded-2xl bg-amber-500/15 border border-amber-500/30 text-xs text-amber-300">
                        2. Dialihkan ke Handler: Admin mengambil alih tiket.
                      </div>
                      <div className="text-[10px] font-mono text-amber-400 text-right">
                        ✓ Deal Selesai oleh Tim
                      </div>
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>

              <div className="pt-2 border-t border-white/5 text-[11px] text-[#94a397] font-mono text-center">
                Dari Menunggu → Menjadi Mengalir
              </div>
            </div>
          </div>

        </div>
      </section>

      {/* ========================================================================= */}
      {/* FOOTER                                                                    */}
      {/* ========================================================================= */}
      <footer className="relative py-16 px-6 max-w-6xl mx-auto z-10 border-t border-white/5">
        <div className="grid grid-cols-1 md:grid-cols-12 gap-10 pb-12">
          
          {/* Brand Col */}
          <div className="md:col-span-5 flex flex-col gap-3">
            <div className="flex items-center gap-2.5">
              <div className="w-7 h-7 rounded-lg bg-[#10b981]/20 border border-[#10b981]/40 flex items-center justify-center text-[#34d399] font-bold text-sm">
                C
              </div>
              <span className="font-semibold text-lg text-white">ChatSolv</span>
            </div>
            <p className="text-xs text-[#94a397] leading-relaxed max-w-sm">
              Percakapan pelanggan tetap berjalan, bahkan ketika tim Anda sedang mengerjakan hal lain.
            </p>
          </div>

          {/* Links 1: Product */}
          <div className="md:col-span-2 flex flex-col gap-2.5">
            <div className="text-xs font-bold text-white uppercase font-mono tracking-wider">Product</div>
            <ul className="space-y-2 text-xs text-[#94a397]">
              <li><a href="#cara-kerja" className="hover:text-white transition-colors">Cara Kerja</a></li>
              <li><a href="#mulai" className="hover:text-white transition-colors">Harga</a></li>
              <li><a href="#hero" className="hover:text-white transition-colors">FAQ</a></li>
            </ul>
          </div>

          {/* Links 2: Company */}
          <div className="md:col-span-2 flex flex-col gap-2.5">
            <div className="text-xs font-bold text-white uppercase font-mono tracking-wider">Company</div>
            <ul className="space-y-2 text-xs text-[#94a397]">
              <li><a href="#hero" className="hover:text-white transition-colors">Tentang ChatSolv</a></li>
              <li><a href="#hero" className="hover:text-white transition-colors">Blog</a></li>
            </ul>
          </div>

          {/* Links 3: Legal & CTA */}
          <div className="md:col-span-3 flex flex-col gap-3">
            <div className="text-xs font-bold text-white uppercase font-mono tracking-wider">Legal</div>
            <ul className="space-y-2 text-xs text-[#94a397]">
              <li><a href="#hero" className="hover:text-white transition-colors">Privacy Policy</a></li>
              <li><a href="#hero" className="hover:text-white transition-colors">Terms of Service</a></li>
            </ul>
            <div className="pt-2">
              <a
                href="#mulai"
                className="inline-block px-4 py-2 rounded-full bg-[#10b981]/15 border border-[#10b981]/30 text-xs font-medium text-[#34d399] hover:bg-[#10b981] hover:text-black transition-all"
              >
                Mulai dengan ChatSolv
              </a>
            </div>
          </div>

        </div>

        {/* Copyright separator */}
        <div className="pt-6 border-t border-white/5 flex flex-col sm:flex-row items-center justify-between text-xs text-[#94a397] font-mono">
          <span>© ChatSolv. All rights reserved.</span>
          <span className="text-white/30 text-[11px]">Designed with calm emerald atmosphere.</span>
        </div>
      </footer>

    </div>
  );
}
