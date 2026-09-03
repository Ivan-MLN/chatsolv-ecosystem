---
name: atmospheric-gradient-mesh
description: Create smooth animated gradient blur mesh backgrounds.
version: 1.0.0
author: Hermes Agent
license: MIT
tags: [design, css, animation, background, mesh-gradient, framer-motion, tailwindcss]
platforms: [linux, macos, windows]
triggers:
  - create gradient blur background
  - smooth gradient mesh
  - animated background blur
  - soft palette atmospheric background
  - sage green gradient
---

# Atmospheric Gradient Mesh Backgrounds

Guidelines and patterns for crafting high-fidelity, dynamic gradient blur mesh backgrounds in modern web apps (Tailwind CSS + Framer Motion / CSS keyframes).

## Core Principles

### 1. Palette Saturation & Base Canvas (Avoid Washed-Out Light Mode)
- In light-mode atmospheric designs, **never blend solely against pure white (`#ffffff`)** or overly washed-out tints. High blur radius (120px+) diffuses color quickly, making soft tones look completely white or unrendered.
- **Tint the base canvas**: Use a tinted base (e.g. pale sage `#d6ebd8` instead of pure white `#ffffff`).
- **Use contrasting blob anchors**: Pair soft tints (`#B0D9B1`, `#D0E7D2`) with saturated anchor colors (`#618264`, `#79AC78`) at 80%–90% opacity.
- **Keep blur radii balanced**: Blur between `80px` and `100px` for defined color presence; `150px+` risks washing out the hue completely.

### 2. Composition & Mesh Layout
- **Multi-node floating mesh**: Use 3 to 4 distinct floating blobs placed along perimeters and corners with asynchronous floating animation durations (16s, 20s, 22s).
- **Center spotlight/wash**: Provide a gentle radial gradient spotlight in the center behind typography for crisp contrast and readability without wiping out surrounding colors.
- **Texture layer**: Add a subtle dot-matrix (`radial-gradient`) or noise overlay at 70%–85% opacity to eliminate digital color banding (quantization) and give an organic, tactile feel.

### 3. Motion Dynamics & Perceived Speed
- **Animation cycle speed**: Durations longer than 15s often appear static/stalled to users. Aim for **6s to 9s cycles** with wide travel distance (±150px to ±260px) and rotation (`0deg` to `360deg` / scale `0.8x` to `1.35x`) so the animation is immediately clear and lively.
- **Fluid GPU keyframes vs React state**: Prefer pure CSS keyframes with `translate3d()`, `will-change: transform`, and smooth easing curve `cubic-bezier(0.45, 0.05, 0.55, 0.95)` for guaranteed 60fps jitter-free motion.
- **Center swirling core**: Include a central swirling node that rotates and scales asynchronously with corner blobs to keep the viewport active.

- **Contrast & Foreground Text Readability on Colored Gradients**:
  - **Avoid White Drop Shadows / Halos on Dark Text**: Never apply white or light outer drop-shadows (`drop-shadow-[0_1px_2px_rgba(255,255,255,0.7)]`) behind dark text over colored/sage gradients, as it creates a dirty, blurry halo artifact. Instead, use pure solid high-contrast dark charcoal / forest green (`#0e1c10`, `#162b19`) for razor-sharp typography.
  - **Gradient text on gradient backgrounds**: Avoid matching text gradient hues too closely with background mesh colors (e.g. mid-sage text over mid-sage canvas).
  - **Use deep anchor shades**: Pick high-contrast terminal tones (e.g. `#1b3d20` to `#122b15`) for highlighted text headings so they maintain crisp WCAG AAA legibility.

## Recommended Code Structure (React + Framer Motion / CSS Keyframes + Tailwind)

