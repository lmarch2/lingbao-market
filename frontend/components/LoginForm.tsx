'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { AlertCircle, Loader2, LogIn } from 'lucide-react';
import { signIn } from 'next-auth/react';
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useLocale } from 'next-intl';
import { apiUrl } from '@/lib/api';

export default function LoginForm() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [captchaId, setCaptchaId] = useState('');
  const [captchaValue, setCaptchaValue] = useState('');
  const [captchaInput, setCaptchaInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const router = useRouter();
  const locale = useLocale();

  const loadCaptcha = async () => {
    try {
      const res = await fetch(apiUrl('/api/v1/auth/captcha'));
      if (!res.ok) {
        throw new Error('Failed to load captcha');
      }
      const data = await res.json();
      setCaptchaId(data.captchaId || '');
      setCaptchaValue(data.code || '');
      setCaptchaInput('');
    } catch (err: any) {
      setCaptchaId('');
      setCaptchaValue('');
      setError(err.message || 'Failed to load captcha');
    }
  };

  useEffect(() => {
    loadCaptcha();
  }, []);

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
        setError('Invalid username, password, or captcha');
        loadCaptcha();
        return;
      }

      router.push(`/${locale}`);
    } catch (err) {
      setError('Login failed. Please try again.');
      loadCaptcha();
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card className="w-full max-w-md mx-auto shadow-lg bg-white border-primary/10">
      <CardHeader>
        <CardTitle className="flex items-center text-primary">
          <LogIn className="mr-2 h-5 w-5" /> Welcome Back
        </CardTitle>
        <CardDescription>Sign in to publish prices.</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Username</label>
            <Input
              type="text"
              placeholder="Your username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              minLength={3}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Password</label>
            <Input
              type="password"
              placeholder="Your password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={6}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Captcha</label>
            <div className="flex gap-2">
              <Input
                type="text"
                placeholder="Enter captcha"
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
            {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : "Sign In"}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
