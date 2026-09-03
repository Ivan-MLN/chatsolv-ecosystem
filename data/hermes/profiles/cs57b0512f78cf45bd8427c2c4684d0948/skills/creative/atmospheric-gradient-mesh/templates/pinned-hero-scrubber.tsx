"use client";

import React, { useState, useEffect, useRef } from "react";
import { motion, AnimatePresence, type Variants } from "framer-motion";
import { ChevronDown, ArrowRight } from "lucide-react";

/**
 * Pinned Hero Scrubber Template
 * 
 * High-performance, in-place hero presentation with discrete phase stepping,
 * fluid atmospheric gradient blur background, and staggered cascading entrances.
 * Clean centered intro with large typography and zero stray artifacts below CTAs.
 */

// Stagger variant for fresh intro view
export const introContainerStagger: Variants = {
  hidden: { opacity: 0 },
  show: {
    opacity: 1,
    transition: {
      staggerChildren: 0.1,
      delayChildren: 0.05,
    }
  },
  exit: {
    opacity: 0,
    transition: {
      duration: 0.2,
      ease: "easeInOut"
    }
  }
};

// Stagger variant for snappy inter-phase scroll transitions
export const phaseContainerStagger: Variants = {
  hidden: { opacity: 0 },
  show: {
    opacity: 1,
    transition: {
      staggerChildren: 0.08,
      delayChildren: 0.05,
    }
  },
  exit: {
    opacity: 0,
    transition: {
      duration: 0.2,
      ease: "easeInOut"
    }
  }
};

export const itemPop: Variants = {
  hidden: { opacity: 0, y: 16, scale: 0.97, filter: "blur(4px)" },
  show: { 
    opacity: 1, 
    y: 0, 
    scale: 1, 
    filter: "blur(0px)",
    transition: {
      type: "spring",
      stiffness: 280,
      damping: 24,
      mass: 0.7
    }
  },
  exit: { 
    opacity: 0, 
    y: -12, 
    scale: 0.98, 
    filter: "blur(2px)",
    transition: { duration: 0.18, ease: "easeOut" }
  }
};