```tsx
"use client";

import { motion } from "framer-motion";
import React from "react";

export function GradientBlurBackground() {
  return (
    <div className="fixed inset-0 pointer-events-none -z-10 overflow-hidden bg-[#d6ebd8]">
      {/* Dot-matrix grid texture */}
      <div className="absolute inset-0 bg-[radial-gradient(rgba(97,130,100,0.2)_1px,transparent_0)] bg-[size:28px_28px] opacity-85" />

      {/* Top ambient wash */}
      <div className="absolute -top-[25%] left-1/2 -translate-x-1/2 w-[1100px] h-[650px] bg-gradient-to-b from-[#79AC78] via-[#B0D9B1] to-transparent rounded-full blur-[100px] opacity-90" />

      {/* Floating Blob 1 - Corner Anchor */}
      <motion.div
        animate={{
          x: [0, 80, -50, 0],
          y: [0, -60, 40, 0],
          scale: [1, 1.25, 0.9, 1],
        }}
        transition={{ duration: 16, repeat: Infinity, ease: "easeInOut" }}
        className="absolute -top-20 -left-20 w-[650px] h-[650px] rounded-full bg-gradient-to-tr from-[#618264] via-[#79AC78] to-[#B0D9B1] blur-[90px] opacity-85"
      />

      {/* Floating Blob 2 - Mid Accent */}
      <motion.div
        animate={{
          x: [0, -90, 60, 0],
          y: [0, 70, -50, 0],
          scale: [1, 0.85, 1.2, 1],
        }}
        transition={{ duration: 20, repeat: Infinity, ease: "easeInOut", delay: 1 }}
        className="absolute top-1/4 -right-24 w-[700px] h-[700px] rounded-full bg-gradient-to-bl from-[#618264] via-[#79AC78] to-[#B0D9B1] blur-[95px] opacity-90"
      />

      {/* Center luminous aura for high-contrast typography */}
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_40%,rgba(208,231,210,0.5)_0%,transparent_70%)] pointer-events-none" />
    </div>
  );
}
```

### 5. Staggered Cascading Entrances & Scroll Hero Patterns
- **Hero Staggering (Self-Assembling UI)**: When creating sticky scroll hero scrubbers or dynamic viewports, wrap elements in Framer Motion spring variants.
- **Differentiated Stagger Rates (Fresh Load vs Scroll Transitions)**:
  - **Fresh Page Load (Index 0)**: Use a deliberate, luxurious entrance timing (`staggerChildren: 0.12s`, `delayChildren: 0.1s`) so the user clearly sees elements cascade in one by one.
  - **Phase Scroll/Step Transitions (Index 1..N)**: Keep inter-phase transitions **balanced & responsive** (`staggerChildren: 0.08s`, `delayChildren: 0.05s`, `exit duration: 0.2s`, transition cooldown `480ms`). Do not make scroll phase transitions too slow (feels laggy) nor too instantaneous/abrupt (feels jarring and loses the self-assembling stagger effect).
- **Density & Content Richness on Subsequent Phases (Phases 2..N)**:
  - Avoid sparse or empty-looking visual cards on later steps. Every phase should showcase rich, domain-dense artifacts:
    - *Phase 1*: Live chat bubble flow + latency/reply stats.
    - *Phase 2*: Multi-source RAG pipelines (DB sync % with SKU counts, strict SOP guardrail status, live logistics shipping API, vector embedding dimensions like Ada-002 1536 / cosine similarity).
    - *Phase 3*: Live agent escalation desks (VIP intent banners, assigned agent status, wait times <5s, 1-click takeover action buttons, VOIP/chat escalation indicators).
    - *Phase 4*: Multi-dimensional growth analytics (7-day daily volume bar charts with accent highlights, 3-grid KPI cards for sentiment/leads/closing rate, automated CSV/webhook pipelines).
  - Balance right-hand dashboard density by providing a 3-chip metric row on the left column across all steps.
