"use client";

import React from "react";

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
    </div>
  );
}
