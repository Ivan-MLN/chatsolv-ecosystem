#!/usr/bin/env node
import { join } from "node:path";
import { tmpdir } from "node:os";
import { writeFileSync } from "node:fs";
import makeWASocket, { useMultiFileAuthState, DisconnectReason } from "@whiskeysockets/baileys";
import { Boom } from "@hapi/boom";

const authDir = join(tmpdir(), "nael-voip-auth");
const qrPath = "/root/nael-ai/voip-qr.png";

async function main() {
  console.log(`Auth dir: ${authDir}`);
  
  const { state, saveCreds } = await useMultiFileAuthState(authDir);
  
  const sock = makeWASocket({
    auth: state,
    printQRInTerminal: false,
  });

  sock.ev.on("creds.update", saveCreds);

  sock.ev.on("connection.update", async (update) => {
    const { connection, lastDisconnect, qr } = update;

    if (qr) {
      console.log("QR code received, generating image...");
      // Generate QR as image
      const QRCode = await import("qrcode");
      const qrBuffer = await QRCode.toBuffer(qr, { 
        type: "png", 
        width: 512,
        margin: 2 
      });
      
      writeFileSync(qrPath, qrBuffer);
      console.log(`✅ QR saved to: ${qrPath}`);
      console.log("Scan this QR with the WhatsApp number you want to use for voice calls.");
    }

    if (connection === "close") {
      const shouldReconnect =
        (lastDisconnect?.error as Boom)?.output?.statusCode !== DisconnectReason.loggedOut;
      console.log("Connection closed, reconnect:", shouldReconnect);
      if (shouldReconnect) {
        setTimeout(main, 2000);
      } else {
        process.exit(0);
      }
    } else if (connection === "open") {
      console.log("✅ VOIP client connected!");
      console.log("Auth saved to:", authDir);
      const me = sock.user?.id || "unknown";
      console.log("Number:", me);
      console.log("\nKeep this running to maintain the connection, or Ctrl+C to stop.");
      console.log("The bot can now use wa_voice_call tool!");
    }
  });
}

main().catch(console.error);
