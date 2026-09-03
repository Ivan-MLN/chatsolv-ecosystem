// Re-render VPS Monitor dashboard locally (wa_sys_monitor only sends to WA,
// it does NOT save the PNG). Copy this, replace values with the live response
// JSON from mcp__baileys__wa_sys_monitor, then:
//   node vps_monitor_chart.mjs
//   node <skill>/scripts/render_png.mjs /root/vps_monitor.html /root/vps_monitor.png
import fs from "fs";

// Extract chartHtml from the bot's compiled MCP so the visual matches the
// bot's output exactly (same style, same section layout).
const src = fs.readFileSync("/root/nael-ai/dist/ai/mcp-baileys.js", "utf8");
const m = src.match(/const chartHtml = \(title, subtitle, cards, bars, sections, footer\) => `<!DOCTYPE html>[\s\S]*?<\/html>`;/);
if (!m) { console.error("chartHtml not found"); process.exit(1); }
const mod = { exports: {} };
new Function("module", "exports", m[0].replace(/^const /, "var ") + "\nmodule.exports = chartHtml;")(mod, mod.exports);
const chartHtml = mod.exports;

const gb = (n) => `${(n / 1024 / 1024 / 1024).toFixed(1)} GB`;
const now = new Date().toISOString().replace("T", " ").slice(0, 19);

// Replace with live values from wa_sys_monitor JSON response
const hostname = "srv1334411", osName = "Debian GNU/Linux 11 (bullseye)", pubIp = "76.13.193.152";
const cpuPct = 15.3, perCore = [{name:"cpu0",pct:17},{name:"cpu1",pct:15},{name:"cpu2",pct:19},{name:"cpu3",pct:17}];
const ramUsed = "10.1 GB", ramTotal = "15.6 GB", ramPct = 65, swap = "0 GB";
const diskPct = 81, upDays = 32, upH = 5;
const load = ["0.85", "0.78", "0.49"], procs = 264;
const rx = "278.1 GB", tx = "305.6 GB";
const topProcs = [
  ["bun (1442992)", "9.3% · 1536 MB"],
  ["vscode-server (1783171)", "4.5% · 732 MB"],
  ["vscode-server (1777877)", "2.7% · 436 MB"],
  ["vscode-server (1776601)", "2.6% · 421 MB"],
  ["claude (1783516)", "2.2% · 369 MB"],
];

const html = chartHtml(
  "VPS Monitor",
  `${hostname} · ${osName} · ${pubIp}`,
  [
    { label: "CPU", string: `${Math.round(cpuPct)}%`, color: "#38bdf8" },
    { label: "RAM", string: ramUsed, color: "#34d399" },
    { label: "Disk /", string: `${diskPct}%`, color: "#fbbf24" },
    { label: "Uptime", string: `${upDays}d ${upH}h`, color: "#a78bfa" },
  ],
  [
    { label: "CPU", string: `${cpuPct.toFixed(1)}%`, pct: cpuPct, color: "linear-gradient(90deg,#38bdf8,#818cf8)" },
    { label: "RAM", string: `${ramUsed} / ${ramTotal}`, pct: ramPct, color: "linear-gradient(90deg,#34d399,#38bdf8)" },
    { label: "Swap", string: swap, pct: 0, color: "linear-gradient(90deg,#fbbf24,#f472b6)" },
    ...perCore.map((c) => ({ label: c.name.toUpperCase(), string: `${c.pct}%`, pct: c.pct, color: "linear-gradient(90deg,#94a3b8,#64748b)" })),
  ],
  [
    { heading: "CPU", rows: [
      ["Model", "AMD EPYC 9354P 32-Core Processor"],
      ["Cores Online", String(perCore.length)],
      ["Load (1/5/15)", load.join(" / ")],
      ["Proses", String(procs)],
    ]},
    { heading: "Memory", rows: [
      ["Total", ramTotal],
      ["Used", `${ramUsed} (${ramPct}%)`],
      ["Available", "5.7 GB"],
      ["Buff/Cache", "5.2 GB"],
      ["Free", "0.8 GB"],
      ["Swap Total", "0 (none)"],
      ["Swap Used", "-"],
    ]},
    { heading: "Disk", rows: [["/", "151 GB / 197 GB (81%)"]] },
    { heading: "Network", rows: [["RX (kumulatif)", rx], ["TX (kumulatif)", tx]] },
    { heading: "Top Proses (by RAM)", rows: topProcs },
    { heading: "Sistem", rows: [
      ["Hostname", hostname],
      ["OS", osName],
      ["Kernel", "5.10.0-45-amd64"],
      ["Waktu", now],
    ]},
  ],
  "diukur dari /proc · pm2 nael-ai"
);

fs.writeFileSync("/root/vps_monitor.html", html);
console.log("written /root/vps_monitor.html", html.length, "bytes");
