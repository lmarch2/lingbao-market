"use client";

import { useMemo } from "react";
import Link from "next/link";
import { useLocale, useTranslations } from "next-intl";
import { signOut, useSession } from "next-auth/react";
import { LogIn, LogOut, UserPlus } from "lucide-react";
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
          className="h-9 w-9 rounded-full border-black/10 dark:border-white/15 bg-transparent hover:bg-black/5 dark:hover:bg-white/5"
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
      <DropdownMenuContent align="end" sideOffset={8} className="w-56">
        {session ? (
          <>
            <DropdownMenuLabel className="flex flex-col gap-1">
              <span className="text-xs uppercase tracking-[0.3em] text-muted-foreground">
                {t("title")}
              </span>
              <span className="text-sm font-semibold">
                {session.user?.name || t("anonymous")}
              </span>
              {session.user?.email && (
                <span className="text-xs text-muted-foreground">
                  {session.user.email}
                </span>
              )}
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={() => signOut({ callbackUrl: `/${locale}` })}>
              <LogOut className="h-4 w-4" />
              {t("sign_out")}
            </DropdownMenuItem>
          </>
        ) : (
          <>
            <DropdownMenuLabel className="text-xs uppercase tracking-[0.3em] text-muted-foreground">
              {t("guest")}
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem asChild>
              <Link href={`/${locale}/auth/login`} className="flex items-center gap-2">
                <LogIn className="h-4 w-4" />
                {t("sign_in")}
              </Link>
            </DropdownMenuItem>
            <DropdownMenuItem asChild>
              <Link href={`/${locale}/auth/register`} className="flex items-center gap-2">
                <UserPlus className="h-4 w-4" />
                {t("register")}
              </Link>
            </DropdownMenuItem>
          </>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
