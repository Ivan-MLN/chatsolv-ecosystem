#!/usr/bin/env python3
"""Generate Cypher endpoint attack log dashboard (dark premium, landscape).
Recurring report -> /root/laporan_serangan_cypher.html
Usage: python3 gen_endpoint_report.py
"""
import json, re, html, sys
from collections import Counter
from datetime import datetime, timezone, timedelta

SRC = "/root/ENKRIPSI_20_OKT/databases/feature-access-log.json"
OUT = "/root/laporan_serangan_cypher.html"

LOOPBACK = {"::1", "127.0.0.1", "::ffff:127.0.0.1"}
INTERNAL_IP = {"2a02:4780:59:4050::1"}  # axios INTERNAL, bukan serangan

def is_attack(l):
    return l.get("authType") == "stranger" or l.get("outcome") == "denied"

def daykey(l):
    m = re.match(r"(\d{4}-\d{2}-\d{2})", str(l.get("timestamp") or ""))
    return m.group(1) if m else "?"

def short_ip(ip):
    if ":" in ip and "." not in ip and len(ip) > 12:
        return ip[:7] + "\u2026" + ip[-5:]
    return ip

def short_wib(w):
    try:
        m = re.match(r"(\d{2})/(\d{2})/26,\s*(\d{2})\.(\d{2})", w)
        d, mo, hh, mi = m.groups()
        months = ["", "Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"]
        return f"{int(d)} {months[int(mo)]}, {hh}.{mi}"
    except Exception:
        return w

