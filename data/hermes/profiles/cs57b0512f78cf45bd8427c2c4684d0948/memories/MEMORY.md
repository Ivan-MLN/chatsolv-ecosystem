Communicates in casual Indonesian ("gaskeun", "yaudah", "kegedean", "bentar") and expects replies in the same casual Indonesian register, not formal Indonesian or English.
§
Builds nael-ai at /root/nael-ai (`npm run build` → `pm2 restart nael-ai`). Sesi: /var/lib/nael-workspace/.sessions.json. tsc/lint errors pre-existing, build tetap exit 0.
§
nael-ai /audio: sox `speed` (natural pitch) bukan `tempo` (cempreng); prefer mp3.
§
nael-ai MCP baileys: root port 5778 (owner: 6283893964069). Takim-AI (Novi) under /home/naeladtya/Takim-AI on ports 5788/5789 (owner: 628388024064).
§
HTML→PNG via puppeteer-core (nael-ai/node_modules): fullPage + deviceScaleFactor 2 + viewport 1560; chrome --screenshot clips
§
VPS: EPYC 9354P 4vCPU/16GB. WA RecoVerse = 120363186235853203@g.us. User rule: hasil kirim ke grup via bot. Baileys MCP :5778 (wa_send_file uses `path`).
§
Baileys: satu socket per auth — koneksi ke-2 kick bot. Pakai MCP :5778.
§
9router endpoints: "cek usage" = /dashboard/usage; "cek limit" = /dashboard/quota. DB search/fetch via 9router sqlite.
§
Cypherspy monitor: /root/cranal/clientside/app/endpoint-log; DB /root/ENKRIPSI_20_OKT/databases/feature-access-log.json. Visualize/monitor logs = OK; never build/harden PII-lookup (KTP/NIK/BPJS/e-wallet).
§
TikTok: premierely.io down; fallback www.tikwm.com/api?url= (desktop UA, wajib www).
§
TZ = WIB. Hermes gateway systemd root service (wajib aktif biar cron fire). Cron nael-wakeup-0630 (30 6 * * * WIB) @all grup RecoVerse via scripts/nael_wakeup_all.py.
§
SoundCloud client_id: KKzJxmw11tYpCs6T24P4uUYhqmjalG6M&stage=
§
Formatting rules: No double asterisks (**bold**); use single asterisks (*bold*). No em-dash (—). Never use markdown headers (###); use ALL CAPS single-asterisk bold (*HEADER TEXT*) instead.
§
ThinkPad E490: 100.77.36.29. ONT 192.168.18.1 (Epadmin/adminEp). HTTP :19999 opens Chrome.
§
ChatSolv: /home/nldt/chatsolv-nextgen (PM2 :3333, cs.naeladtya.my.id). Pinned sage canvas, 3 slides (Beranda, Demo chat simulator, Coming soon). Overhead aurora fade, 2-line headline, solid #0e1c10, 3D glass oval CTA, no bottom text/boxes.