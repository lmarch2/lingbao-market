"use client";

import { useMemo } from "react";
import Link from "next/link";
import { useLocale, useTranslations } from "next-intl";
import { signOut, useSession } from "next-auth/react";
import { LogIn, LogOut, ShieldCheck, UserPlus } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

export default function AccountMenu() {
  const { data: session } = useSession();
  const locale = useLocale();
  const t = useTranslations("Account");

  const initials = useMemo(() => {
    const name = session?.user?.name || session?.user?.email || "";
    const parts = name.trim().split(/\s+/).slice(0, 2);
    return parts.map((part) => part[0]?.toUpperCase() ?? "").join("") || "U";
  }, [session?.user?.name, session?.user?.email]);

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="outline"
          size="icon"
          className="h-11 w-11 md:h-9 md:w-9 rounded-full border-black/10 dark:border-white/15 bg-transparent hover:bg-black/5 dark:hover:bg-white/5"
          aria-label={t("menu_aria")}
        >
          {session?.user?.image ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              src={session.user.image}
              alt={session.user.name ?? t("avatar_alt")}
              className="h-8 w-8 rounded-full object-cover"
            />
          ) : (
            <span className="text-xs font-semibold tracking-wide">{initials}</span>
          )}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" sideOffset={8} className="w-auto min-w-[180px] max-w-[280px]">
        <>
          <DropdownMenuLabel className="flex flex-col gap-1">
            <span className="text-xs uppercase tracking-[0.3em] text-muted-foreground">
              {session ? t("title") : t("guest")}
            </span>
            {session && (
              <>
                <span className="text-sm font-semibold">
                  {session.user?.name || t("anonymous")}
                </span>
                {session.user?.email && (
                  <span className="text-xs text-muted-foreground">
                    {session.user.email}
                  </span>
                )}
              </>
            )}
          </DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem asChild>
            <Link href={`/${locale}/admin`} className="flex items-center gap-2 min-h-[44px] md:min-h-[36px]">
              <ShieldCheck className="h-4 w-4" />
              {t("admin_console")}
            </Link>
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          {session ? (
            <DropdownMenuItem 
              onClick={() => signOut({ callbackUrl: `/${locale}` })}
              className="min-h-[44px] md:min-h-[36px]"
            >
              <LogOut className="h-4 w-4" />
              {t("sign_out")}
            </DropdownMenuItem>
          ) : (
            <>
              <DropdownMenuItem asChild>
                <Link href={`/${locale}/auth/login`} className="flex items-center gap-2 min-h-[44px] md:min-h-[36px]">
                  <LogIn className="h-4 w-4" />
                  {t("sign_in")}
                </Link>
              </DropdownMenuItem>
              <DropdownMenuItem asChild>
                <Link href={`/${locale}/auth/register`} className="flex items-center gap-2 min-h-[44px] md:min-h-[36px]">
                  <UserPlus className="h-4 w-4" />
                  {t("register")}
                </Link>
              </DropdownMenuItem>
            </>
          )}
        </>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
