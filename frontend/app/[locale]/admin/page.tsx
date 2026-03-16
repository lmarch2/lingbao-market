"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import { useSession } from "next-auth/react";
import { apiUrl } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

type AdminUser = {
  id: string;
  username: string;
  isAdmin: boolean;
  banned: boolean;
};

type SessionWithToken = {
  accessToken?: string;
};

type FeedbackMessage = {
  id: string;
  code: string;
  reason: string;
  reporter: string;
  createdAt: number;
  resolved: boolean;
  resolvedAt?: number;
  resolvedBy?: string;
  action?: string;
  removedTime?: number;
  removedPrice?: number;
};

type AdminLogEntry = {
  id: string;
  type: string;
  message: string;
  actor: string;
  timestamp: number;
  metadata?: Record<string, string>;
};

type ManualBilibiliImportResult = {
  status: string;
  imported: number;
  warning?: string;
};

function getLogMetadataEntries(metadata?: Record<string, string>) {
  return Object.entries(metadata ?? {}).filter(([, value]) => value.trim() !== "");
}

export default function AdminPage() {
  const t = useTranslations("Admin");
  const locale = useLocale();
  const { data: session } = useSession();
  const token = (session as SessionWithToken | null)?.accessToken;

  const [users, setUsers] = useState<AdminUser[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [accessDenied, setAccessDenied] = useState(false);
  const [creating, setCreating] = useState(false);
  const [feedback, setFeedback] = useState<FeedbackMessage[]>([]);
  const [feedbackLoading, setFeedbackLoading] = useState(false);
  const [logs, setLogs] = useState<AdminLogEntry[]>([]);
  const [logsLoading, setLogsLoading] = useState(false);
  const [importing, setImporting] = useState(false);
  const [importResult, setImportResult] = useState<ManualBilibiliImportResult | null>(null);
  const [resolvingId, setResolvingId] = useState<string | null>(null);
  const [form, setForm] = useState({
    username: "",
    password: "",
    isAdmin: false,
  });

  const headers = useMemo(() => {
    const base: Record<string, string> = {
      "Content-Type": "application/json",
    };
    if (token) {
      base.Authorization = `Bearer ${token}`;
    }
    return base;
  }, [token]);

  const logMetadataLabels = useMemo<Record<string, string>>(
    () => ({
      action: t("log_meta_action"),
      banned: t("log_meta_banned"),
      code: t("log_meta_code"),
      commentPages: t("log_meta_comment_pages"),
      error: t("log_meta_error"),
      feedbackId: t("log_meta_feedback_id"),
      imported: t("log_meta_imported"),
      isAdmin: t("log_meta_is_admin"),
      keyword: t("log_meta_keyword"),
      limit: t("log_meta_limit"),
      minPrice: t("log_meta_min_price"),
      removedPrice: t("log_meta_removed_price"),
      removedTime: t("log_meta_removed_time"),
      result: t("log_meta_result"),
      searchPages: t("log_meta_search_pages"),
      searchPageSize: t("log_meta_search_page_size"),
      server: t("log_meta_server"),
      username: t("log_meta_username"),
      warning: t("log_meta_warning"),
    }),
    [t]
  );

  const formatLogMetadataValue = useCallback(
    (key: string, value: string) => {
      if ((key === "banned" || key === "isAdmin") && value === "true") {
        return t("log_value_true");
      }
      if ((key === "banned" || key === "isAdmin") && value === "false") {
        return t("log_value_false");
      }
      if (key === "action" && value === "keep") {
        return t("log_action_keep");
      }
      if (key === "action" && value === "delete") {
        return t("log_action_delete");
      }
      if (key === "result" && value === "success") {
        return t("log_result_success");
      }
      if (key === "result" && value === "warning") {
        return t("log_result_warning");
      }
      if (key === "result" && value === "error") {
        return t("log_result_error");
      }
      return value;
    },
    [t]
  );

  const fetchUsers = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const res = await fetch(apiUrl("/api/v1/admin/users"), { headers });
      if (!res.ok) {
        if (res.status === 401 || res.status === 403) {
          setAccessDenied(true);
          return;
        }
        throw new Error(t("error_load"));
      }
      const data = (await res.json()) as AdminUser[];
      setAccessDenied(false);
      setUsers(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("error_load"));
    } finally {
      setLoading(false);
    }
  }, [headers, t]);

  useEffect(() => {
    fetchUsers();
  }, [fetchUsers]);

  const fetchFeedback = useCallback(async () => {
    setFeedbackLoading(true);
    setError("");
    try {
      const res = await fetch(apiUrl("/api/v1/admin/feedback?includeResolved=true"), { headers });
      if (!res.ok) {
        const payload = await res.json().catch(() => ({}));
        throw new Error(payload.error || t("error_feedback_load"));
      }
      const data = (await res.json()) as FeedbackMessage[];
      setFeedback(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("error_feedback_load"));
    } finally {
      setFeedbackLoading(false);
    }
  }, [headers, t]);

  const fetchLogs = useCallback(async () => {
    setLogsLoading(true);
    setError("");
    try {
      const res = await fetch(apiUrl("/api/v1/admin/logs"), { headers });
      if (!res.ok) {
        const payload = await res.json().catch(() => ({}));
        throw new Error(payload.error || t("error_logs_load"));
      }
      const data = (await res.json()) as AdminLogEntry[];
      setLogs(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("error_logs_load"));
    } finally {
      setLogsLoading(false);
    }
  }, [headers, t]);

  useEffect(() => {
    fetchFeedback();
    fetchLogs();
  }, [fetchFeedback, fetchLogs]);

  const handleCreate = async () => {
    setCreating(true);
    setError("");
    try {
      const res = await fetch(apiUrl("/api/v1/admin/users"), {
        method: "POST",
        headers,
        body: JSON.stringify({
          username: form.username.trim(),
          password: form.password,
          isAdmin: form.isAdmin,
        }),
      });
      if (!res.ok) {
        const payload = await res.json().catch(() => ({}));
        throw new Error(payload.error || t("error_create"));
      }
      setForm({ username: "", password: "", isAdmin: false });
      await fetchUsers();
    } catch (err) {
      setError(err instanceof Error ? err.message : t("error_create"));
    } finally {
      setCreating(false);
    }
  };

  const handleBanToggle = async (user: AdminUser) => {
    setError("");
    try {
      const res = await fetch(apiUrl(`/api/v1/admin/users/${user.username}/ban`), {
        method: "PATCH",
        headers,
        body: JSON.stringify({ banned: !user.banned }),
      });
      if (!res.ok) {
        const payload = await res.json().catch(() => ({}));
        throw new Error(payload.error || t("error_update"));
      }
      await fetchUsers();
    } catch (err) {
      setError(err instanceof Error ? err.message : t("error_update"));
    }
  };

  const handleDeleteUser = async (user: AdminUser) => {
    if (!confirm(t("confirm_delete_user", { username: user.username }))) {
      return;
    }
    setError("");
    try {
      const res = await fetch(apiUrl(`/api/v1/admin/users/${user.username}`), {
        method: "DELETE",
        headers,
      });
      if (!res.ok) {
        const payload = await res.json().catch(() => ({}));
        throw new Error(payload.error || t("error_delete_user"));
      }
      await fetchUsers();
    } catch (err) {
      setError(err instanceof Error ? err.message : t("error_delete_user"));
    }
  };

  const handleResolveFeedback = async (item: FeedbackMessage, action: "keep" | "delete") => {
    if (item.resolved) {
      return;
    }

    if (action === "delete") {
      const confirmed = confirm(t("confirm_delete_code", { code: item.code }));
      if (!confirmed) {
        return;
      }
    }

    setResolvingId(item.id);
    setError("");
    try {
      const res = await fetch(apiUrl(`/api/v1/admin/feedback/${item.id}/resolve`), {
        method: "POST",
        headers,
        body: JSON.stringify({ action }),
      });
      if (!res.ok) {
        const payload = await res.json().catch(() => ({}));
        throw new Error(payload.error || t("error_feedback_resolve"));
      }
      await Promise.all([fetchFeedback(), fetchLogs()]);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("error_feedback_resolve"));
    } finally {
      setResolvingId(null);
    }
  };

  const handleManualImport = async () => {
    setImporting(true);
    setImportResult(null);
    setError("");
    try {
      const res = await fetch(apiUrl("/api/v1/admin/imports/bilibili"), {
        method: "POST",
        headers,
      });
      const payload = (await res.json().catch(() => ({}))) as Partial<ManualBilibiliImportResult> & {
        error?: string;
        details?: string;
      };
      if (!res.ok) {
        throw new Error(payload.details || payload.error || t("error_import"));
      }
      setImportResult({
        status: payload.status || "ok",
        imported: payload.imported ?? 0,
        warning: payload.warning,
      });
      await fetchLogs();
    } catch (err) {
      setError(err instanceof Error ? err.message : t("error_import"));
    } finally {
      setImporting(false);
    }
  };

  return (
    <div className="container mx-auto px-4 md:px-6 py-10">
      <div className="flex flex-col gap-3 mb-8">
        <h1 className="text-3xl font-semibold font-display">{t("title")}</h1>
        <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
        {error && <p className="text-sm text-destructive">{error}</p>}
      </div>

      {accessDenied && (
        <div className="rounded-3xl border border-black/10 dark:border-white/10 bg-white/70 dark:bg-white/5 p-8 text-center">
          <p className="text-sm text-muted-foreground">{t("no_access")}</p>
          <Button asChild className="mt-6 h-10 px-6">
            <a href={`/${locale}/auth/login`}>{t("go_login")}</a>
          </Button>
        </div>
      )}

      {!accessDenied && (
        <Tabs defaultValue="users" className="space-y-4">
          <TabsList>
            <TabsTrigger value="users">{t("tab_users")}</TabsTrigger>
            <TabsTrigger value="messages">{t("tab_messages")}</TabsTrigger>
            <TabsTrigger value="logs">{t("tab_logs")}</TabsTrigger>
          </TabsList>

          <TabsContent value="users">
            <div className="grid grid-cols-1 lg:grid-cols-[1fr_320px] gap-6 lg:gap-8">
              <section className="rounded-3xl border border-black/10 dark:border-white/10 bg-white/70 dark:bg-white/5 p-4 sm:p-6">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-4">
                  <h2 className="text-lg font-semibold">{t("users_title")}</h2>
                  <Button variant="outline" size="sm" onClick={fetchUsers} disabled={loading} className="w-full sm:w-auto">
                    {loading ? t("loading") : t("refresh")}
                  </Button>
                </div>

                <div className="overflow-x-auto -mx-4 sm:mx-0">
                  <table className="w-full text-sm min-w-[500px]">
                    <thead className="text-xs uppercase tracking-[0.2em] text-muted-foreground">
                      <tr>
                        <th className="py-2 text-left">{t("col_user")}</th>
                        <th className="py-2 text-left">{t("col_role")}</th>
                        <th className="py-2 text-left">{t("col_status")}</th>
                        <th className="py-2 text-right">{t("col_actions")}</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-black/5 dark:divide-white/10">
                      {users.map((user) => (
                        <tr key={user.id}>
                          <td className="py-3 font-medium">{user.username}</td>
                          <td className="py-3 text-muted-foreground">
                            {user.isAdmin ? t("role_admin") : t("role_user")}
                          </td>
                          <td className="py-3 text-muted-foreground">
                            {user.banned ? t("status_banned") : t("status_active")}
                          </td>
                          <td className="py-3 text-right">
                            <div className="flex items-center justify-end gap-1 sm:gap-2">
                              <Button
                                size="sm"
                                variant="outline"
                                onClick={() => handleBanToggle(user)}
                                className="h-9 min-h-[36px] px-2 sm:px-3"
                              >
                                {user.banned ? t("unban") : t("ban")}
                              </Button>
                              <Button
                                size="sm"
                                variant="ghost"
                                className="text-destructive h-9 min-h-[36px] px-2 sm:px-3"
                                onClick={() => handleDeleteUser(user)}
                              >
                                {t("delete")}
                              </Button>
                            </div>
                          </td>
                        </tr>
                      ))}
                      {users.length === 0 && !loading && (
                        <tr>
                          <td colSpan={4} className="py-6 text-center text-muted-foreground">
                            {t("no_users")}
                          </td>
                        </tr>
                      )}
                    </tbody>
                  </table>
                </div>
              </section>

              <aside className="space-y-6">
                <section className="rounded-3xl border border-black/10 dark:border-white/10 bg-white/70 dark:bg-white/5 p-6 space-y-4">
                  <div>
                    <h2 className="text-lg font-semibold">{t("create_title")}</h2>
                    <p className="text-xs text-muted-foreground">{t("create_desc")}</p>
                  </div>
                  <div className="space-y-3">
                    <Input
                      placeholder={t("username_placeholder")}
                      value={form.username}
                      onChange={(event) =>
                        setForm((prev) => ({ ...prev, username: event.target.value }))
                      }
                    />
                    <Input
                      type="password"
                      placeholder={t("password_placeholder")}
                      value={form.password}
                      onChange={(event) =>
                        setForm((prev) => ({ ...prev, password: event.target.value }))
                      }
                    />
                    <label className="flex items-center gap-2 text-sm text-muted-foreground">
                      <input
                        type="checkbox"
                        checked={form.isAdmin}
                        onChange={(event) =>
                          setForm((prev) => ({ ...prev, isAdmin: event.target.checked }))
                        }
                        className="h-4 w-4 rounded border-black/10 dark:border-white/20"
                      />
                      {t("is_admin")}
                    </label>
                    <Button
                      className="w-full h-10"
                      onClick={handleCreate}
                      disabled={creating || !form.username || !form.password}
                    >
                      {creating ? t("creating") : t("create")}
                    </Button>
                  </div>
                </section>
              </aside>
            </div>
          </TabsContent>

          <TabsContent value="messages">
            <section className="rounded-3xl border border-black/10 dark:border-white/10 bg-white/70 dark:bg-white/5 p-4 sm:p-6">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-4">
                <h2 className="text-lg font-semibold">{t("messages_title")}</h2>
                <Button variant="outline" size="sm" onClick={fetchFeedback} disabled={feedbackLoading} className="w-full sm:w-auto">
                  {feedbackLoading ? t("loading") : t("refresh")}
                </Button>
              </div>

              <div className="overflow-x-auto -mx-4 sm:mx-0">
                <table className="w-full text-sm min-w-[760px]">
                  <thead className="text-xs uppercase tracking-[0.2em] text-muted-foreground">
                    <tr>
                      <th className="py-2 text-left">{t("col_code")}</th>
                      <th className="py-2 text-left">{t("col_feedback_reason")}</th>
                      <th className="py-2 text-left">{t("col_feedback_reporter")}</th>
                      <th className="py-2 text-left">{t("col_feedback_status")}</th>
                      <th className="py-2 text-right">{t("col_actions")}</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-black/5 dark:divide-white/10">
                    {feedback.map((item) => (
                      <tr key={item.id}>
                        <td className="py-3 font-mono font-medium">{item.code}</td>
                        <td className="py-3 text-muted-foreground max-w-[320px] truncate" title={item.reason}>
                          {item.reason}
                        </td>
                        <td className="py-3 text-muted-foreground">{item.reporter}</td>
                        <td className="py-3 text-muted-foreground">
                          {item.resolved ? t("feedback_status_resolved") : t("feedback_status_open")}
                        </td>
                        <td className="py-3 text-right">
                          <div className="flex items-center justify-end gap-1 sm:gap-2">
                            <Button
                              size="sm"
                              variant="outline"
                              disabled={item.resolved || resolvingId === item.id}
                              onClick={() => handleResolveFeedback(item, "keep")}
                              className="h-9 min-h-[36px] px-2 sm:px-3"
                            >
                              {t("feedback_keep")}
                            </Button>
                            <Button
                              size="sm"
                              variant="ghost"
                              className="text-destructive h-9 min-h-[36px] px-2 sm:px-3"
                              disabled={item.resolved || resolvingId === item.id}
                              onClick={() => handleResolveFeedback(item, "delete")}
                            >
                              {t("feedback_delete")}
                            </Button>
                          </div>
                        </td>
                      </tr>
                    ))}
                    {feedback.length === 0 && !feedbackLoading && (
                      <tr>
                        <td colSpan={5} className="py-6 text-center text-muted-foreground">
                          {t("no_feedback")}
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </section>
          </TabsContent>

          <TabsContent value="logs">
            <section className="rounded-3xl border border-black/10 dark:border-white/10 bg-white/70 dark:bg-white/5 p-4 sm:p-6">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-4">
                <div className="space-y-1">
                  <h2 className="text-lg font-semibold">{t("logs_title")}</h2>
                  <p className="text-xs text-muted-foreground">{t("import_desc")}</p>
                  {importResult && (
                    <p className="text-xs text-muted-foreground">
                      {importResult.warning
                        ? t("import_result_warning", {
                            imported: importResult.imported,
                            warning: importResult.warning,
                          })
                        : t("import_result_success", { imported: importResult.imported })}
                    </p>
                  )}
                </div>
                <div className="flex flex-col sm:flex-row gap-2 w-full sm:w-auto">
                  <Button onClick={handleManualImport} disabled={importing} className="w-full sm:w-auto">
                    {importing ? t("importing") : t("import_now")}
                  </Button>
                  <Button variant="outline" size="sm" onClick={fetchLogs} disabled={logsLoading} className="w-full sm:w-auto">
                    {logsLoading ? t("loading") : t("refresh")}
                  </Button>
                </div>
              </div>

              <div className="overflow-x-auto -mx-4 sm:mx-0">
                <table className="w-full text-sm min-w-[860px]">
                  <thead className="text-xs uppercase tracking-[0.2em] text-muted-foreground">
                    <tr>
                      <th className="py-2 text-left">{t("col_log_time")}</th>
                      <th className="py-2 text-left">{t("col_log_type")}</th>
                      <th className="py-2 text-left">{t("col_log_actor")}</th>
                      <th className="py-2 text-left">{t("col_log_message")}</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-black/5 dark:divide-white/10">
                    {logs.map((entry) => {
                      const metadataEntries = getLogMetadataEntries(entry.metadata);
                      return (
                        <tr key={entry.id}>
                          <td className="py-3 text-muted-foreground whitespace-nowrap">
                            {new Date(entry.timestamp).toLocaleString(locale)}
                          </td>
                          <td className="py-3 text-muted-foreground font-mono">{entry.type}</td>
                          <td className="py-3 text-muted-foreground">{entry.actor}</td>
                          <td className="py-3">
                            <div className="space-y-2">
                              <p className="text-muted-foreground">{entry.message}</p>
                              {metadataEntries.length > 0 && (
                                <div className="flex flex-wrap gap-2">
                                  {metadataEntries.map(([key, value]) => (
                                    <span
                                      key={`${entry.id}-${key}`}
                                      className="inline-flex items-center gap-1 rounded-full border border-black/10 dark:border-white/10 bg-black/[0.03] dark:bg-white/[0.04] px-2 py-1 text-xs"
                                    >
                                      <span className="text-muted-foreground">
                                        {logMetadataLabels[key] ?? key}
                                      </span>
                                      <span className="font-mono break-all text-foreground/80">
                                        {formatLogMetadataValue(key, value)}
                                      </span>
                                    </span>
                                  ))}
                                </div>
                              )}
                            </div>
                          </td>
                        </tr>
                      );
                    })}
                    {logs.length === 0 && !logsLoading && (
                      <tr>
                        <td colSpan={4} className="py-6 text-center text-muted-foreground">
                          {t("no_logs")}
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </section>
          </TabsContent>
        </Tabs>
      )}
    </div>
  );
}
