export function apiUrl(path: string) {
  const base = (process.env.NEXT_PUBLIC_API_URL || "").replace(/\/$/, "");
  return `${base}${path}`;
}

export function serverApiUrl(path: string) {
  const base = (
    process.env.AUTH_BACKEND_URL ||
    process.env.NEXT_PUBLIC_API_URL ||
    "http://backend:8080"
  ).replace(/\/$/, "");
  return `${base}${path}`;
}
