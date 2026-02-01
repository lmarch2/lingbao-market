'use client';

import { useEffect, useState } from 'react';
import { Loader2, UserPlus, AlertCircle, Eye, EyeOff } from 'lucide-react';
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { useTranslations } from 'next-intl';
import { useRouter } from '@/i18n/navigation';
import { apiUrl } from "@/lib/api";

export default function RegisterForm() {
  const t = useTranslations('Auth');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [captchaId, setCaptchaId] = useState('');
  const [captchaValue, setCaptchaValue] = useState('');
  const [captchaInput, setCaptchaInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const router = useRouter();

  const loadCaptcha = async () => {
    try {
      const res = await fetch(apiUrl('/api/v1/auth/captcha'));
      if (!res.ok) {
        throw new Error(t('error_captcha_load'));
      }
      const data = await res.json();
      setCaptchaId(data.captchaId || '');
      setCaptchaValue(data.code || '');
      setCaptchaInput('');
    } catch (err: any) {
      setCaptchaId('');
      setCaptchaValue('');
      setError(err.message || t('error_captcha_load'));
    }
  };

  useEffect(() => {
    loadCaptcha();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    if (password !== confirmPassword) {
      setError(t('error_passwords_mismatch'));
      setLoading(false);
      return;
    }
    if (!captchaInput.trim() || !captchaId) {
      setError(t('error_captcha_required'));
      setLoading(false);
      return;
    }

    try {
      const res = await fetch(apiUrl('/api/v1/auth/register'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          username: username.trim(),
          password,
          captchaId,
          captchaCode: captchaInput.trim(),
        }),
      });

      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error || t('error_login_failed'));
      }

      router.push('/auth/login');
    } catch (err: any) {
      setError(err.message);
      loadCaptcha();
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card className="w-full max-w-md mx-auto shadow-lg bg-white border-primary/10">
      <CardHeader>
        <CardTitle className="flex items-center text-primary">
          <UserPlus className="mr-2 h-5 w-5" /> {t('register_title')}
        </CardTitle>
        <CardDescription>{t('register_description')}</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">{t('username_label')}</label>
            <Input
              type="text"
              placeholder={t('pick_username_placeholder')}
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
                placeholder={t('choose_password_placeholder')}
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
            <label className="text-sm font-medium">{t('confirm_password_label')}</label>
            <div className="relative">
              <Input
                type={showConfirmPassword ? 'text' : 'password'}
                placeholder={t('reenter_password_placeholder')}
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                required
                minLength={6}
                className="pr-10"
              />
              <button
                type="button"
                onClick={() => setShowConfirmPassword((prev) => !prev)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                aria-label={showConfirmPassword ? t('hide_password') : t('show_password')}
              >
                {showConfirmPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
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
            {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : t('register_button')}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
