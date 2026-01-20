import { useState } from 'react';
import useSWR from 'swr';
import { AnimatePresence, motion, useReducedMotion } from 'framer-motion';
import { ArrowDownUp, Loader2, RefreshCw } from 'lucide-react';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import MarketCard from './MarketCard';
import { Button } from '@/components/ui/button';
import { useTranslations } from 'next-intl';
import { apiUrl } from '@/lib/api';

interface PriceItem {
  code: string;
  price: number;
  server?: string;
  ts: number;
}

const fetcher = (url: string) => fetch(url).then((res) => res.json());

export default function PriceFeed() {
  const t = useTranslations('Feed');
  const [sortBy, setSortBy] = useState<'time' | 'price'>('time');
  const reduceMotion = useReducedMotion();
  
  const { data: prices, error, isLoading, mutate } = useSWR<PriceItem[]>(
    apiUrl(`/api/v1/feed?sort=${sortBy}`),
    fetcher,
    { 
      refreshInterval: 5000, // Slow down to 5s to reduce load
      revalidateOnFocus: false, // Don't spam when tab switching
      errorRetryCount: 3
    }
  );

  return (
    <div className="w-full space-y-6">
        {/* Controls Bar */}
        <div className="flex flex-col sm:flex-row justify-between items-center gap-4 px-2">
            <div className="flex items-center space-x-2 text-sm text-muted-foreground">
                <span className="relative flex h-2.5 w-2.5">
                  <span className={`${reduceMotion ? '' : 'animate-ping'} absolute inline-flex h-full w-full rounded-full bg-black/30 dark:bg-white/30 opacity-60`}></span>
                  <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-black/40 dark:bg-white/40"></span>
                </span>
                <span className="font-medium">{t('live_label')}</span>
                <span className="text-xs opacity-50">• {prices?.length || 0} {t('listings')}</span>
            </div>
            
            <div className="flex items-center gap-2 w-full sm:w-auto">
                <Select value={sortBy} onValueChange={(v: 'time' | 'price') => setSortBy(v)}>
                    <SelectTrigger className="w-full sm:w-[160px] h-9 bg-white/70 dark:bg-white/5 backdrop-blur-sm border-black/10 dark:border-white/10">
                        <SelectValue placeholder={t('sort_placeholder')} />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="time">{t('sort_latest')}</SelectItem>
                        <SelectItem value="price">{t('sort_highest')}</SelectItem>
                    </SelectContent>
                </Select>
                
                <Button 
                    variant="outline" 
                    size="icon" 
                    className="h-9 w-9 shrink-0 border-black/10 dark:border-white/10"
                    onClick={() => mutate()}
                    aria-label={t('refresh')}
                >
                    <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
                </Button>
            </div>
        </div>

        {/* Content State */}
        {isLoading && !prices && (
            <div className="flex flex-col items-center justify-center py-20 gap-4">
                <Loader2 className="h-10 w-10 animate-spin text-black/40 dark:text-white/40" />
                <p className="text-sm text-muted-foreground">{t('loading')}</p>
            </div>
        )}

        {error && (
            <div className="rounded-xl border border-destructive/20 bg-destructive/5 p-8 text-center">
                <p className="text-destructive font-medium">{t('error')}</p>
                <Button variant="link" onClick={() => mutate()} className="mt-2 text-destructive">{t('retry')}</Button>
            </div>
        )}

        {/* Responsive Grid System */}
        <motion.div 
            layout={!reduceMotion}
            className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 sm:gap-6"
        >
            <AnimatePresence mode='popLayout'>
                {prices?.map((item, idx) => (
                    <MarketCard 
                        key={`${item.code}-${item.ts}`} 
                        item={item} 
                        index={idx}
                    />
                ))}
            </AnimatePresence>
        </motion.div>
        
        {!isLoading && prices?.length === 0 && (
            <div className="flex flex-col items-center justify-center py-20 text-muted-foreground gap-2">
                <div className="h-12 w-12 rounded-full border border-black/10 dark:border-white/10 flex items-center justify-center">
                    <ArrowDownUp className="h-6 w-6 opacity-20" />
                </div>
                <p>{t('empty')}</p>
            </div>
        )}
    </div>
  );
}