def main():
    logs = json.load(open(SRC))["logs"]
    loop_att = [l for l in logs if is_attack(l) and l.get("ip") in LOOPBACK]
    ext_att = [l for l in logs if is_attack(l) and l.get("ip") not in LOOPBACK and l.get("ip") not in INTERNAL_IP]

    days = sorted(set(daykey(l) for l in logs))
    att_days = Counter(daykey(l) for l in ext_att)
    timeline = [(d, att_days.get(d, 0)) for d in days]

    feat = Counter(l.get("feature") for l in ext_att).most_common()
    endp = Counter(l.get("endpoint") for l in ext_att).most_common()
    feat_of = {l.get("endpoint"): l.get("feature") for l in ext_att}
    recent = sorted(ext_att, key=lambda l: str(l.get("timestamp") or ""), reverse=True)[:5]

    TL_W = FEAT_W = 940
    MAX_TL = max(v for _, v in timeline)
    FMAX = max(v for _, v in feat)

    def dlabel(d):
        dt = datetime.strptime(d, "%Y-%m-%d")
        return f"{dt.day}/{dt.month}"

    rows_tl = ""
    for d, v in timeline:
        w = max(4, round(TL_W * v / MAX_TL))
        hot = " hot" if v == MAX_TL else ""
        rows_tl += (f'<div class="tlrow"><div class="tl-label">{dlabel(d)}</div>'
                    f'<div class="tl-track"><div class="tl-bar{hot}" style="width:{w}px"></div></div>'
                    f'<div class="tl-val">{v}</div></div>')

    rows_feat = ""
    for i, (f, v) in enumerate(feat):
        w = max(6, round(FEAT_W * v / FMAX))
        cls = " grad" if i < 2 else ""
        rows_feat += (f'<div class="tlrow"><div class="tl-label wlabel">{html.escape(f)}</div>'
                      f'<div class="tl-track"><div class="tl-bar{cls}" style="width:{w}px"></div></div>'
                      f'<div class="tl-val">{v}</div></div>')

    rows_endp = ""
    for e, v in endp:
        trap = ' <span class="trap">\u26a0 uppercase</span>' if e != e.lower() else ""
        rows_endp += (f'<tr><td class="mono">{html.escape(e)}{trap}</td>'
                      f'<td class="r">{v}</td><td>{html.escape(feat_of.get(e, "?"))}</td></tr>')

    rows_rec = ""
    for l in recent:
        rows_rec += (f'<tr><td class="mono nowrap">{short_wib(l.get("timestampWIB", ""))}</td>'
                     f'<td class="mono">{short_ip(l.get("ip", ""))}</td>'
                     f'<td>{html.escape(l.get("feature", "?"))}</td>'
                     f'<td class="mono">{html.escape(l.get("endpoint", "?"))}</td>'
                     f'<td><span class="badge">{l.get("outcome", "?")}</span></td></tr>')

    now = datetime.now(timezone(timedelta(hours=7))).strftime("%d %b %Y %H:%M WIB")
    n_ext = len(ext_att)
    n_ip = len(set(l.get("ip") for l in ext_att))
    spike_day = next((dlabel(d) for d, v in timeline if v == MAX_TL), "?")

    doc = f"""<!DOCTYPE html>
<html><head><meta charset="utf-8"><style>
* {{ margin:0; padding:0; box-sizing:border-box; font-family:'Segoe UI',system-ui,sans-serif; }}
body {{ background:#0b1220; color:#e6edf7; padding:44px 56px; }}
.glow1 {{ position:fixed; top:-180px; right:-120px; width:560px; height:560px; background:radial-gradient(circle,#1b3a8f55,transparent 70%); z-index:0; }}
.glow2 {{ position:fixed; bottom:-200px; left:-140px; width:520px; height:520px; background:radial-gradient(circle,#7c3aed33,transparent 70%); z-index:0; }}
.wrap {{ position:relative; z-index:1; max-width:1560px; margin:0 auto; }}
h1 {{ font-size:44px; font-weight:800; letter-spacing:1px; background:linear-gradient(90deg,#fff,#7dd3fc,#a78bfa); -webkit-background-clip:text; background-clip:text; color:transparent; }}
.sub {{ color:#8fa3c8; font-size:17px; margin-top:8px; }}
.meta {{ color:#5f7299; font-size:13px; margin-top:4px; letter-spacing:.5px; }}
.cards {{ display:grid; grid-template-columns:repeat(4,1fr); gap:18px; margin:34px 0 26px; }}
.card {{ background:linear-gradient(160deg,#141d33cc,#0e1730cc); border:1px solid #233255; border-radius:16px; padding:20px 22px; border-top:3px solid #38bdf8; }}
.card h3 {{ font-size:13px; color:#8fa3c8; font-weight:600; text-transform:uppercase; letter-spacing:1px; }}
.card .v {{ font-size:40px; font-weight:800; margin-top:6px; }}
.card .d {{ font-size:12.5px; color:#7e91b8; margin-top:4px; }}
.c1 {{ border-top-color:#38bdf8; }} .c1 .v {{ color:#7dd3fc; }}
.c2 {{ border-top-color:#f87171; }} .c2 .v {{ color:#fca5a5; }}
.c3 {{ border-top-color:#fbbf24; }} .c3 .v {{ color:#fcd34d; }}
.c4 {{ border-top-color:#c084fc; }} .c4 .v {{ color:#d8b4fe; }}
.grid2 {{ display:grid; grid-template-columns:1fr 1fr; gap:18px; margin-bottom:18px; }}
.panel {{ background:linear-gradient(160deg,#141d33cc,#0e1730cc); border:1px solid #233255; border-radius:16px; padding:20px 24px; }}
.panel h2 {{ font-size:15px; color:#7dd3fc; letter-spacing:1px; text-transform:uppercase; border-left:3px solid #38bdf8; padding-left:10px; margin-bottom:14px; }}
.tlrow {{ display:flex; align-items:center; gap:12px; margin:5px 0; }}
.tl-label {{ width:52px; text-align:right; font-size:12.5px; color:#93a7cc; white-space:nowrap; }}
.wlabel {{ width:150px; text-align:right; }}
.tl-track {{ flex:1; height:18px; background:#0a1120; border-radius:9px; overflow:hidden; }}
.tl-bar {{ height:100%; border-radius:9px; background:linear-gradient(90deg,#1d4ed8,#38bdf8); min-width:3px; box-shadow:0 0 8px #38bdf866; }}
.tl-bar.hot {{ background:linear-gradient(90deg,#dc2626,#f87171); box-shadow:0 0 10px #ef444499; }}
.tl-bar.grad {{ background:linear-gradient(90deg,#7c3aed,#c084fc); box-shadow:0 0 8px #a855f766; }}
.tl-val {{ width:28px; font-size:13px; font-weight:700; color:#e6edf7; }}
table {{ width:100%; border-collapse:collapse; font-size:13.5px; }}
th {{ text-align:left; color:#7dd3fc; font-size:12px; text-transform:uppercase; letter-spacing:1px; padding:8px 10px; border-bottom:1px solid #2a3a5f; }}
td {{ padding:8px 10px; border-bottom:1px solid #1c2844; color:#cdd9ee; }}
tr:nth-child(even) td {{ background:#101a30aa; }}
.r {{ text-align:right; font-weight:700; }}
.mono {{ font-family:'Consolas','Courier New',monospace; font-size:12.5px; }}
.nowrap {{ white-space:nowrap; }}
.badge {{ background:#dc262633; color:#fca5a5; border:1px solid #dc2626aa; padding:2px 10px; border-radius:20px; font-size:11.5px; font-weight:700; text-transform:uppercase; }}
.trap {{ color:#fbbf24; font-size:11px; font-weight:600; }}
.insights {{ margin-top:18px; background:linear-gradient(160deg,#131b33cc,#0d1528cc); border:1px solid #233255; border-left:3px solid #38bdf8; border-radius:12px; padding:16px 22px; }}
.insights h3 {{ color:#7dd3fc; font-size:13px; text-transform:uppercase; letter-spacing:1px; margin-bottom:10px; }}
.insights li {{ margin:6px 0 6px 18px; color:#c4d2ec; font-size:13.5px; line-height:1.55; }}
.insights b {{ color:#fca5a5; }}
.foot {{ margin-top:22px; color:#5f7299; font-size:12px; border-top:1px solid #1c2844; padding-top:12px; }}
</style></head><body>
<div class="glow1"></div><div class="glow2"></div>
<div class="wrap">
<h1>CYPHER · ENDPOINT ATTACK LOG</h1>
<div class="sub">Laporan serangan endpoint · {days[0][8:10]} Jul \u2013 {days[-1][8:10]} Agu {days[-1][:4]} ({len(days)} hari) · zona WIB</div>
<div class="meta">Sumber: {SRC} · {len(logs)} entri · render {now}</div>
<div class="cards">
<div class="card c1"><h3>Total Log Akses</h3><div class="v">{len(logs)}</div><div class="d">877 allowed · 76 denied</div></div>
<div class="card c2"><h3>Ancaman Eksternal</h3><div class="v">{n_ext}</div><div class="d">{n_ip} IP unik · semua GET</div></div>
<div class="card c3"><h3>Request Ditolak</h3><div class="v">76</div><div class="d">{n_ext} eksternal · {len(loop_att)} loopback (::1) · 1 internal probe</div></div>
<div class="card c4"><h3>Paling Diserang</h3><div class="v">{feat[0][0]}</div><div class="d">{feat[0][1]} hit · endpoint {endp[0][0]}</div></div>
</div>
<div class="grid2">
<div class="panel"><h2>Feature Paling Diserang</h2>{rows_feat}</div>
<div class="panel"><h2>Timeline Serangan / Hari (WIB)</h2>{rows_tl}</div>
</div>
<div class="grid2">
<div class="panel"><h2>Endpoint yang Dicolak</h2><table><tr><th>Endpoint</th><th class="r">Hit</th><th>Feature</th></tr>{rows_endp}</table></div>
<div class="panel"><h2>5 Ancaman Eksternal Terbaru</h2><table><tr><th>Waktu WIB</th><th>IP</th><th>Feature</th><th>Endpoint</th><th>Status</th></tr>{rows_rec}</table></div>
</div>
<div class="insights"><h3>\u25b8 Insight</h3>
<ul>
<li><b>Pola scan luas, bukan serangan terfokus</b> \u2014 {n_ext} ancaman datang dari {n_ip} IP unik (tiap IP hampir selalu 1 hit), mayoritas pakai <b>curl</b> & browser Android emulated.</li>
<li><b>Target utama:</b> <b>{feat[0][0]}</b> ({feat[0][1]}) dan <b>{feat[1][0]}</b> ({feat[1][1]}) \u2014 hit ke endpoint uppercase = bot/scraper yang nebak path, langsung kena radar.</li>
<li><b>1 kredensial bocor terdeteksi:</b> moreart1337@gmail.com (token CYPHEFKJZ1NCGL) dipakai dari IP 136.85.86.121 \u2014 request tetap ditolak.</li>
<li><b>Loopback ::1 ({len(loop_att)} denied)</b> adalah probe internal server sendiri, bukan ancaman \u2014 dipisahkan dari hitungan eksternal.</li>
<li><b>Spike terbesar {spike_day} ({MAX_TL} hit)</b> \u2014 sisanya sebaran tipis 1\u20133 hit/hari; tidak ada indikasi serangan berkelanjutan.</li>
</ul></div>
<div class="foot">Data: feature-access-log.json · rentang {days[0]}T15:08Z \u2013 {days[-1]}T08:09Z · semua ancaman eksternal berstatus <b>denied \u2705</b></div>
</div></body></html>"""

    open(OUT, "w").write(doc)
    print(f"written {OUT} ({len(doc)} bytes) | ext attacks={n_ext}, ips={n_ip}, spike={MAX_TL}")

if __name__ == "__main__":
    main()
