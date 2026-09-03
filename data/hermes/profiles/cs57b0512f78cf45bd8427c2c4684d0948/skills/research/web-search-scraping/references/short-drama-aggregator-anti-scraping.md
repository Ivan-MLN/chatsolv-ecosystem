## Short Drama Aggregators & Multi-Provider Anti-Scraping Architecture (e.g. Narto-Drama, DraCinku)

Short drama streaming sites fall into two main architectural patterns:

### Pattern A: Edge-Fleet Dynamic Aggregators (e.g. `narto-drama.com`)

1. **Cloudflare WAF & Anti-Bot Protection**
   - The entire front-end sits behind Cloudflare. Un-stealthed HTTP scrapers trigger JS challenges, Cloudflare Turnstile, or `403/530` blocks instantly.

2. **Client-Side Edge Fleet & Dynamic JS Rendering**
   - No `.m3u8` or `.mp4` stream URLs exist in static HTML source code.
   - Streams are fetched dynamically via JavaScript running against distributed edge fleets (e.g., `edge.narto-drama.com`, `dramabox-edge.*.workers.dev`, `stream-e1.narto-drama.com`).

3. **Signed Short-Lived Tokens & Required Headers**
   - Endpoint video URLs are signed dynamically (e.g., `dramabox.narto-drama.com/<token>`).
   - Requires session context (`laravel-session`, `XSRF-TOKEN`), custom headers (`X-Requested-With: XMLHttpRequest`), and strict `Referer`/`Origin` checks.

4. **Local Gating & Anti-Bottleneck Logic**
   - JavaScript offloads API requests to local edge paths (`/e/rs/...`) and checks `localStorage` cache gates before triggering fetches to avoid overloading FPM backends.

5. **Adsterra Popunder & NavGuard Hijacking**
   - Pages inject script wrappers (`NavGuard`) overriding `window.open` and `location.assign`, combined with Adsterra popunders, breaking naive browser automation loops.

**Strategy for Pattern A:** Use Playwright / Puppeteer with stealth plugins (or Camoufox via Scrapling) and intercept dynamic XHR/Fetch network calls (`.m3u8`, `.mp4`, or edge stream tokens) as the player initializes.

---

### Pattern B: WordPress + Third-Party Embed Sites (e.g. `dracinku.com`)

1. **Static HTML Scraping works for Metadata & Links (200 OK)**
   - Sites built on WordPress (e.g. Muvipro theme) do not heavily obfuscate their page HTML or block static HTTP GET requests.
   - Drama titles, episode lists, thumbnails, and embed `<iframe>` URLs can be extracted directly via basic `requests` + `BeautifulSoup` or regex.

2. **Third-Party Iframe Storage (MEGA.nz, UPN, Doodstream)**
   - Videoplayer source links point to external hosters rather than internal edge APIs.
   - E.g. `<iframe src="https://mega.nz/embed/ua4AEIwI#D5a5-OBSs1ZB1hGUsdiBtsOfexDQvC4twDWmOR8noi0">` or UPN (`https://wigita.upn.one/#...`).
  - Python quick scraper recipe: extract `/YYYY/MM/DD/title/` post URLs from home page, hit detail page, extract `entry-title`, `wp-post-image`, and iframe srcs matching `mega.nz/embed/` or `?player=\d+`.

3. **Downloading Raw Video Streams**
   - While metadata extraction is trivial, downloading raw `.mp4` video files requires resolving the third-party embed (e.g. MEGA API decryption using the hash key `#D5a5-...` in the URL or bypassing Cloudflare on hosters like UPN).

**Strategy for Pattern B:** Use static HTTP GET to extract metadata and iframe links fast; parse hash keys / embed parameters for third-party hosters if downloading media is required.