- **Architectural & Layout Variety Across Step Sequences**:
  - **Avoid Uniform/Repetitive 2-Column Boxes**: Never repeat the exact same layout (Left: Text, Right: Rounded Card) across all sequential steps, which makes multi-step presentations feel monotonous.
  - **Vary Alignment & Component Structure per Phase**:
    - *Step 0 (Intro)*: Centered display hero canvas with symmetric feature pills.
    - *Step 1 (Simulation)*: Asymmetric split with realistic device hardware frame (e.g. glass smartphone container with camera notch, multi-turn looping chat dialogues, animated typing indicators, and send inputs).
    - *Step 2 (Data Flow / Architecture)*: **Inverted Layout** (Visual Multi-Node Network Topology on the Left, Technical Narrative on the Right).
    - *Step 3 (Workflow / Escalation)*: **3-Card Horizontal Pipeline / Command Deck** with centered header (Deal Queue -> AI Briefing -> Admin Action). Ensure perfect symmetrical alignment (equal-height stretched cards, centered header stack, horizontal footer baseline).
    - *Step 4 (Analytics / Metrics)*: **Asymmetrical High-Tech Bento Grid** (~50% width Bar Chart, dual KPI pods, and spanning bottom webhook integration strip).
  - Changing composition and alignment across phases gives the interface a dynamic, art-directed rhythm while keeping the color system and motion physics cohesive.
- **Dynamic Looping Simulator Animation & Auto-Scroll**:
  - For product simulation cards (e.g. WhatsApp chat, terminal CLI, VoIP handoff), implement an active auto-looping animation engine using `useEffect` intervals and state slicing (`visibleMessageCount` + `isTyping` state + multi-dialogue scenarios). This makes the demo feel alive rather than static text.
  - **Auto-Scroll Behavior for Active Chat Streams**: In simulated phone or chat viewports, new incoming bubbles and typing indicators will drown or clip at the bottom unless an auto-scroller is wired. Attach a `ref` to the scrollable message list and run `chatScrollContainerRef.current.scrollTo({ top: chatScrollContainerRef.current.scrollHeight, behavior: "smooth" })` inside a `useEffect` keyed on `[visibleMessageCount, isTyping, chatConvIndex]`. Style the list with `overflow-y-auto`, `max-h-[300px]`, `scroll-smooth`, and hide default scrollbars (`scrollbarWidth: "none"`) to preserve a pristine native app look.
- **Strict Verbatim Landing Page Implementation (Client Material Locked)**:
  - When the user supplies complete copywriting/marketing copy for a landing page (e.g. Hero, The Hidden Problem, How It Works, Control, Result, Footer) with instructions not to alter any wording, reproduce the text **100% verbatim** across all titles, descriptions, punchlines, and micro-copy.
  - Pair the exact copy with art-directed atmospheric visuals (e.g. dark obsidian canvas with subtle emerald/mint glow, conversation bottleneck funnels, horizontal flow node animations, minimal control surface status panels, and serene payoff streams) instead of generic marketing templates, ensuring zero AI buzzwords or altered semantics.
- **Entrance Pop Variants**: Animate each element from `{ opacity: 0, y: 18, scale: 0.96, filter: "blur(4px)" }` to `{ opacity: 1, y: 0, scale: 1, filter: "blur(0px)" }` using high-stiffness spring dynamics (`stiffness: 340`, `damping: 24`, `mass: 0.6`).
- **Initial Welcome View vs Split-Screen Scrubber**:
  - **Index 0 (Welcome / Intro)**: Fully centered layout (Badge → Big Headline → Subheadline → CTAs → 4 Matrix feature cards) to establish emotional anchor and value proposition.
  - **Index 1..N (Interactive Phases)**: Transition into a 2-column split-screen (Left: Narrative details + key metrics; Right: Live interactive simulator/visual cards) driven by viewport scroll progress.
