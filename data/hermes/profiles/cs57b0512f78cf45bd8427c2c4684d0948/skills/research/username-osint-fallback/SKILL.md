---
name: username-osint-fallback
description: Use when sherlock fails; fallback search for usernames.
version: 1.0.0
---

# Username OSINT Fallback

Use this workflow when `sherlock` or username search tools fail (e.g. Python syntax/version incompatibility like `TypeError: unsupported operand type(s) for |: 'type' and 'NoneType'`, missing binary, rate limits, or site blocking).

## Known Sherlock & Interaction Pitfalls
- **Python < 3.10 Incompatibility**: `sherlock_project` uses PEP 604 type annotations (`str | None`) which crashes Python 3.9 runtime with `TypeError: unsupported operand type(s) for |: 'type' and 'NoneType'`. When this happens, skip `sherlock` CLI and immediately use Serper API fallback.
- **Phone Number OSINT**: Search exact formats (`"628388024064"`, `"08388024064"`) to discover public profiles (like GitHub readme direct wa.me links) and card carrier prefix identification.
- **User Preference - Game Presentation**: When playing interactive games in chat/group (e.g. Ular Tangga, Susun Kata), if the user specifies "stop pke gambar" or asks to avoid images, immediately stop rendering/sending PIL images and switch to clean, text-only updates/messages.
- **User Preference - Media KIE / Edukasi**: When creating KIE (Komunikasi, Informasi, & Edukasi) material (e.g., Stunting, Public Health), structure content into clear, well-formatted bullet points (5 Pilar, HPK, Slogan) with authoritative references (Kemenkes/BKKBN). When images/posters are requested, support dual official logos (e.g., Pemkab local logo + BKKBN/Kemenkes logo) seamlessly in the header via HTML/Puppeteer rendering.

## Fallback Workflow

When requested to search or dox/investigate a username across platforms:

1. **Direct Search Engine Queries**
   Run targeted queries using search API (Serper / Google / Bing):
   - `"username"`
   - `"user_name"`
   - `site:github.com "username"`
   - `site:instagram.com "username"`
   - `site:tiktok.com "username"`
   - `site:linkedin.com/in "username"`
   - `site:x.com "username"`
   - `"username" comment OR komentar` (to discover public comment threads on IG Reels/TikTok/YouTube)

2. **Parsing & Verification**
   - Filter out irrelevant keyword matches and generic search aggregator noise.
   - Verify specific user profiles, comments, interactions, and public bios.

3. **Presenting Findings**
   - Group discovered profiles by platform.
   - Provide direct links to verified public profiles or activity posts.
