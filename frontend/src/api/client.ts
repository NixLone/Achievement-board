export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL as string | undefined;

export class ApiError extends Error {
  constructor(
    public code: string,
    message: string,
    public status?: number
  ) {
    super(message);
    this.name = "ApiError";
  }
}

function getToken(): string | null {
  return localStorage.getItem("firegoals-token");
}

function setToken(token: string) {
  localStorage.setItem("firegoals-token", token);
}

export function hasApiBaseUrl() {
  return Boolean(API_BASE_URL);
}

export async function apiFetch<T>(path: string, options: RequestInit = {}): Promise<T> {
  if (!API_BASE_URL) {
    throw new ApiError("API_BASE_URL_MISSING", "API URL не настроен");
  }
  const token = getToken();
  const headers = new Headers(options.headers ?? {});
  headers.set("Content-Type", "application/json");
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }
  const response = await fetch(`${API_BASE_URL}${path}`, { ...options, headers });
  const contentType = response.headers.get("Content-Type") ?? "";
  if (!response.ok) {
    if (contentType.includes("application/json")) {
      const payload = (await response.json()) as { error?: { code?: string; message?: string } };
      const code = payload?.error?.code ?? "API_ERROR";
      const message = payload?.error?.message ?? "Ошибка запроса";
      throw new ApiError(code, message, response.status);
    }
    throw new ApiError("API_ERROR", "Сервер недоступен", response.status);
  }
  if (contentType.includes("application/json")) {
    return response.json() as Promise<T>;
  }
  return {} as T;
}

export function storeToken(token: string) {
  setToken(token);
}
