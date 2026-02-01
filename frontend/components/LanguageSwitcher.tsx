'use client';

import { useLocale, useTranslations } from 'next-intl';
import { usePathname, useRouter } from '@/i18n/navigation';
import { Globe } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import type { Locale } from '@/i18n/routing';

export default function LanguageSwitcher() {
  const t = useTranslations('Language');
  const locale = useLocale();
  const router = useRouter();
  const pathname = usePathname();

  const switchLocale = (newLocale: Locale) => {
    router.replace(pathname, { locale: newLocale });
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="outline"
          size="icon"
          className="h-11 w-11 md:h-9 md:w-9 rounded-full border-black/10 dark:border-white/15 bg-transparent hover:bg-black/5 dark:hover:bg-white/5"
          aria-label={t('switch')}
        >
          <Globe className="h-[1.2rem] w-[1.2rem]" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="min-w-[120px]">
        <DropdownMenuItem
          onClick={() => switchLocale('en')}
          className={`min-h-[44px] md:min-h-[36px] ${locale === 'en' ? 'bg-accent' : ''}`}
        >
          {t('en')}
        </DropdownMenuItem>
        <DropdownMenuItem
          onClick={() => switchLocale('zh')}
          className={`min-h-[44px] md:min-h-[36px] ${locale === 'zh' ? 'bg-accent' : ''}`}
        >
          {t('zh')}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