export function PinnedHeroScrubber() {
  const [activeStep, setActiveStep] = useState(0);
  const isTransitioningRef = useRef(false);

  // Auto-reset scroll position on refresh
  useEffect(() => {
    if (typeof window !== "undefined") {
      window.history.scrollRestoration = "manual";
      window.scrollTo(0, 0);
    }
  }, []);

  // Intercept wheel, touch, and keyboard events for in-place step navigation
  useEffect(() => {
    const handleWheel = (e: WheelEvent) => {
      e.preventDefault();
      if (isTransitioningRef.current) return;

      if (e.deltaY > 20) {
        setActiveStep((prev) => {
          if (prev < 4) {
            isTransitioningRef.current = true;
            setTimeout(() => { isTransitioningRef.current = false; }, 450);
            return prev + 1;
          }
          return prev;
        });
      } else if (e.deltaY < -20) {
        setActiveStep((prev) => {
          if (prev > 0) {
            isTransitioningRef.current = true;
            setTimeout(() => { isTransitioningRef.current = false; }, 450);
            return prev - 1;
          }
          return prev;
        });
      }
    };

    window.addEventListener("wheel", handleWheel, { passive: false });
    return () => window.removeEventListener("wheel", handleWheel);
  }, []);

  return (
    <div className="fixed inset-0 w-screen h-screen overflow-hidden select-none bg-[#d6ebd8] text-[#162b19]">
      {/* Background Gradient Mesh Component */}
      <div className="fixed inset-0 pointer-events-none -z-10 overflow-hidden bg-[#d6ebd8]">
        <div className="absolute inset-0 bg-[radial-gradient(rgba(97,130,100,0.22)_1px,transparent_0)] bg-[size:28px_28px] opacity-80" />
      </div>

      <div className="relative w-full h-full flex flex-col justify-between py-6 px-6 md:px-12 max-w-7xl mx-auto z-10">
        {/* Top Navbar */}
        <motion.header
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, ease: [0.22, 1, 0.36, 1] }}
          className="w-full flex items-center justify-between z-20 shrink-0"
        >
          <div onClick={() => setActiveStep(0)} className="flex items-center gap-3 cursor-pointer">
            <div className="w-10 h-10 rounded-2xl bg-[#618264] text-white flex items-center justify-center font-bold text-lg shadow-md">
              C
            </div>
            <span className="font-extrabold text-2xl tracking-tight">BrandName</span>
          </div>

          <div className="flex items-center gap-2 px-4 py-2 rounded-full bg-white/60 border border-white/80 text-xs font-mono font-bold shadow-sm">
            {[0, 1, 2, 3, 4].map((step) => (
              <button
                key={step}
                onClick={() => setActiveStep(step)}
                className={`transition-all duration-300 rounded-full cursor-pointer ${
                  activeStep === step ? "w-7 h-2 bg-[#618264]" : "w-2 h-2 bg-[#79AC78]/40 hover:bg-[#79AC78]"
                }`}
              />
            ))}
          </div>
        </motion.header>

        {/* Dynamic Stepped Content */}
        <div className="w-full flex-1 flex items-center justify-center my-auto py-4 z-20 overflow-hidden">
          <AnimatePresence mode="wait">
            {activeStep === 0 ? (
              <motion.div
                key="step-0-intro"
                variants={introContainerStagger}
                initial="hidden"
                animate="show"
                exit="exit"
                className="max-w-4xl flex flex-col items-center text-center gap-6 px-4 my-auto"
              >
                <motion.div variants={itemPop} className="px-4 py-1.5 rounded-full bg-white/60 border border-white/80 text-xs font-mono font-bold tracking-wider uppercase">
                  01 — HERO
                </motion.div>
                <motion.h1 variants={itemPop} className="text-4xl sm:text-6xl md:text-7xl font-extrabold tracking-tight leading-[1.1]">
                  Commanding Display Title
                </motion.h1>
                <motion.div variants={itemPop} className="space-y-3 text-base sm:text-lg md:text-xl font-medium leading-relaxed max-w-2xl">
                  <p>Clean, proportional body narrative with generous line-height and balanced weight.</p>
                </motion.div>
                <motion.div variants={itemPop} className="flex items-center gap-4 pt-2">
                  <button className="px-8 py-3.5 rounded-2xl bg-[#618264] text-white font-extrabold text-base shadow-xl flex items-center gap-2">
                    <span>Primary Action</span>
                    <ArrowRight className="w-5 h-5" />
                  </button>
                </motion.div>
              </motion.div>
            ) : (
              <motion.div
                key={`step-${activeStep}`}
                variants={phaseContainerStagger}
                initial="hidden"
                animate="show"
                exit="exit"
                className="w-full grid grid-cols-1 lg:grid-cols-12 gap-8 items-center my-auto"
              >
                {/* Left Narrative */}
                <div className="lg:col-span-6 flex flex-col gap-4">
                  <motion.h2 variants={itemPop} className="text-3xl sm:text-5xl font-extrabold">
                    Phase {activeStep} Narrative
                  </motion.h2>
                  <motion.p variants={itemPop} className="text-base sm:text-lg opacity-80">
                    Step-by-step feature breakdown with high-fidelity interactive proof on the right.
                  </motion.p>
                </div>
                {/* Right Interactive Card */}
                <motion.div variants={itemPop} className="lg:col-span-6 flex items-center justify-center">
                  <div className="w-full max-w-md p-6 rounded-3xl bg-white/60 border border-white/80 shadow-xl">
                    <span className="font-mono text-xs font-bold text-[#618264]">Phase {activeStep} Interactive Module</span>
                  </div>
                </motion.div>
              </motion.div>
            )}
          </AnimatePresence>
        </div>

        {/* Footer Navigation Cue */}
        <footer className="w-full flex items-center justify-between text-xs font-mono text-[#618264] pt-3 border-t border-[#B0D9B1]/30 z-20 shrink-0">
          <div onClick={() => setActiveStep((prev) => (prev < 4 ? prev + 1 : 0))} className="flex items-center gap-2 cursor-pointer">
            <ChevronDown className="w-4 h-4 animate-bounce" />
            <span>SCROLL / CLICK TO ADVANCE</span>
          </div>
          <span className="font-bold">ChatSolv Dynamic Hero System</span>
        </footer>
      </div>
    </div>
  );
}