- **Scroll Hijacking vs Natural Document Flow vs Pinned Hero Scrubber**: 
  - **Pinned Hero Scrubber Pattern**: When the user wants the entire hero section pinned (scrolling changes text/visual states in-place without scrolling down the document body), use a `fixed inset-0 overflow-hidden` viewport driven by discrete state (`activeStep: 0..N`).
  - **State Transition Handling**: Intercept `wheel`, `touchmove`, and `keydown` events with a debounce guard (e.g. 350ms transition lock) and provide interactive pagination pills (`INTRO`, `01`, `02`...) so users can advance by click, wheel, or swipe.
  - **Hydration & SSR Opacity Safety**: When combining `AnimatePresence mode="wait"` with `staggerChildren` and `initial="hidden"`, ensure client components properly animate to `show` upon mount. Never leave SSR HTML hardcoded at `opacity: 0` without client hydration triggers, or the page will appear completely blank on first render.
  - **Auto-Reset on Refresh**: Embed `window.history.scrollRestoration = "manual"` and `window.scrollTo(0, 0)` in `useEffect` so refreshing always re-triggers the initial Welcome/Intro entrance animation from the top.
- **Pinned Hero Scrubber Template**: Linked starter template in `templates/pinned-hero-scrubber.tsx` demonstrating the complete pinned hero scrubber pattern with event listeners, stagger variants, and auto-scroll handling.
- **Hero Art Direction & Transitions Reference**: See `references/hero-art-direction-and-transitions.md` for deep design guidelines on atmospheric overhead lighting, 3D tactile glass pill buttons, in-place zoom navigation, borderless hamburger-to-X morphs, and solid shadow-free typography on colored gradients.
- **Production Build vs Dev Server for Complex Framer Motion**: In Next.js 15/16 + Turbopack, heavy client animation bundles can occasionally fail to mount smoothly under `next dev`. Always run `next build` -> `next start` (or verify production bundle) to ensure all static segments and Turbopack chunks pre-render with immediate element visibility.

- **Single-Sentence Line Formatting & Word-Level Blur Writer Reveals**:
  - **Single Sentence per Line Layout**: When a headline consists of distinct sentence units, format them so that each sentence occupies its own dedicated line without mid-sentence wrapping (e.g. Line 1: `Pelanggan Tidak Menghilang Tiba-Tiba.`, Line 2: `Mereka Berhenti Menunggu.`) by calibrating responsive typography sizing (`text-2xl sm:text-4xl md:text-5xl lg:text-[3.4rem] whitespace-normal sm:whitespace-nowrap`) and viewport max-widths.
  - **Word-by-Word Blur Writer Animation**: For cinematic headline reveals, split sentences by words and render each word as a `motion.span` with stagger timing (`staggerChildren: 0.08s`) animating from `{ opacity: 0, filter: "blur(12px)", y: 10, scale: 0.95 }` to `{ opacity: 1, filter: "blur(0px)", y: 0, scale: 1 }` via a responsive spring (`stiffness: 220, damping: 20`).
  - **No Cluttered Numbers / Stray Labels**: Do not plaster artificial numeric step prefixes (`01 -`, `02 -`) across every section unless explicitly demanded. Let clean typography, structural spacing, and natural narrative progression guide the user.

- **Typography Scale & Layout Cleanliness in Centered Hero Views**:
  - **Avoid Cramped / Small Font Hierarchy**: When the user requests a prominent centered hero, use large, commanding display scales:
    - *Display Headline (H1)*: `text-4xl sm:text-6xl md:text-7xl font-extrabold tracking-tight leading-[1.08] sm:leading-[1.1]` with high-contrast text.
    - *Body Copy*: `text-base sm:text-lg md:text-xl font-medium leading-relaxed max-w-2xl` using dedicated fonts (like *Plus Jakarta Sans* `var(--font-pjs)` or *Geist*) rather than standard small body text.
    - *Primary CTA Buttons*: `px-8 py-3.5 text-base font-extrabold rounded-2xl` (height ~50px).
  - **Eliminate Stray/Stacked Artifacts Below Centered Intro CTAs**: The initial Hero/Intro canvas must remain completely distraction-free. Do not place random floating preview cards, extra text blocks, or secondary chat widgets beneath the centered CTA buttons on Step 0—keep the viewport clean with only: Top Badge → Dominant Headline → Proportional Narrative → CTA Buttons. Subsequent detail/chat widgets belong strictly in the sequential workflow phases (Steps 1..N).

