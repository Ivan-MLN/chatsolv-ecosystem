"use client";

import React, { useState, useEffect, useRef } from "react";
import { motion, AnimatePresence, type Variants } from "framer-motion";
import { ArrowRight, Send, Sparkles, User, Bot, CheckCheck, RefreshCw } from "lucide-react";
import Image from "next/image";
import { GradientBlurBackground } from "@/components/GradientBlurBackground";

const sentenceContainer: Variants = {
  hidden: { opacity: 0 },
  show: {
    opacity: 1,
    transition: {
      staggerChildren: 0.08,
      delayChildren: 0.12,
    }
  },
  exit: {
    opacity: 0,
    transition: {
      duration: 0.18,
      ease: "easeInOut"
    }
  }
};

const wordBlurVariant: Variants = {
  hidden: { 
    opacity: 0, 
    filter: "blur(10px)", 
    y: 10,
    scale: 0.96
  },
  show: { 
    opacity: 1, 
    filter: "blur(0px)", 
    y: 0,
    scale: 1,
    transition: {
      type: "spring",
      stiffness: 200,
      damping: 20,
      mass: 0.5
    }
  },
  exit: {
    opacity: 0,
    filter: "blur(6px)",
    y: -8,
    transition: { duration: 0.16 }
  }
};

const introContainerStagger: Variants = {
  hidden: { opacity: 0 },
  show: {
    opacity: 1,
    transition: {
      staggerChildren: 0.14,
      delayChildren: 0.08,
    }
  },
  exit: {
    opacity: 0,
    transition: {
      duration: 0.18,
      ease: "easeInOut"
    }
  }
};

const itemVariants: Variants = {
  hidden: { opacity: 0, y: 16, filter: "blur(6px)" },
  show: { 
    opacity: 1, 
    y: 0, 
    filter: "blur(0px)",
    transition: {
      type: "spring",
      stiffness: 240,
      damping: 22,
      mass: 0.55
    }
  },
  exit: { 
    opacity: 0, 
    y: -8, 
    filter: "blur(4px)",
    transition: { duration: 0.16, ease: "easeInOut" }
  }
};

interface Message {
  id: string;
  sender: "user" | "bot";
  text: string;
  timestamp: string;
}

const INITIAL_MESSAGES: Message[] = [
  {
    id: "m1",
    sender: "bot",
    text: "Halo! Selamat datang di ChatSolv Demo AI 👋\nAda informasi produk, harga, atau layanan yang ingin Anda tanyakan?",
    timestamp: "Baru saja",
  },
];

const PRESET_PROMPTS = [
  "Bisa jelaskan apa itu ChatSolv?",
  "Berapa harga langganannya?",
  "Apakah bisa integrasi dengan nomor WhatsApp sendiri?",
  "Bagaimana cara kerjanya?",
];

