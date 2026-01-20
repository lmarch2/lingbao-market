import { useState } from 'react';
import { Plus, Loader2, LogIn } from 'lucide-react';
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { motion, useReducedMotion } from 'framer-motion';
import { useLocale, useTranslations } from 'next-intl';
import { useSession } from 'next-auth/react';
import Link from 'next/link';
import { apiUrl } from '@/lib/api';

export default function SubmitForm() {
  const t = useTranslations('Submit');
  const { data: session } = useSession();
  const locale = useLocale();
  const [code, setCode] = useState('');
  const [price, setPrice] = useState('');
  const [loading, setLoading] = useState(false);
  const [focused, setFocused] = useState(false);
  const reduceMotion = useReducedMotion();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!code || !price) return;

    setLoading(true);
    try {
      const token = (session as any)?.accessToken;
      const res = await fetch(apiUrl('/api/v1/submit'), {
        method: 'POST',
        headers: { 
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
          code: code.toUpperCase(),
          price: Number(price),
          server: 'S1'
        }),
      });
      if (res.ok) {
        setCode('');
        setPrice('');
      }
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  if (!session) {
    return (
      <div className="w-full max-w-2xl mx-auto mb-10 flex flex-col items-center gap-4 text-center">
        <p className="text-sm text-muted-foreground max-w-lg">
          {t('signin_prompt')}
        </p>
        <Button asChild className="h-11 px-7 rounded-full bg-foreground text-background hover:bg-foreground/90">
          <Link href={`/${locale}/auth/login`}>
            <LogIn className="mr-2 h-5 w-5" /> {t('signin_button')}
          </Link>
        </Button>
      </div>
    );
  }


  return (
    <motion.div 
      initial={reduceMotion ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: reduceMotion ? 0 : 0.5 }}
      className="w-full max-w-2xl mx-auto mb-10"
    >
      <div 
        className={`
          relative border transition-all duration-500 rounded-3xl p-2 bg-white/70 dark:bg-white/5 backdrop-blur-sm
          ${focused ? 'border-black/25 dark:border-white/30 ring-1 ring-black/10 dark:ring-white/10' : 'border-black/10 dark:border-white/10'}
        `}
      >
        <div className="pointer-events-none absolute left-6 right-6 top-2 h-px bg-[linear-gradient(90deg,transparent,rgba(15,23,42,0.35),transparent)] dark:bg-[linear-gradient(90deg,transparent,rgba(248,250,252,0.4),transparent)] bg-[length:200%_100%] animate-[flow_12s_linear_infinite] motion-reduce:animate-none" />
        <form
          onSubmit={handleSubmit}
          onFocusCapture={() => setFocused(true)}
          onBlurCapture={(event) => {
            if (!event.currentTarget.contains(event.relatedTarget as Node | null)) {
              setFocused(false);
            }
          }}
          className="flex flex-col sm:flex-row gap-3 px-4 pb-3 pt-3"
        >
          {/* Code Input Group */}
          <div className="flex-1 group relative">
            <label htmlFor="code-input" className="sr-only">{t('title')}</label>
            <div className="flex items-center w-full rounded-2xl border border-black/10 dark:border-white/15 bg-transparent ring-offset-background focus-within:ring-1 focus-within:ring-black/10 dark:focus-within:ring-white/10">
              <span className="flex select-none items-center pl-3 text-muted-foreground font-mono text-[11px] font-semibold tracking-wide">
                {t('title')}
              </span>
              <Input
                id="code-input"
                type="text"
                placeholder={t('code_placeholder')}
                value={code}
                onChange={(e) => setCode(e.target.value)}
                className="border-0 bg-transparent shadow-none focus-visible:ring-0 pl-2 h-11 text-base uppercase font-mono tracking-[0.2em]"
                maxLength={12}
                required
              />
            </div>
          </div>
          
          {/* Price Input Group */}
          <div className="w-full sm:w-48 group relative">
            <label htmlFor="price-input" className="sr-only">{t('price_placeholder')}</label>
            <div className="flex items-center w-full rounded-2xl border border-black/10 dark:border-white/15 bg-transparent ring-offset-background focus-within:ring-1 focus-within:ring-black/10 dark:focus-within:ring-white/10">
              <span className="flex select-none items-center pl-3 text-muted-foreground font-mono text-[11px] font-semibold">
                $
              </span>
              <Input
                id="price-input"
                type="number"
                placeholder={t('price_placeholder')}
                value={price}
                onChange={(e) => setPrice(e.target.value)}
                className="border-0 bg-transparent shadow-none focus-visible:ring-0 pl-2 h-11 text-base font-semibold tabular-nums"
                min={1}
                max={999}
                required
              />
            </div>
          </div>

          <Button 
            type="submit" 
            disabled={loading}
            size="lg"
            className="h-11 px-7 rounded-full bg-foreground text-background font-semibold hover:bg-foreground/90 transition-colors"
          >
            {loading ? <Loader2 className="h-5 w-5 animate-spin" /> : <div className="flex items-center"><Plus className="mr-2 h-5 w-5" /> {t('button')}</div>}
          </Button>
        </form>
      </div>
    </motion.div>
  );
}
