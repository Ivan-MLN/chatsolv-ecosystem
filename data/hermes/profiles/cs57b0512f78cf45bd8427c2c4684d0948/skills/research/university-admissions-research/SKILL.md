---
name: university-admissions-research
description: Research uni PMB deadlines, biaya, akreditasi, rating.
version: 1.0.0
author: cathlyne
license: MIT
metadata:
  hermes:
    tags: [Universities, PMB, Admissions, Scraping, Indonesia, Rankings]
    related_skills: [scrapling]
---

# University Admissions Research (PMB Indonesia)

Use when the user asks about universities: "pendaftaran masih dibuka", PMB, biaya kuliah / UKT, biaya pendaftaran, akreditasi, peringkat kampus, rating/ulasan kampus — especially Indonesian private universities (swasta) for IT/CS.

## Workflow

1. **Official PMB site first, always.** URL patterns: `pmb.<campus>.ac.id`, `admission.<campus>.ac.id`, `pendaftaran.<campus>.ac.id`. Deadlines in banners/body text on the PMB homepage are the most trustworthy source (e.g. "PMB Gelombang 3 ... 27 Agustus 2026"). Verify deadline + open/closed status from the official site, not from search snippets.
2. **Search engines from this VPS are bot-checked** (Google, DDG, Bing web, Ecosia, Brave, lite-DDG → CAPTCHA/Cloudflare). Bing RSS passes: `curl -s "https://www.bing.com/search?format=rss&q=..."` with a desktop UA. Use it ONLY for link/URL discovery — the RSS results are duplicated across queries and unreliable as a data source.
3. **Scrape with scrapling `Fetcher`** (HTTP impersonation). Distinguish failure types:
   - `Connection refused` → geo/ASN block of the whole domain (BINUS, Telkom, BSI, Mercu Buana, Trisakti all refused from VPS). Try the browser tool once; if it also fails, tell the user to check manually.
   - `403` → Cloudflare front (Gunadarma). Try browser or `StealthyFetcher`.
   - `200` → good; extract text via `p.css_first("body").get().get()` then regex-keyword scan.
4. **Fee tables are often JS-rendered OR published as an image.** If the page's content is a picture (AMIKOM's `biaya-kuliah` page), do NOT browser-screenshot-pixel-hunt: curl the raw HTML, regex for `https?://[^\s"']+\.(png|jpg|jpeg)` in `wp-content/uploads/`, download the largest image, and transcribe it with `vision_analyze`. This yields exact Rp figures and table structure.
5. **Rankings:** eduRank profiles are JS-rendered — use the browser, and use the site's **internal search box** (URL slug guesses 404; `edurank.org/computer-science/id/` style paths 404 too). Their data: overall rank (world/Asia/Indonesia), per-subject rank, enrollment, acceptance rate (estimated ~40% when not published — always flag it as an estimate).
6. **Campus marketing claims are self-promotion.** "Ranking #1 Dunia" banners from the university's own brochure ≠ EduRank/independent rank. Report both, labeled clearly.
7. **Ratings & reviews:** Google Maps search works from the VPS browser even though Google Search is blocked — `https://www.google.com/maps/search/<univ name>` → star rating + review count + address/phone/website in one shot.

## Pitfalls

- **Aggregator data goes stale**: RuangGuru listed AMIKOM as "akreditasi B, Rp7.17–11.6jt/sem" while the official site (SK BAN-PT No. 3107/.../XII/2025) says **Unggul** with Rp10.75jt/sem S1 Informatika. Trust official SK BAN-PT numbers over aggregators; cite the SK number when you have it.
- **Fees change yearly** — always re-verify from the current-year official table/image; don't carry last year's numbers forward.
- **KIP Kuliah typically closes ~31 Juli** for the academic year; by August only regular/fast-track/mandiri jalur remain open.
- **Re-registration after failing selection is often free** (AMIKOM: "daftar ulang gratis") — worth mentioning to the user.
- For campus admission questions the user usually wants a **deadline table + biaya awal (sem 1 total) + rating** — lead with those, not prose.
- User rule: hasil riset/gambar/file langsung kirim ke grup WA RecoVerse Team via bot (`wa_send_file` / `wa_send_message`) — jangan nunggu diminta.

## References

- `references/amikom-uty-2026.md` — verified detail + source URLs for AMIKOM & UTY (deep-dive data, fee tables, rankings, ratings).