import { AlertCircle } from 'lucide-react';
import { useTranslations } from 'next-intl';

export function HeaderNotice() {
  const t = useTranslations('Notices');
  return (
    <div className="space-y-2 mb-4">
      {/* Blue Info Box */}
      <div className="bg-white/80 dark:bg-slate-800/80 backdrop-blur-sm p-3 rounded-2xl shadow-sm border border-blue-100 dark:border-blue-900/30">
        <div className="flex gap-3">
            <div className="w-1 bg-blue-500 rounded-full my-1"></div>
            <div className="text-xs text-slate-600 dark:text-slate-300 space-y-1 leading-relaxed">
                <p dangerouslySetInnerHTML={{ __html: t.raw('welcome_1') }}></p>
                <p>{t('welcome_2')}</p>
                <p>{t('welcome_3')}</p>
            </div>
        </div>
      </div>

      {/* Warning Box (Simplified) */}
      <div className="bg-amber-50 dark:bg-amber-950/20 p-2 rounded-xl border border-amber-100 dark:border-amber-900/30 flex items-center justify-center text-center">
        <p className="text-[10px] text-amber-600 flex items-center gap-1 font-medium">
            <AlertCircle className="w-3 h-3" />
            {t('warning')}
        </p>
      </div>
    </div>
  );
}
