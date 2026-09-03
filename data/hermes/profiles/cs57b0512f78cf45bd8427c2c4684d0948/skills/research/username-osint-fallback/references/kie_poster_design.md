# Media KIE Poster & Educational Graphics Guidelines

When generating visual KIE (Komunikasi, Informasi, & Edukasi) posters or infographics (e.g. Stunting, Public Health, Government Campaigns):

## Official Logo Integration
1. **Dual Logo Header Bar**: When multiple logos are provided or requested (e.g. Local Regency / Pemkab logo + National Ministry / BKKBN / Kemenkes logo):
   - Place a dedicated clean white header bar (`.top-logo-bar`) at the top of the canvas.
   - Align logos side-by-side (left: local authority / Pemkab, right: national ministry / BKKBN).
   - Include official typography next to the local logo (e.g. `PEMERINTAH KABUPATEN TAKALAR`).

## Visual Elements & Image Showcases
- Incorporate realistic stock photos or relevant photography (e.g. Mother & Child, Protein-Rich Food, Posyandu Checkups) in a hero grid layout (`.hero-grid`).
- Use high-contrast, modern dark slate / blue gradient themes (`#0f172a`, `#1e293b`) with clear typography and distinct cards for key pillars (e.g. 1.000 HPK).

## Rendering Pipeline
- Use HTML + CSS rendered via headless Chrome (`/usr/bin/google-chrome --headless --screenshot=... --window-size=1200,1850`) to ensure crisp text alignment, crisp vector badge rendering, and exact pixel scaling without truncation.
