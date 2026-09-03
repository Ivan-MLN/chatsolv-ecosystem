import type { Metadata } from "next";
import { Geist, Plus_Jakarta_Sans } from "next/font/google";
import "./globals.css";

const geist = Geist({
  subsets: ["latin"],
  variable: "--font-geist",
});

const plusJakartaSans = Plus_Jakarta_Sans({
  subsets: ["latin"],
  weight: ["400", "500", "600", "700", "800"],
  variable: "--font-pjs",
});

export const metadata: Metadata = {
  title: "ChatSolv — Solusi Customer Service WhatsApp Otomatis",
  description: "ChatSolv menjaga percakapan tetap berjalan ketika tim Anda sedang sibuk, sehingga pelanggan tidak kehilangan momentum.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="id">
      <body className={`${geist.variable} ${plusJakartaSans.variable} bg-[#d6ebd8] text-[#162b19] antialiased`}>
        {children}
      </body>
    </html>
  );
}
