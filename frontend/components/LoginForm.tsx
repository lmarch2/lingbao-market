'use client';

import { useCallback, useEffect, useState } from 'react';
import { AlertCircle, Eye, EyeOff, Loader2, LogIn } from 'lucide-react';
import { getSession, signIn } from 'next-auth/react';
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useTranslations } from 'next-intl';
import { useRouter } from '@/i18n/navigation';
import { apiUrl } from '@/lib/api';

type CaptchaResponse = {
  captchaId: string;
  code: string;
};

function getErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof Error && error.message) {
    return error.message;
  }
  return fallback;
}

export default function LoginForm() {
  const t = useTranslations('Auth');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [captchaId, setCaptchaId] = useState('');
  const [captchaValue, setCaptchaValue] = useState('');
  const [captchaInput, setCaptchaInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const router = useRouter();

  const loadCaptcha = useCallback(async () => {
    try {
      const res = await fetch(apiUrl('/api/v1/auth/captcha'));
      if (!res.ok) {
        throw new Error(t('error_captcha_load'));
      }
      const data = (await res.json()) as CaptchaResponse;
      setCaptchaId(data.captchaId || '');
      setCaptchaValue(data.code || '');
      setCaptchaInput('');
    } catch (error) {
      setCaptchaId('');
      setCaptchaValue('');
      setError(getErrorMessage(error, t('error_captcha_load')));
    }
  }, [t]);

  useEffect(() => {
    void loadCaptcha();
  }, [loadCaptcha]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const result = await signIn('credentials', {
        username: username.trim(),
        password,
        captchaId,
        captchaCode: captchaInput.trim(),
        redirect: false,
      });

      if (result?.error) {
        setError(t('error_invalid_credentials'));
        void loadCaptcha();
        return;
      }

      const session = await getSession();
      const token = session?.accessToken;
      if (token) {
        const adminCheck = await fetch(apiUrl('/api/v1/admin/users'), {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });
        if (adminCheck.ok) {
          router.push('/admin');
          return;
        }
      }
      router.push('/');
    } catch {
      setError(t('error_login_failed'));
      void loadCaptcha();
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card className="w-full max-w-md mx-auto shadow-lg bg-white border-primary/10">
      <CardHeader>
        <CardTitle className="flex items-center text-primary">
          <LogIn className="mr-2 h-5 w-5" /> {t('login_title')}
        </CardTitle>
        <CardDescription>{t('login_description')}</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">{t('username_label')}</label>
            <Input
              type="text"
              placeholder={t('username_placeholder')}
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              minLength={3}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">{t('password_label')}</label>
            <div className="relative">
              <Input
                type={showPassword ? 'text' : 'password'}
                placeholder={t('password_placeholder')}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                minLength={6}
                className="pr-10"
              />
              <button
                type="button"
                onClick={() => setShowPassword((prev) => !prev)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                aria-label={showPassword ? t('hide_password') : t('show_password')}
              >
                {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              </button>
            </div>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">{t('captcha_label')}</label>
            <div className="flex gap-2">
              <Input
                type="text"
                placeholder={t('captcha_placeholder')}
                value={captchaInput}
                onChange={(e) => setCaptchaInput(e.target.value)}
                required
              />
              <Button
                type="button"
                variant="outline"
                onClick={loadCaptcha}
                className="min-w-[110px] font-mono tracking-widest"
              >
                {captchaValue || '----'}
              </Button>
            </div>
          </div>

          {error && (
            <div className="bg-destructive/10 text-destructive text-sm p-3 rounded-md flex items-center">
              <AlertCircle className="w-4 h-4 mr-2" /> {error}
            </div>
          )}

          <Button type="submit" className="w-full" disabled={loading}>
            {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : t('login_button')}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
