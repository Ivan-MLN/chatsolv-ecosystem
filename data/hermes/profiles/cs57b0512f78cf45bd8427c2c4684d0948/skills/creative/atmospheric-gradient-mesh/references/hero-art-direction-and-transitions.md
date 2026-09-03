# Hero Art Direction, Atmospheric Lighting, and Interactive Transitions

Comprehensive architectural and design guidelines derived from production iterations on modern SaaS hero experiences.

## 1. Cinematic Opening Transitions (Spaceship-Inspired Ambient Aurora)

- **Avoid Hard Geometry**: Never use sharp CSS polygon clip-paths (`clip-path: polygon(...)`) or angular light cones to simulate spotlight beams. They look artificial, jagged, and dated.
- **Ultra-Soft Ambient Aurora Sweep**: Use a high-radius soft ellipse (`rounded-[100%]`, `w-[900px] h-[650px]`, `blur-[90px]`, `scale: 0.6 -> 2.6`) positioned slightly above the top edge (`-top-32 left-1/2 -translate-x-1/2`).
- **Dark-to-Bright Entrance Pacing**:
  - Hold an atmospheric dark overlay (`#0b160e` or `#0f1d12` deep theme tone) for the first 0.5s–1.2s while hero headline words and CTA buttons sequentially cascade into view via spring physics.
  - Set word-level blur writer reveals to begin smoothly with a short, responsive delay (`delayChildren: 0.12s` to `0.35s`, `staggerChildren: 0.08s`).
  - Lift the dark atmosphere overlay (2.0s–2.4s total duration) seamlessly right after all headline words, supporting sentences, and CTA buttons have settled in place.

## 2. 3D Tactile Glass Pill Buttons (Glassmorphism 2.0)

- **Shape**: Pure oval capsule (`rounded-full`, height 48–52px, horizontal padding px-7 to px-8).
- **Surface Lighting & Bevels**:
  - *Primary (Themed/Emerald)*: Linear gradient top-to-bottom, specular top rim highlight (`border-top: 1px solid rgba(255,255,255,0.45)`), tactile bottom shelf (`border-bottom: 3px solid #3b533e`), and soft ambient drop shadow (`0 8px 20px -6px rgba(...)`).
  - *Secondary (Frosted Glass)*: Translucent glass (`backdrop-filter: blur(20px)`), glossy top highlight (`border-top: 1px solid rgba(255,255,255,0.95)`), subtle mint bottom rim (`border-bottom: 3px solid #b6d3b8`), and inset top glow.
- **Physical Micro-Interactions**:
  - *Hover*: Smooth upward float (`translateY(-2px)`), expanded shadow, and subtle icon shift (`group-hover:translate-x-1`).
  - *Active (Click)*: Physical compression (`translateY(2px)`), reduced bottom shelf thickness, and tightened shadow.

## 3. Minimalist In-Place Zoom Navigation Bar

- **Resting State**: Plain, clean high-contrast text links (`Masalah`, `Cara Kerja`, `Kendali`, `Hasil`) directly over the background with zero permanent containers, boxes, or borders.
- **Hover State**: Do NOT shift links vertically (`translateY`). Trigger an in-place **3D scale/zoom pop** (`hover:scale-110`) that smoothly materializes a frosted glass pill background (`bg-white/80 border border-white shadow-md`), maintaining strict baseline alignment.
- **Active Section State**: Maintain a distinct colored pill (`bg-[#618264] text-white shadow-lg scale-105`).
- **Seamless Navbar Bleed**: Use `bg-gradient-to-b from-[canvas-color]/90 via-[canvas-color]/60 to-transparent backdrop-blur-md` with no bottom border so the header dissolves naturally into the background mesh.

## 4. Pure Borderless 3-Line Hamburger to 'X' Morph & Right Drawer

- **Resting**: Three horizontal rounded bars (`w-6 h-[3px] rounded-full`) without button background or border (`border-none bg-transparent`).
- **Open State (Morph to X)**:
  - Top line: `rotate: 45deg, y: 8px`
  - Middle line: `scaleX: 0, opacity: 0`
  - Bottom line: `rotate: -45deg, y: -8px`
  - Smooth spring transition (`duration: 0.3s, ease: [0.22, 1, 0.36, 1]`).
- **Right Sidebar Drawer**: On mobile screens, slide in a right-hand frosted glass drawer (`bg-[canvas]/95 backdrop-blur-3xl border-l`) paired with a dimmed backdrop blur and staggered item reveals for navigation links and bottom CTA buttons.

## 5. Optical Centering & Typography Discipline

- **Optical Upper-Middle Bias**: Avoid placing hero text at mathematical 50vh (which feels heavy and droopy at the bottom). Elevate the central composition 20–35px upward (`-translate-y-6`) into the active breathing room between navbar and bottom footer.
- **Single Sentence per Line**: Size display headlines responsively (`clamp(32px, 4.5vw, 68px)`) with `whitespace-nowrap` on desktop so distinct clauses sit on dedicated lines without awkward mid-sentence breaks.
- **Zero White Drop Shadows / Halos**: Never add white or pale outer drop-shadows behind dark text on colored gradients. Rely strictly on high-contrast solid typography (`#0e1c10`, `#1a381d`) for crisp editorial legibility.
- **Supporting Sentence Restraint**: Place exactly one subtle supporting sentence (max-width 600px–680px, font 16px–18px, line-height 1.5) with controlled gaps (24px above, 28px below to CTAs).
