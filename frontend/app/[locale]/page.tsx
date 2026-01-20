'use client';

import PriceFeed from '@/components/PriceFeed';
import SubmitForm from '@/components/SubmitForm';
import AccountMenu from '@/components/AccountMenu';
import { ModeToggle } from '@/components/mode-toggle';
import { Github } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useTranslations } from 'next-intl';
import { motion, useReducedMotion } from 'framer-motion';

export default function Home() {
  const tNav = useTranslations('Navbar');
  const tSubmit = useTranslations('Submit');
  const tFeed = useTranslations('Feed');
  const tFooter = useTranslations('Footer');
  const reduceMotion = useReducedMotion();

  return (
    <div className="min-h-screen bg-[#f8f7f4] dark:bg-[#0f1114] font-sans selection:bg-black/10 selection:text-black dark:selection:bg-white/10 dark:selection:text-white relative overflow-x-hidden">
      <div className="fixed inset-0 z-0 pointer-events-none">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,_rgba(15,23,42,0.08),_transparent_48%)] dark:bg-[radial-gradient(circle_at_top,_rgba(148,163,184,0.1),_transparent_48%)]" />
        <div className="absolute inset-0 bg-[linear-gradient(90deg,rgba(15,23,42,0.05)_1px,transparent_1px),linear-gradient(0deg,rgba(15,23,42,0.05)_1px,transparent_1px)] bg-[size:56px_56px] opacity-40 dark:opacity-15" />
        <div className="absolute inset-y-0 left-[-20%] w-[140%] bg-[linear-gradient(110deg,transparent,rgba(15,23,42,0.08),transparent)] bg-[length:200%_100%] animate-[flow_18s_linear_infinite] motion-reduce:animate-none" />
        <motion.div
          animate={reduceMotion ? undefined : { y: [0, -14, 0], x: [0, 12, 0] }}
          transition={{ duration: 16, repeat: Infinity, ease: 'easeInOut' }}
          className="absolute top-[8%] right-[8%] h-40 w-40 rounded-full border border-black/5 dark:border-white/10 blur-2xl"
        />
        <motion.div
          animate={reduceMotion ? undefined : { y: [0, 18, 0], x: [0, -16, 0] }}
          transition={{ duration: 20, repeat: Infinity, ease: 'easeInOut' }}
          className="absolute bottom-[10%] left-[10%] h-56 w-56 rounded-full border border-black/5 dark:border-white/10 blur-3xl"
        />
      </div>

      {/* Navbar */}
      <header className="sticky top-0 z-50 w-full border-b border-black/5 dark:border-white/10 bg-[#f8f7f4]/90 dark:bg-[#0f1114]/80 backdrop-blur-xl">
        <div className="container flex h-16 items-center justify-between mx-auto px-4 md:px-6">
          <div className="flex items-center gap-2">
            <div className="h-8 w-8 rounded-full border border-black/10 dark:border-white/20 flex items-center justify-center">
              <span className="text-xs font-semibold tracking-[0.3em] ml-[0.2em]">LB</span>
            </div>
            <div className="hidden sm:flex flex-col leading-tight">
              <span className="font-semibold text-sm tracking-tight">{tNav('title')}</span>
              <span className="text-xs text-muted-foreground">{tNav('subtitle')}</span>
            </div>
          </div>
          
          <div className="flex items-center gap-2">
            <ModeToggle />
            <AccountMenu />
            <Button variant="ghost" size="icon" className="rounded-full" aria-label="GitHub">
              <Github className="h-5 w-5" />
            </Button>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 md:px-6 py-10 md:py-14 relative z-10">
        <section id="submit" className="mt-6">
          <div className="flex flex-col gap-2 mb-6">
            <h2 className="text-2xl sm:text-3xl font-semibold font-display">{tSubmit('title')}</h2>
            <p className="text-sm text-muted-foreground max-w-2xl">{tSubmit('description')}</p>
          </div>
          <SubmitForm />
        </section>

        <section id="market" className="mt-16">
          <div className="flex items-center justify-between flex-wrap gap-4 mb-6">
            <div>
              <h2 className="text-2xl sm:text-3xl font-semibold font-display">{tFeed('title')}</h2>
              <p className="text-sm text-muted-foreground mt-2">{tFeed('subtitle')}</p>
            </div>
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <div className="h-2 w-2 rounded-full bg-black/40 dark:bg-white/40 animate-pulse" />
              {tFeed('live_label')}
            </div>
          </div>
          <PriceFeed />
        </section>
      </main>

      
      {/* Minimal Footer */}
      <footer className="py-8 text-center text-sm text-muted-foreground border-t border-black/5 dark:border-white/10 bg-white/40 dark:bg-white/5 backdrop-blur-sm">
        <div className="container mx-auto">
            <p>{tFooter('copyright')}</p>
        </div>
      </footer>
    </div>
  );
}
