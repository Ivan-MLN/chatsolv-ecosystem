# Worked snippets: Bing RSS + live-site scraping (blocked VPS)

Walias of working code from searching for "universities swasta IT with open admission 2026".

## 1. Search via Bing RSS (works when the browser hits CAPTCHA/Cloudflare)

Loop several queries; each returns a parseable feed in ~0.5s:

```python
import subprocess, re, html, urllib.parse
UA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/124.0 Safari/537.36"

def search(q, limit=10):
    url = "https://www.bing.com/search?format=rss&q=" + urllib.parse.quote(q)
    r = subprocess.run(["curl","-s","--max-time","30",url,"-H",f"user-agent: {UA}"],
                       capture_output=True, text=True)
    items = re.findall(r'<item>(.*?)</item>', r.stdout, re.S)
    out=[]
    for it in items[:limit]:
        t  = re.search(r'<title>(.*?)</title>', it, re.S)
        l  = re.search(r'<link>(.*?)</link>', it, re.S)
        d  = re.search(r'<description>(.*?)</description>', it, re.S)
        out.append((html.unescape(t.group(1)) if t else '?',
                    (l.group(1) if l else '?'),
                    re.sub(r'<[^>]+>',' ', html.unescape(d.group(1)) if d else '')[:200]))
    return out
```

### Pitfall that actually happened
Three different queries (PMB BINUS / Telkom / UMN / BSI / UNSIKA) all returned the **same 4 items** (pmb.uinsgd, enrollment.ui, pmb.upi, pmb.unjani). That's a stale/cached serve, NOT per-query truth. Hit that in the "search for university admission pages" flow — the fix was to stop trusting that search result and open the candidate sites directly.

## Scraping an individual page with `scrapling` HTTP Fetcher

```bash
pip install scrapling   # pure-python, only python3 needed; no browser binary
```

```python
from scrapling.fetchers import Fetcher
import re
p    = Fetcher.get("https://pmb.uty.ac.id/", timeout=25)        # 0.2.99
txt  = re.sub(r"\s+", " ", p.css_first("body").get().get())
```

### API notes
- **0.2.x `Fetcher.get()` does NOT accept an `impersonate=`/`stealthy_headers=` kwarg** — calling with them throws `TypeError: unexpected keyword`. Just call `Fetcher.get(url, timeout=25)`.
- `Adaptor` (what `css_first` returns) has `.get()`, and calling `.get()` again unwraps to the inner result (`.get().get()`); `Adaptor` itself has no `.text_content()`.
- Sites rendered entirely by JS come back near-empty via HTTP; `curl` gets similar — for those either use the browser tools or accept they need JS.

## Extracting structured deadlines from university admission pages

UTY PMB embedded the admission windows as inline JSON. Instead of reading marketing prose, harvest the keys:

```python
for m in re.finditer(r'"nama_jalur":"(.*?)".*?"tgl_mulai":"(.*?)".*?"tgl_selesai":"(.*?)"', txt):
    print(m.group(1), "|", m.group(2), "->", m.group(3))
```
Used this to conclude "no-tok Special / Fast Track Rapor / Reguler online+di kampus" admissions in 2026/2027 all open until **2026-08-31 / 2026-09-10**. AMIKOM's static marketing page gave `PMB Gelombang 3 dibuka hingga 27 Agustus 2026`.

## Interpretation for a "open admissions" style answer
- **Verified reachable + phrase open date** → "still open" fact.
- **Connection refused / Errno 111** (happened with BINUS, Mercu Buana, Trisakti, Telkom, etc from this VPS) → not verifiable from server; matched them for the user.

## Result summary shape of a "find universities" answer
Prefer a short verified-first list (university, PMB URL, date found), then a "could not verify from here — check manually" bucket. Indonesian PMB terms: Gelombang (wave 1/2/3), Jalur (track)), TA 2026/2027 (academic year). Useful keywords: `PMB`, `pendaftaran mahasiswa baru`, `dibuka hingga`, `ditutup`.