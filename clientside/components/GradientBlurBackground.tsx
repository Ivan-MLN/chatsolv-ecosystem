"use client";

import React from "react";
import { motion } from "framer-motion";

export function GradientBlurBackground() {
  return (
    <div className="fixed inset-0 pointer-events-none -z-10 overflow-hidden bg-[#d6ebd8]">
      {/* 1. Base Ambient Animated Fluid Mesh Blobs */}
      <div className="absolute inset-0 bg-noise opacity-80" />

      {/* Blob 1 - Top Left */}
      <div className="animate-blob-1 absolute -top-16 -left-16 w-[750px] h-[750px] rounded-full bg-gradient-to-tr from-[#618264] via-[#79AC78] to-[#B0D9B1] blur-[75px] opacity-95" />

      {/* Blob 2 - Center Right */}
      <div className="animate-blob-2 absolute top-1/6 -right-24 w-[800px] h-[800px] rounded-full bg-gradient-to-bl from-[#618264] via-[#79AC78] to-[#B0D9B1] blur-[80px] opacity-95" />

      {/* Blob 3 - Bottom Left */}
      <div className="animate-blob-3 absolute -bottom-24 -left-16 w-[820px] h-[820px] rounded-full bg-gradient-to-t from-[#618264] via-[#79AC78] to-[#D0E7D2] blur-[85px] opacity-90" />

      {/* Blob 4 - Bottom Right */}
      <div className="animate-blob-4 absolute -bottom-20 -right-20 w-[720px] h-[720px] rounded-full bg-gradient-to-tl from-[#79AC78] via-[#B0D9B1] to-[#618264] blur-[75px] opacity-90" />

      {/* Center Dynamic Swirling Core */}
      <div className="animate-blob-center absolute top-1/2 left-1/2 w-[650px] h-[650px] rounded-full bg-gradient-to-r from-[#79AC78]/70 via-[#B0D9B1]/80 to-[#D0E7D2]/80 blur-[70px] pointer-events-none" />

      {/* 2. Deep Sage Dark Atmosphere: Bertahan tenang saat teks & tombol berurutan muncul, lalu terangkat mulus 100% */}
      <motion.div
        initial={{ opacity: 1 }}
        animate={{ 
          opacity: [1, 1, 0.7, 0.25, 0],
          backgroundColor: ["#0f1d12", "#0f1d12", "#19331f", "#2f5236", "#d6ebd8"]
        }}
        transition={{ 
          duration: 2.2, 
          times: [0, 0.35, 0.65, 0.85, 1],
          ease: [0.25, 1, 0.4, 1] 
        }}
        className="absolute inset-0 z-20 pointer-events-none"
      />

      {/* 3. Spaceship-Inspired Atmospheric Overhead Emitter & Ambient Aurora Sweep (Smooth Ellipse, Zero Geometric Hard Cutoffs) */}
      <motion.div
        initial={{ opacity: 0, scale: 0.6 }}
        animate={{ 
          opacity: [0, 0.9, 1, 0.35, 0],
          scale: [0.6, 1, 1.4, 2, 2.6]
        }}
        transition={{ 
          duration: 2.2, 
          times: [0, 0.25, 0.55, 0.85, 1],
          ease: [0.22, 1, 0.36, 1] 
        }}
        className="absolute -top-32 left-1/2 -translate-x-1/2 w-[900px] h-[650px] rounded-[100%] bg-gradient-to-b from-[#B0D9B1] via-[#79AC78]/60 to-transparent blur-[90px] z-25 pointer-events-none"
      />

      {/* 4. Soft Center-Weighted Ambient Flare */}
      <motion.div
        initial={{ opacity: 0, scale: 0.8 }}
        animate={{ 
          opacity: [0, 0.7, 0.9, 0.25, 0],
          scale: [0.8, 1.1, 1.6, 2.2]
        }}
        transition={{ 
          duration: 2.2, 
          times: [0, 0.25, 0.6, 1],
          ease: "easeOut"
        }}
        className="absolute top-1/4 left-1/2 -translate-x-1/2 w-[750px] h-[450px] rounded-full bg-gradient-to-b from-[#D0E7D2]/80 via-[#B0D9B1]/40 to-transparent blur-[85px] z-25 pointer-events-none"
      />

      <div className="absolute inset-0 pointer-events-none" />
    </div>
  );
}