- **Hero Art-Direction, Optical Alignment & Layout Balance**:
  - **Optical Vertical Centering (Upper-Middle Weighting)**: Do NOT mathematically center hero content against the entire 100vh viewport (which creates an optically bottom-heavy feel). Shift the central hero content ~20–35px upward (e.g. `-translate-y-6`) into the usable space between navbar and bottom divider/footer.
  - **2-Line Headline Composition**: Structure display headlines so Line 1 is naturally wider than Line 2 (creating a grounded, anchored visual triangle). Avoid mid-sentence wrapping by sizing with responsive fluid typography (`clamp(32px, 4.5vw, 68px)`) and `whitespace-nowrap` on desktop.
  - **Word-by-Word Blur Writer Reveal**: For smooth editorial entrances, split sentences into words (`sentence.split(" ")`) and animate each word with a spring-driven blur-to-sharp variant (`filter: "blur(10px)"` → `filter: "blur(0px)"`, `staggerChildren: 0.07s`).
  - **Single Supporting Sentence**: Place one subtle supporting sentence below the headline (max-width 600–680px, font 16–18px, line-height 1.5, dark muted green-gray) with a controlled 24–28px gap above and 28–32px gap below to the dual CTAs.
  - **Shared Container & Baseline Harmony**: Use one unified max-width container (e.g. `max-w-[1180px]` or `max-w-[1240px]`) shared across Navbar, Central Hero, and Footer. Ensure Navbar elements (logo vs progress indicator) and Footer elements (scroll hint vs copyright) align to strict optical baselines.
  - **No Clutter / Unnecessary Numbering**: Avoid artificial `01 - ...`, `02 - ...` badges across every screen. Clean typography, spacing rhythm, and natural step transitions look significantly more premium and human-designed.
  - **Avoid Horizontal Clipping & Invisible Border Traps**: Do not place `overflow-hidden` on inner text wrappers or apply harsh radial edge masks that clip the outermost glyphs (e.g. first letter on the left or ending periods on the right). Ensure ample container padding (`px-4 sm:px-8`).
  - **3D Tactile Glass Pill Buttons (Oval Glassmorphism 2.0)**:
    - Shape: Full oval capsule (`rounded-full` with `h-[50px]`–`h-[52px]` and `px-7`–`px-8`).
    - Surface Depth & Lighting:
      - *Primary (Emerald/Accent)*: Top specular highlight (`border-top: 1px solid rgba(255,255,255,0.45)`), tactile bottom shelf (`border-bottom: 3px solid #3b533e`), vertical gradient (`linear-gradient(180deg, #6d9370 0%, #527355 100%)`), and soft diffused drop shadow (`0 8px 20px -6px rgba(...)`).
      - *Secondary (Glass/Frosted)*: Translucent background (`backdrop-filter: blur(20px)`), glossy top highlight (`border-top: 1px solid rgba(255,255,255,0.95)`), soft mint lower border (`border-bottom: 3px solid #b6d3b8`), and inset top glow (`inset 0 1px 1px #fff`).
    - Micro-Interactions:
      - *Hover*: Smooth elevation `translateY(-2px)`, expanded shadow, and subtle child icon translation (`group-hover:translate-x-1`).
      - *Active/Click*: Tactile sink `translateY(2px)`, compressed bottom shelf (`border-bottom: 1px solid ...`), and tightened inner shadow.

  - **Navbar Design Patterns (Frosted Glass vs Full-Width Gradient Dissolve)**:
    - *Floating Glass Frosted Navbar*: Styled as an elevated pill/capsule (`h-14`–`h-16 rounded-[22px]`) using `bg-white/45 backdrop-blur-2xl border border-white/80` with a top specular highlight (`inset 0 1px 1px rgba(255,255,255,0.9)`) and soft ambient shadow (`0 10px 30px -10px rgba(...)`).
    - *Full-Width Gradient Dissolve Navbar*: For an ultra-seamless, modern bleed where the navbar dissolves naturally into the background mesh without a harsh border, use full-width `fixed top-0 left-0 right-0 h-20 bg-gradient-to-b from-[canvas-color]/90 via-[canvas-color]/60 to-transparent backdrop-blur-md`. This removes hard divider lines while keeping the top navigation legible.
    - *Desktop Navigation Interactions (Minimalist Text to Zoom-Scale Glass Pill on Hover)*:
      - In resting state, keep navigation links as clean, uncluttered plain text without permanent boxes, borders, or backgrounds.
      - On hover, do NOT shift links vertically (`translateY`); instead, trigger an **in-place 3D Zoom / Scale Pop** (`hover:scale-110`) that smoothly materializes a frosted glass capsule (`bg-white/80 border border-white shadow-md`), giving a lively tactile feel without disturbing baseline alignment.
      - On active section state, maintain a distinct colored pill (`bg-[#618264] text-white shadow-lg scale-105`).
    - *3D Tactile Oval Buttons (`rounded-full`)*:
      - Always use `rounded-full` for modern pill CTA buttons instead of standard boxy rounded corners.
      - Combine a vertical gradient, top specular highlight (`border-top: 1px solid rgba(255,255,255,0.45)`), tactile bottom shelf (`border-bottom: 3px solid #3b533e`), and drop shadows for an elevated, clickable physical feel.
    - *3-Stage Center Outward Bloom Welcome Transition*:
      - For cinematic landing page entries on refresh, transition from a deep dark tone (e.g. `#102013`) to the bright canvas by expanding a center light bloom portal from the center outward (`scale: 0.15 -> 2.8`, `blur: 90px`, `opacity: 0 -> 0.9 -> 0`).
      - Synchronize headline text reveal with the expanding center bloom (`delayChildren: 0.6s`, `staggerChildren: 0.08s`) so text crystallizes into focus precisely as the canvas is illuminated.
    - *Atmospheric Overhead Lighting & Spaceship-Inspired Ambient Aurora*:
      - Avoid sharp geometric light cones or hard polygonal masks (`clip-path: polygon(...)`) which look artificial and harsh.
      - Instead, create an **ambient overhead aurora sweep** using an elongated, ultra-soft diffused ellipse (`rounded-[100%]`, `w-[900px] h-[650px]`, `blur-[90px]`, `scale: 0.6 -> 2.6`) radiating from the top-center down across the hero canvas.
      - Combine with a deep-dark atmospheric start (`#0f1d12`) that holds calmly while text elements and 3D buttons reveal one by one, then dissolves smoothly (3.0s–3.6s duration) into the bright default canvas right as all elements complete their entrance.
    - *Zero White Drop Shadow / Halo Artifacts on Colored Gradients*:
      - Never add white or pale outer drop-shadows (`drop-shadow-[0_1px_2px_rgba(255,255,255,0.7)]`) behind dark text on colored/sage gradient canvases. It produces a dirty, unpolished halo effect. Use solid, high-contrast dark forest/charcoal typography (`#0e1c10`, `#1a381d`) for razor-sharp editorial legibility.
    - *Single-Sentence Line Layout & Optical Centering*:
      - Ensure headlines keep distinct sentence clauses on dedicated lines without mid-sentence wrapping (`whitespace-normal sm:whitespace-nowrap`).
      - Center the hero composition optically into the upper-middle area (~20–35px upward shift) rather than mathematically at 50vh, creating a balanced, art-directed weight.
    - *3D Tactile Glass Pill Buttons & In-Place Zoom Navigation*:
      - CTA Buttons: Use `rounded-full` oval capsules with a top specular highlight (`border-top: 1px solid rgba(255,255,255,0.45)`), tactile bottom shelf (`border-bottom: 3px solid ...`), and subtle spring depression on active click.
      - Navbar Links: Keep resting state clean and container-free; trigger an **in-place 3D scale/zoom pop** (`hover:scale-110`) that smoothly materializes a frosted glass pill without disturbing baseline alignment.
    - *Single-Page Pinned Frame Scrolling vs Multi-Page Native Flow*:
      - For sticky hero scrubbers that advance phases in-place without jarring page reloads or blank areas below, use a pinned viewport with debounced wheel/touch/keyboard event handling.
      - Prevent browser pull-to-refresh / overscroll reload artifacts by anchoring the container and handling scroll transitions gracefully.
    - *Pure Monochromatic Palette Saturation (Eliminating White-Wash Blends)*:
      - When the user specifies an organic or themed color palette (e.g. Sage Green `#527355`, `#618264`, `#79AC78`, `#b8dcb9`), ensure all ambient layers, animated blobs, and UI tints strictly use the themed color spectrum.
      - Remove harsh pure white `#ffffff` or white-tinted gradient washes in moving mesh blobs and secondary buttons, replacing them with pale-tinted variants (e.g. `#d0e7d2`, `#b8dcb9`) to maintain an authentic, saturated monochromatic atmosphere.
    - *Brand Logo Asset Optimization*:
      - When using raster PNG assets in the navbar, ensure background transparency is clean (crop bounding boxes and strip solid white backgrounds via PIL/canvas script to produce crisp transparent PNGs) so the logo blends natively onto dark or gradient mesh backdrops without halo artifacts.
    - *Mobile Navigation (Pure Borderless 3-Line Hamburger → Smooth X)*:
      - On mobile/small screens, render a pure borderless 3-line hamburger button (`border-none`, `bg-transparent`) with three distinct lines (`w-6 h-[3px] bg-[#162b19] rounded-full`).
      - On click/toggle, animate the top and bottom lines to `rotate: 45deg / -45deg` and `y: 8px / -8px` while collapsing the middle line (`scaleX: 0`, `opacity: 0`) with smooth Framer Motion spring physics to form a seamless **X** close icon, smoothly revealing a floating glass dropdown menu.

