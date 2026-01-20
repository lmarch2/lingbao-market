import {NextIntlClientProvider} from 'next-intl';
import {getMessages} from 'next-intl/server';
import { Geist_Mono, Instrument_Serif, Manrope } from "next/font/google";
import { ThemeProvider } from "@/components/theme-provider";
import { SessionProvider } from "next-auth/react";
import { auth } from "@/auth";
import "../globals.css";

const manrope = Manrope({
  variable: "--font-body",
  subsets: ["latin"],
});

const instrumentSerif = Instrument_Serif({
  variable: "--font-display",
  subsets: ["latin"],
  weight: "400",
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});
 
export default async function LocaleLayout({
  children,
  params
}: {
  children: React.ReactNode;
  params: Promise<{locale: string}>;
}) {
  // Providing all messages to the client
  // side is the easiest way to get started
  const messages = await getMessages();
  const {locale} = await params;
  const session = await auth();
 
  return (
    <html lang={locale} suppressHydrationWarning>
      <body className={`${manrope.variable} ${instrumentSerif.variable} ${geistMono.variable} antialiased`}>
        <SessionProvider session={session}>
          <NextIntlClientProvider messages={messages}>
            <ThemeProvider
              attribute="class"
              defaultTheme="system"
              enableSystem
              disableTransitionOnChange
            >
              {children}
            </ThemeProvider>
          </NextIntlClientProvider>
        </SessionProvider>
      </body>
    </html>
  );
}
