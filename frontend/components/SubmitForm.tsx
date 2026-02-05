import { useState } from 'react';
import { Plus, Loader2, LogIn } from 'lucide-react';
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { motion, useReducedMotion } from 'framer-motion';
import { useLocale, useTranslations } from 'next-intl';
import { useSession } from 'next-auth/react';
import Link from 'next/link';
import { apiUrl } from '@/lib/api';

type ParsedListingPaste = {
  code?: string;
  price?: string;
};

const CODE_INPUT_MAX_LENGTH = 12;
const PRICE_MAX = 999;

function getAccessToken(session: unknown): string | undefined {
  if (!session || typeof session !== 'object') return undefined;
  const token = (session as { accessToken?: unknown }).accessToken;
  return typeof token === 'string' ? token : undefined;
}

function normalizeCode(raw: string): string | undefined {
  const compact = raw.trim().replace(/\s+/g, '');
  if (!compact) return undefined;
  return compact.toUpperCase();
}

function extractBracketedText(text: string): string | undefined {
  const patterns = [
    /【([^】]{1,64})】/,
    /\[([^\]]{1,64})\]/,
    /（([^）]{1,64})）/,
    /\(([^)]{1,64})\)/,
    /《([^》]{1,64})》/,
  ];

  for (const re of patterns) {
    const match = text.match(re);
    if (match?.[1]) return match[1].trim();
  }
  return undefined;
}

function extractCodeFromText(text: string): string | undefined {
  const bracketed = extractBracketedText(text);
  if (bracketed) {
    const normalized = normalizeCode(bracketed);
    if (!normalized) return undefined;
    return normalized.slice(0, CODE_INPUT_MAX_LENGTH);
  }

  const tokens = text
    .replace(/[\r\n]+/g, ' ')
    .split(/[|｜\s]+/)
    .map((token) => normalizeCode(token))
    .filter((token): token is string => Boolean(token));

  for (let i = tokens.length - 1; i >= 0; i -= 1) {
    const token = tokens[i];
    if (token.length <= CODE_INPUT_MAX_LENGTH && /[A-Z0-9]/.test(token)) {
      return token;
    }
  }

  const compact = normalizeCode(text);
  if (compact && compact.length <= CODE_INPUT_MAX_LENGTH) return compact;

  return undefined;
}

function normalizePrice(raw: string): string | undefined {
  const num = Number(raw);
  if (!Number.isFinite(num)) return undefined;
  const value = Math.floor(num);
  if (value < 1 || value > PRICE_MAX) return undefined;
  return String(value);
}

function extractPriceFromText(text: string): string | undefined {
  const normalized = text.replace(/[,，]/g, ' ');
  const patterns = [
    /[￥¥]\s*(\d+(?:\.\d+)?)/,
    /(\d+(?:\.\d+)?)\s*(?:块|元|金)/,
    /(?:价格|价钱|售价|卖|出)\s*[:：]?\s*(\d+(?:\.\d+)?)/,
  ];

  for (const re of patterns) {
    const match = normalized.match(re);
    if (match?.[1]) {
      const value = normalizePrice(match[1]);
      if (value) return value;
    }
  }

  const numbers = normalized.match(/\d+(?:\.\d+)?/g);
  if (numbers && numbers.length === 1) {
    return normalizePrice(numbers[0]);
  }
  return undefined;
}

function parseListingPaste(text: string): ParsedListingPaste {
  const normalized = text.replace(/\r\n/g, '\n').trim();
  return {
    code: extractCodeFromText(normalized),
    price: extractPriceFromText(normalized),
  };
}

export default function SubmitForm() {
  const t = useTranslations('Submit');
  const { data: session } = useSession();
  const locale = useLocale();
  const [code, setCode] = useState('');
  const [price, setPrice] = useState('');
  const [loading, setLoading] = useState(false);
  const [focused, setFocused] = useState(false);
  const reduceMotion = useReducedMotion();

  // NOTE: Sharing no longer requires login. Keep this switch for easy rollback.
  const requireLoginToShare = false;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!code || !price) return;

    setLoading(true);
    try {
      const token = getAccessToken(session);
      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
      };
      if (token) {
        headers.Authorization = `Bearer ${token}`;
      }
      const res = await fetch(apiUrl('/api/v1/submit'), {
        method: 'POST',
        headers,
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

  if (requireLoginToShare && !session) {
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
                onPaste={(e) => {
                  const pasted = e.clipboardData.getData('text');
                  if (!pasted) return;

                  const parsed = parseListingPaste(pasted);
                  if (!parsed.code && !parsed.price) return;

                  e.preventDefault();
                  if (parsed.code) setCode(parsed.code);
                  if (parsed.price) setPrice(parsed.price);
                }}
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
            className="h-11 px-6 rounded-full bg-foreground text-background font-semibold hover:bg-foreground/90 transition-colors whitespace-nowrap"
          >
            {loading ? <Loader2 className="h-5 w-5 animate-spin" /> : <><Plus className="h-5 w-5 shrink-0" /><span className="ml-2">{t('button')}</span></>}
          </Button>
        </form>
      </div>
    </motion.div>
  );
}