- **Interactive Stream & Chat Simulator Viewports in Pinned Layouts**:
  - When embedding an interactive conversation / chat component inside a pinned scroll scrubber:
    - **Event Isolation on Scroll Areas**: Add a specific selector check (e.g. `target.closest(".chat-scroll-area")`) inside the window `wheel` and `touchmove` listeners so user scrolling inside the message history does not accidentally flip the global slide step.
    - **React Pure Rendering Guard**: When generating IDs or timestamps inside event handlers / timeouts (e.g. `handleSendMessage`), use monotonic counter refs (`useRef`) or pass values via state rather than calling impure functions like `Date.now()` or `Math.random()` directly in render pathways, preventing React Compiler / ESLint `react-hooks/purity` build errors.
    - **Clean Editorial Coming Soon Viewports**: If the design specifies a container-free look for teaser/coming soon slides, avoid wrapping them in distinct white modal/cards; retain the exact same full-canvas display typography, 2-line structure, and word-by-word blur writer reveals as the intro hero for visual harmony.

## Pitfalls to Avoid
- **Pure white canvas with soft pastels**: Makes the entire page look empty/broken.
- **Identical animation timings**: Syncing blob durations makes movement look mechanical; vary each blob's timing by 3–6s and add slight initial delays.
- **Neglecting port collisions on deployment**: Always check active ports when serving Next.js apps under PM2.
