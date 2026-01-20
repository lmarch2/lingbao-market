'use client';

import { motion, useReducedMotion } from 'framer-motion';
import { Copy, Check, Clock, Trash2, TrendingUp } from 'lucide-react';
import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { formatDistanceToNow } from 'date-fns';
import { Badge } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { useTranslations } from 'next-intl';
import { apiUrl } from '@/lib/api';

interface PriceItem {
  code: string;
  price: number;
  server?: string;
  ts: number;
}

interface MarketCardProps {
  item: PriceItem;
  index: number;
  adminToken?: string;
  canDelete?: boolean;
  onDeleted?: () => void;
}

export default function MarketCard({
  item,
  index,
  adminToken,
  canDelete = false,
  onDeleted,
}: MarketCardProps) {
  const t = useTranslations('Card');
  const [copied, setCopied] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const reduceMotion = useReducedMotion();

  const isHigh = item.price >= 900;
  
  const handleCopy = () => {
    // Clean, direct copy text
    const text = t('copy_text', { price: item.price, code: item.code });
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleDelete = async () => {
    if (!adminToken) {
      return;
    }
    const confirmed = window.confirm(t('confirm_delete', { code: item.code }));
    if (!confirmed) {
      return;
    }
    setDeleting(true);
    try {
      const res = await fetch(apiUrl(`/api/v1/admin/prices/${item.code}`), {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${adminToken}`,
        },
      });
      if (res.ok) {
        onDeleted?.();
      }
    } finally {
      setDeleting(false);
    }
  };

  return (
    <motion.div
      layout
      initial={reduceMotion ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, scale: 0.95 }}
      transition={{ duration: 0.4, ease: [0.23, 1, 0.32, 1], delay: index * 0.05 }}
      whileHover={reduceMotion ? undefined : { y: -3, transition: { duration: 0.2 } }}
      className="h-full"
    >
      <Card className="h-full group relative overflow-hidden border-black/10 dark:border-white/10 bg-white/70 dark:bg-white/5 backdrop-blur-sm hover:border-black/25 dark:hover:border-white/25 transition-colors duration-300">

        {isHigh && (
          <div className="absolute right-4 top-4 h-px w-16 bg-black/30 dark:bg-white/40" />
        )}

        <div className="p-5 flex flex-col h-full gap-4">
          
          {/* Header: Code & Badge */}
          <div className="flex justify-between items-start">
            <div className="flex flex-col">
              <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-semibold mb-1">{t('label')}</span>
              <span className="font-mono text-xl font-bold tracking-tight text-foreground">
                {item.code}
              </span>
            </div>
            {isHigh && (
              <Badge variant="secondary" className="bg-transparent border border-black/10 dark:border-white/20 text-muted-foreground font-medium">
                <TrendingUp className="w-3 h-3 mr-1" />
                {t('hot')}
              </Badge>
            )}
          </div>

          {/* Body: Price */}
          <div className="flex-1 flex items-baseline gap-1">
            <span className="text-4xl font-extrabold tracking-tight text-foreground tabular-nums">
              {item.price}
            </span>
            <span className="text-sm text-muted-foreground font-medium">{t('currency')}</span>
          </div>

          {/* Footer: Meta & Action */}
          <div className="flex items-center justify-between pt-4 border-t border-black/10 dark:border-white/10 mt-auto">
            <div className="flex items-center text-xs text-muted-foreground">
              <Clock className="w-3 h-3 mr-1" />
              {formatDistanceToNow(item.ts, { addSuffix: true }).replace('about ', '')}
            </div>
            
            <div className="flex items-center gap-2">
              {canDelete && adminToken && (
                <Button
                  onClick={handleDelete}
                  size="icon"
                  variant="ghost"
                  disabled={deleting}
                  aria-label={t('delete')}
                  className="h-8 w-8 rounded-full text-destructive hover:bg-destructive/10"
                >
                  <Trash2 className="w-4 h-4" />
                </Button>
              )}
              <Button
                onClick={handleCopy}
                size="icon"
                variant="ghost"
                aria-label={copied ? t('copy_success') : t('copy_text', { price: item.price, code: item.code })}
                className={`h-8 w-8 rounded-full transition-all duration-300 ${
                  copied ? 'bg-black/10 text-foreground hover:bg-black/15 dark:bg-white/10 dark:hover:bg-white/20' : 'hover:bg-black/5 hover:text-foreground dark:hover:bg-white/10'
                }`}
              >
                {copied ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
              </Button>
            </div>
          </div>
        </div>
      </Card>
    </motion.div>
  );
}
