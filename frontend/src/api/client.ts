// frontend/src/api/client.ts
// Shared API client with Bearer token support.
// Fixes "UNAUTHORIZED: Missing token" by always attaching Authorization header when token exists.

export class ApiError extends Error {
  status: number;
  code?: string;
  constructor(message: string, status: number, code?: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

const TOKEN_KEY = "firegoals_token";

export function storeToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token);
}

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY);
}

export function getApiBaseUrl(): string {
  const fromEnv = (import.meta as any).env?.VITE_API_BASE_URL as string | undefined;
  return (fromEnv ?? "").replace(/\/$/, "");
}

export function hasApiBaseUrl(): boolean {
  return getApiBaseUrl().length > 0;
}

export async function apiFetch<T = any>(path: string, init: RequestInit = {}): Promise<T> {
  const base = getApiBaseUrl();
  const url = path.startsWith("http") ? path : `${base}${path.startsWith("/") ? "" : "/"}${path}`;

  const headers = new Headers(init.headers ?? {});
  if (!headers.has("Content-Type") && init.body != null) {
    headers.set("Content-Type", "application/json");
  }

  // Attach Bearer token automatically if present.
  const token = getToken();
  if (token && !headers.has("Authorization")) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const res = await fetch(url, { ...init, headers });

  // 204 No Content
  if (res.status === 204) return undefined as unknown as T;

  const contentType = res.headers.get("content-type") ?? "";
  const isJSON = contentType.includes("application/json");

  if (!res.ok) {
    let message = `${res.status} ${res.statusText}`;
    let code: string | undefined;

    try {
      if (isJSON) {
        const errBody = await res.json();
        const e = errBody?.error;
        if (e?.message) message = e.message;
        if (e?.code) code = e.code;
      } else {
        const txt = await res.text();
        if (txt) message = txt;
      }
    } catch {
      // ignore parse errors
    }

    throw new ApiError(message, res.status, code);
  }

  if (!isJSON) {
    return (await res.text()) as unknown as T;
  }
  return (await res.json()) as T;
}