export function HeroScrollScrubber() {
  // activeStep: 0 (Welcome), 1 (Demo Chat), 2 (Coming Soon)
  const [activeStep, setActiveStep] = useState(0);
  const isTransitioningRef = useRef(false);
  const touchStartYRef = useRef(0);

  // Demo Chat states
  const [messages, setMessages] = useState<Message[]>(INITIAL_MESSAGES);
  const [inputMsg, setInputMsg] = useState("");
  const [isTyping, setIsTyping] = useState(false);
  const chatBottomRef = useRef<HTMLDivElement>(null);

  // Auto scroll chat to bottom
  useEffect(() => {
    if (activeStep === 1) {
      chatBottomRef.current?.scrollIntoView({ behavior: "smooth" });
    }
  }, [messages, isTyping, activeStep]);

  // Reset scroll to top on refresh
  useEffect(() => {
    if (typeof window !== "undefined") {
      window.history.scrollRestoration = "manual";
      window.scrollTo(0, 0);
    }
  }, []);

  // Intercept scroll/wheel to transition states smoothly inside pinned viewport
  useEffect(() => {
    const handleWheel = (e: WheelEvent) => {
      // Check if user is scrolling inside chat conversation container
      const target = e.target as HTMLElement;
      if (target && target.closest(".chat-scroll-area")) {
        return; // allow normal scrolling inside chat box
      }

      e.preventDefault();
      if (isTransitioningRef.current) return;

      if (e.deltaY > 20) {
        setActiveStep((prev) => {
          if (prev < 2) {
            isTransitioningRef.current = true;
            setTimeout(() => { isTransitioningRef.current = false; }, 500);
            return prev + 1;
          }
          return prev;
        });
      } else if (e.deltaY < -20) {
        setActiveStep((prev) => {
          if (prev > 0) {
            isTransitioningRef.current = true;
            setTimeout(() => { isTransitioningRef.current = false; }, 500);
            return prev - 1;
          }
          return prev;
        });
      }
    };

    const handleTouchStart = (e: TouchEvent) => {
      touchStartYRef.current = e.touches[0].clientY;
    };

    const handleTouchMove = (e: TouchEvent) => {
      const target = e.target as HTMLElement;
      if (target && target.closest(".chat-scroll-area")) {
        return;
      }

      const touchY = e.touches[0].clientY;
      const diff = touchStartYRef.current - touchY;
      if (Math.abs(diff) > 40 && !isTransitioningRef.current) {
        if (diff > 0) {
          setActiveStep((prev) => {
            if (prev < 2) {
              isTransitioningRef.current = true;
              setTimeout(() => { isTransitioningRef.current = false; }, 500);
              return prev + 1;
            }
            return prev;
          });
        } else {
          setActiveStep((prev) => {
            if (prev > 0) {
              isTransitioningRef.current = true;
              setTimeout(() => { isTransitioningRef.current = false; }, 500);
              return prev - 1;
            }
            return prev;
          });
        }
        touchStartYRef.current = touchY;
      }
    };

    const handleKeyDown = (e: KeyboardEvent) => {
      if (isTransitioningRef.current) return;
      if (e.key === "ArrowDown" || e.key === "PageDown") {
        setActiveStep((prev) => Math.min(2, prev + 1));
      } else if (e.key === "ArrowUp" || e.key === "PageUp") {
        setActiveStep((prev) => Math.max(0, prev - 1));
      }
    };

    window.addEventListener("wheel", handleWheel, { passive: false });
    window.addEventListener("touchstart", handleTouchStart, { passive: true });
    window.addEventListener("touchmove", handleTouchMove, { passive: false });
    window.addEventListener("keydown", handleKeyDown);

    return () => {
      window.removeEventListener("wheel", handleWheel);
      window.removeEventListener("touchstart", handleTouchStart);
      window.removeEventListener("touchmove", handleTouchMove);
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, []);

  const messageCountRef = useRef(1);

  const handleSendMessage = (textToSend?: string) => {
    const text = textToSend || inputMsg.trim();
    if (!text) return;

    messageCountRef.current += 1;
    const currentId = messageCountRef.current;

    const userMessage: Message = {
      id: `u-${currentId}`,
      sender: "user",
      text: text,
      timestamp: "12:00",
    };

    setMessages((prev) => [...prev, userMessage]);
    if (!textToSend) setInputMsg("");
    setIsTyping(true);

    // Dynamic smart frontend simulator reply
    setTimeout(() => {
      messageCountRef.current += 1;
      const botId = messageCountRef.current;
      let replyText = "Terima kasih atas pertanyaannya! ChatSolv otomatis merespons pesan pelanggan dengan cerdas dan cepat 24/7.";
      
      const lower = text.toLowerCase();
      if (lower.includes("apa") || lower.includes("jelaskan") || lower.includes("chatsolv")) {
        replyText = "ChatSolv adalah platform otomatisasi Customer Service WhatsApp berbasis kecerdasan buatan. Kami membantu bisnis menjawab pesan pelanggan seketika dengan pemahaman SOP dan katalog produk Anda ✨";
      } else if (lower.includes("harga") || lower.includes("biaya") || lower.includes("langganan") || lower.includes("paket")) {
        replyText = "Paket langganan ChatSolv sangat fleksibel mulai dari bisnis skala UMKM hingga Enterprise dengan integrasi multi-nomor dan analytics real-time 📊";
      } else if (lower.includes("nomor") || lower.includes("wa") || lower.includes("whatsapp") || lower.includes("integrasi")) {
        replyText = "Bisa banget! Anda bisa langsung menghubungkan nomor WhatsApp bisnis yang sudah ada tanpa perlu ganti nomor baru 📱";
      } else if (lower.includes("kerja") || lower.includes("cara") || lower.includes("setup")) {
        replyText = "Cukup hubungkan nomor via QR scan, unggah data produk / FAQ bisnis Anda, dan AI ChatSolv akan langsung aktif melayani pelanggan secara otomatis 🚀";
      }

      const botMessage: Message = {
        id: `b-${botId}`,
        sender: "bot",
        text: replyText,
        timestamp: "12:00",
      };

      setMessages((prev) => [...prev, botMessage]);
      setIsTyping(false);
    }, 900);
  };

  // Split intro words for Blur Writer effect
  const introLine1 = "Pelanggan Tidak Menghilang Tiba-Tiba.".split(" ");
  const introLine2 = "Mereka Berhenti Menunggu.".split(" ");

  const comingSoonLine1 = "Fitur Baru Sedang Disiapkan.".split(" ");
  const comingSoonLine2 = "Coming Soon.".split(" ");

  return (
    <div className="fixed inset-0 w-screen h-screen overflow-hidden select-none bg-[#d6ebd8] text-[#162b19]">
      {/* Signature Sage Green Animated Fluid Mesh Background */}
      <GradientBlurBackground />

      {/* Structured Max-Width Container strictly aligned for Header & Center */}
      <div className="relative w-full h-full max-w-[1240px] mx-auto px-4 sm:px-10 flex flex-col justify-between py-5 z-10">
        
        {/* TOP NAVBAR */}
        <motion.header 
          initial={{ opacity: 0, y: -16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, ease: [0.22, 1, 0.36, 1] }}
          className="w-full flex items-center justify-between z-30 shrink-0 h-16 relative"
        >
          <div 
            onClick={() => setActiveStep(0)} 
            className="flex items-center gap-2 cursor-pointer group"
          >
            <Image
              src="/chatsolv-logo-transparent.png"
              alt="ChatSolv Official Logo"
              width={220}
              height={56}
              priority
              className="h-10 sm:h-12 w-auto object-contain transition-transform duration-200 group-hover:scale-105"
            />
          </div>

          <div className="flex items-center gap-2 sm:gap-3">
            <button
              onClick={() => setActiveStep(0)}
              className={`px-4 sm:px-5 py-2 rounded-full text-[13px] sm:text-[14px] font-extrabold tracking-tight transition-all duration-200 cursor-pointer select-none ${
                activeStep === 0
                  ? "bg-[#618264] text-white shadow-lg shadow-[#618264]/30 border border-[#49684d] scale-105"
                  : "text-[#1b3d20] bg-transparent hover:bg-white/80 hover:text-[#0e1c10]"
              }`}
            >
              Beranda
            </button>
            <button
              onClick={() => setActiveStep(1)}
              className={`px-4 sm:px-5 py-2 rounded-full text-[13px] sm:text-[14px] font-extrabold tracking-tight transition-all duration-200 cursor-pointer select-none ${
                activeStep === 1
                  ? "bg-[#618264] text-white shadow-lg shadow-[#618264]/30 border border-[#49684d] scale-105"
                  : "text-[#1b3d20] bg-transparent hover:bg-white/80 hover:text-[#0e1c10]"
              }`}
            >
              Demo Interaktif
            </button>
            <button
              onClick={() => setActiveStep(2)}
              className={`px-4 sm:px-5 py-2 rounded-full text-[13px] sm:text-[14px] font-extrabold tracking-tight transition-all duration-200 cursor-pointer select-none ${
                activeStep === 2
                  ? "bg-[#618264] text-white shadow-lg shadow-[#618264]/30 border border-[#49684d] scale-105"
                  : "text-[#1b3d20] bg-transparent hover:bg-white/80 hover:text-[#0e1c10]"
              }`}
            >
              Coming Soon
            </button>
          </div>
        </motion.header>

        {/* Dynamic Center Stage Viewport (Strictly Centered, Zero Offsets) */}
        <div className="w-full flex-1 flex items-center justify-center my-auto z-20 py-2">
          <AnimatePresence mode="wait">
            
            {/* ========================================================================= */}
            {/* 00 WELCOME HERO */}
            {/* ========================================================================= */}
            {activeStep === 0 && (
              <motion.div
                key="section-00-hero-welcome"
                variants={introContainerStagger}
                initial="hidden"
                animate="show"
                exit="exit"
                className="w-full max-w-[1140px] flex flex-col items-center justify-center text-center my-auto"
              >
                <motion.div 
                  variants={sentenceContainer}
                  className="w-full flex flex-col items-center justify-center gap-2"
                >
                  <h1 className="w-full flex flex-wrap sm:flex-nowrap items-center justify-center gap-x-2 sm:gap-x-3 text-[clamp(32px,4.5vw,66px)] font-extrabold text-[#0e1c10] tracking-[-0.035em] leading-[1.08] whitespace-normal sm:whitespace-nowrap">
                    {introLine1.map((word, idx) => (
                      <motion.span
                        key={`w1-${idx}`}
                        variants={wordBlurVariant}
                        className="inline-block"
                      >
                        {word}
                      </motion.span>
                    ))}
                  </h1>

                  <div className="w-full flex flex-wrap sm:flex-nowrap items-center justify-center gap-x-2 sm:gap-x-3 text-[clamp(32px,4.5vw,66px)] font-extrabold text-[#1a381d] tracking-[-0.035em] leading-[1.08] whitespace-normal sm:whitespace-nowrap">
                    {introLine2.map((word, idx) => (
                      <motion.span
                        key={`w2-${idx}`}
                        variants={wordBlurVariant}
                        className="inline-block"
                      >
                        {word}
                      </motion.span>
                    ))}
                  </div>
                </motion.div>

                <motion.p 
                  variants={itemVariants}
                  className="mt-6 text-base sm:text-[17px] text-[#253e28] font-medium leading-[1.5] max-w-[640px] mx-auto tracking-normal"
                >
                  ChatSolv membantu bisnis merespons lebih cepat, tetap konsisten, dan menjaga setiap pelanggan tetap terlayani.
                </motion.p>

                <motion.div 
                  variants={itemVariants}
                  className="mt-7 flex flex-col sm:flex-row items-center justify-center gap-4"
                >
                  <button 
                    onClick={() => setActiveStep(1)}
                    className="btn-3d-emerald h-[52px] px-8 rounded-full text-white font-extrabold text-[15px] flex items-center gap-2.5 cursor-pointer select-none group shadow-lg shadow-[#618264]/25"
                  >
                    <span>Coba Demo Interaktif</span>
                    <ArrowRight className="w-4 h-4 transition-transform group-hover:translate-x-1" />
                  </button>
                  <button 
                    onClick={() => setActiveStep(2)}
                    className="btn-3d-glass h-[52px] px-8 rounded-full text-[#162b19] font-extrabold text-[14px] flex items-center gap-2 cursor-pointer select-none"
                  >
                    <span>Coming Soon</span>
                  </button>
                </motion.div>
              </motion.div>
            )}

            {/* ========================================================================= */}
            {/* 01 DEMO CONVERSATION PAGE */}
            {/* ========================================================================= */}
            {activeStep === 1 && (
              <motion.div
                key="section-01-demo-conversation"
                variants={introContainerStagger}
                initial="hidden"
                animate="show"
                exit="exit"
                className="w-full max-w-[780px] h-[calc(100vh-140px)] max-h-[640px] flex flex-col my-auto rounded-[28px] bg-white/70 backdrop-blur-2xl border border-white/90 shadow-2xl shadow-[#618264]/15 overflow-hidden"
              >
                {/* Chat Top Header */}
                <div className="px-5 py-3.5 border-b border-[#618264]/20 flex items-center justify-between bg-white/50 shrink-0">
                  <div className="flex items-center gap-3">
                    <div className="relative">
                      <div className="w-10 h-10 rounded-full bg-gradient-to-tr from-[#618264] to-[#79AC78] flex items-center justify-center text-white shadow-sm">
                        <Bot className="w-5 h-5" />
                      </div>
                      <span className="absolute bottom-0 right-0 w-3 h-3 rounded-full bg-emerald-500 border-2 border-white"></span>
                    </div>
                    <div>
                      <div className="flex items-center gap-2">
                        <h3 className="text-sm sm:text-base font-extrabold text-[#0e1c10]">ChatSolv AI Assistant</h3>
                        <span className="px-2 py-0.5 rounded-full bg-[#618264]/15 text-[#1b3d20] text-[10px] font-bold">Demo</span>
                      </div>
                      <p className="text-xs text-[#49684d] font-medium flex items-center gap-1">
                        <Sparkles className="w-3 h-3 text-[#618264]" />
                        Respons Otomatis Aktif
                      </p>
                    </div>
                  </div>

                  <button
                    onClick={() => setMessages(INITIAL_MESSAGES)}
                    title="Reset Chat"
                    className="p-2 rounded-full hover:bg-white/80 text-[#2a452e] transition-colors cursor-pointer"
                  >
                    <RefreshCw className="w-4 h-4" />
                  </button>
                </div>

                {/* Message Stream Area */}
                <div className="chat-scroll-area flex-1 overflow-y-auto p-4 sm:p-5 space-y-4 select-text">
                  {messages.map((msg) => (
                    <motion.div
                      key={msg.id}
                      initial={{ opacity: 0, y: 10, scale: 0.98 }}
                      animate={{ opacity: 1, y: 0, scale: 1 }}
                      transition={{ duration: 0.2 }}
                      className={`flex gap-2.5 sm:gap-3 ${msg.sender === "user" ? "justify-end" : "justify-start"}`}
                    >
                      {msg.sender === "bot" && (
                        <div className="w-8 h-8 rounded-full bg-[#618264] text-white flex items-center justify-center shrink-0 mt-0.5 shadow-sm">
                          <Bot className="w-4 h-4" />
                        </div>
                      )}

                      <div
                        className={`max-w-[82%] sm:max-w-[70%] p-3.5 sm:p-4 rounded-2xl shadow-sm text-sm leading-relaxed ${
                          msg.sender === "user"
                            ? "bg-[#618264] text-white rounded-br-none font-medium shadow-md shadow-[#618264]/20"
                            : "bg-white/90 text-[#162b19] border border-[#B0D9B1]/40 rounded-bl-none font-medium"
                        }`}
                      >
                        <p className="whitespace-pre-line">{msg.text}</p>
                        <div
                          className={`mt-1.5 text-[10px] flex items-center justify-end gap-1 ${
                            msg.sender === "user" ? "text-emerald-100" : "text-[#618264]"
                          }`}
                        >
                          <span>{msg.timestamp}</span>
                          {msg.sender === "user" && <CheckCheck className="w-3 h-3" />}
                        </div>
                      </div>

                      {msg.sender === "user" && (
                        <div className="w-8 h-8 rounded-full bg-white text-[#162b19] border border-[#B0D9B1]/50 flex items-center justify-center shrink-0 mt-0.5 shadow-sm">
                          <User className="w-4 h-4" />
                        </div>
                      )}
                    </motion.div>
                  ))}

                  {isTyping && (
                    <motion.div
                      initial={{ opacity: 0, y: 6 }}
                      animate={{ opacity: 1, y: 0 }}
                      className="flex gap-2.5 items-center text-xs text-[#49684d] font-semibold pl-1"
                    >
                      <div className="w-7 h-7 rounded-full bg-[#618264]/20 text-[#162b19] flex items-center justify-center">
                        <Bot className="w-3.5 h-3.5" />
                      </div>
                      <div className="px-3.5 py-2 rounded-full bg-white/80 border border-[#B0D9B1]/40 flex items-center gap-1.5 shadow-sm">
                        <span className="w-1.5 h-1.5 rounded-full bg-[#618264] animate-bounce"></span>
                        <span className="w-1.5 h-1.5 rounded-full bg-[#618264] animate-bounce [animation-delay:0.2s]"></span>
                        <span className="w-1.5 h-1.5 rounded-full bg-[#618264] animate-bounce [animation-delay:0.4s]"></span>
                      </div>
                    </motion.div>
                  )}
                  <div ref={chatBottomRef} />
                </div>

                {/* Preset Suggestions */}
                <div className="px-4 py-2 bg-white/40 border-t border-[#618264]/10 flex items-center gap-2 overflow-x-auto no-scrollbar shrink-0">
                  {PRESET_PROMPTS.map((prompt, idx) => (
                    <button
                      key={idx}
                      onClick={() => handleSendMessage(prompt)}
                      className="shrink-0 text-xs px-3 py-1.5 rounded-full bg-white/80 hover:bg-white text-[#1b3d20] border border-[#B0D9B1]/60 font-semibold transition-all hover:scale-105 active:scale-95 cursor-pointer shadow-xs"
                    >
                      {prompt}
                    </button>
                  ))}
                </div>

                {/* Chat Input Bar */}
                <form
                  onSubmit={(e) => {
                    e.preventDefault();
                    handleSendMessage();
                  }}
                  className="p-3 sm:p-4 bg-white/70 border-t border-[#618264]/20 flex items-center gap-2.5 shrink-0"
                >
                  <input
                    type="text"
                    value={inputMsg}
                    onChange={(e) => setInputMsg(e.target.value)}
                    placeholder="Ketik pesan untuk memulai chat demo..."
                    className="flex-1 bg-white/95 border border-[#B0D9B1]/60 rounded-full px-4 sm:px-5 py-3 text-sm text-[#0e1c10] placeholder:text-[#618264]/60 focus:outline-none focus:ring-2 focus:ring-[#618264]/40 font-medium shadow-inner"
                  />
                  <button
                    type="submit"
                    disabled={!inputMsg.trim()}
                    className="w-11 h-11 rounded-full bg-[#618264] hover:bg-[#527055] disabled:opacity-40 disabled:hover:bg-[#618264] text-white flex items-center justify-center transition-all cursor-pointer shrink-0 shadow-md shadow-[#618264]/30"
                  >
                    <Send className="w-4 h-4" />
                  </button>
                </form>
              </motion.div>
            )}

            {/* ========================================================================= */}
            {/* 02 COMING SOON PAGE */}
            {/* ========================================================================= */}
            {activeStep === 2 && (
              <motion.div
                key="section-02-coming-soon"
                variants={introContainerStagger}
                initial="hidden"
                animate="show"
                exit="exit"
                className="w-full max-w-[1140px] flex flex-col items-center justify-center text-center my-auto"
              >
                <motion.div 
                  variants={sentenceContainer}
                  className="w-full flex flex-col items-center justify-center gap-2"
                >
                  <h1 className="w-full flex flex-wrap sm:flex-nowrap items-center justify-center gap-x-2 sm:gap-x-3 text-[clamp(32px,4.5vw,66px)] font-extrabold text-[#0e1c10] tracking-[-0.035em] leading-[1.08] whitespace-normal sm:whitespace-nowrap">
                    {comingSoonLine1.map((word, idx) => (
                      <motion.span
                        key={`cs1-${idx}`}
                        variants={wordBlurVariant}
                        className="inline-block"
                      >
                        {word}
                      </motion.span>
                    ))}
                  </h1>

                  <div className="w-full flex flex-wrap sm:flex-nowrap items-center justify-center gap-x-2 sm:gap-x-3 text-[clamp(32px,4.5vw,66px)] font-extrabold text-[#1a381d] tracking-[-0.035em] leading-[1.08] whitespace-normal sm:whitespace-nowrap">
                    {comingSoonLine2.map((word, idx) => (
                      <motion.span
                        key={`cs2-${idx}`}
                        variants={wordBlurVariant}
                        className="inline-block"
                      >
                        {word}
                      </motion.span>
                    ))}
                  </div>
                </motion.div>

                <motion.p 
                  variants={itemVariants}
                  className="mt-6 text-base sm:text-[17px] text-[#253e28] font-medium leading-[1.5] max-w-[640px] mx-auto tracking-normal"
                >
                  Kami sedang menyiapkan generasi berikutnya untuk otomatisasi komunikasi bisnis Anda.
                </motion.p>

                <motion.div 
                  variants={itemVariants}
                  className="mt-7 flex flex-col sm:flex-row items-center justify-center gap-4"
                >
                  <button 
                    onClick={() => setActiveStep(0)}
                    className="btn-3d-emerald h-[52px] px-8 rounded-full text-white font-extrabold text-[15px] flex items-center gap-2.5 cursor-pointer select-none group"
                  >
                    <span>Kembali ke Beranda</span>
                    <ArrowRight className="w-4 h-4 transition-transform group-hover:translate-x-1" />
                  </button>
                </motion.div>
              </motion.div>
            )}

          </AnimatePresence>
        </div>

      </div>
    </div>
  );
}
