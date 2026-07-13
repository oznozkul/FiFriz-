import type { ApiResponse } from "../types/document";

const configuredBaseUrl = import.meta.env.VITE_API_BASE_URL;
// Empty string = same-origin (/api via Vite proxy in dev or nginx in Docker/production).
export const apiBaseUrl =
  configuredBaseUrl !== undefined ? String(configuredBaseUrl) : "";

const tokenStorageKey = "fifriz_access_token";

export function getAccessToken() {
  return localStorage.getItem(tokenStorageKey);
}

export function setAccessToken(token: string) {
  localStorage.setItem(tokenStorageKey, token);
}

export function clearAccessToken() {
  localStorage.removeItem(tokenStorageKey);
}

function handleUnauthorized() {
  clearAccessToken();
  window.dispatchEvent(new CustomEvent("auth:logout"));
  const path = window.location.pathname;
  if (!path.startsWith("/login") && !path.startsWith("/register") && !path.startsWith("/share/") && !path.startsWith("/forgot-password") && !path.startsWith("/reset-password")) {
    window.location.assign("/login");
  }
}

export async function apiRequest<T>(
  path: string,
  options: RequestInit = {},
  fallbackMessage = "API isteği başarısız oldu.",
): Promise<T> {
  const headers = new Headers(options.headers);
  if (!headers.has("Content-Type") && options.body && !(options.body instanceof FormData)) {
    headers.set("Content-Type", "application/json");
  }

  const token = getAccessToken();
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(`${apiBaseUrl}${path}`, {
    ...options,
    headers,
  });

  if (response.status === 401) {
    handleUnauthorized();
    throw new Error("Oturum süresi doldu. Lütfen tekrar giriş yapın.");
  }

  if (response.status === 204) {
    return undefined as T;
  }

  const payload = (await response.json().catch(() => null)) as ApiResponse<T> | null;
  if (!response.ok || payload?.success === false || !payload) {
    if (response.status === 403) {
      throw new Error(payload?.error?.message ?? "Bu işlem için yetkiniz yok.");
    }
    throw new Error(payload?.error?.message ?? fallbackMessage);
  }

  return payload.data as T;
}

export function apiGet<T>(path: string, fallbackMessage?: string) {
  return apiRequest<T>(path, { method: "GET" }, fallbackMessage);
}

export function apiPost<T>(path: string, body?: unknown, fallbackMessage?: string) {
  return apiRequest<T>(
    path,
    {
      method: "POST",
      body: body === undefined ? undefined : JSON.stringify(body),
    },
    fallbackMessage,
  );
}

export function apiPut<T>(path: string, body?: unknown, fallbackMessage?: string) {
  return apiRequest<T>(
    path,
    {
      method: "PUT",
      body: body === undefined ? undefined : JSON.stringify(body),
    },
    fallbackMessage,
  );
}

export function apiPatch<T>(path: string, body?: unknown, fallbackMessage?: string) {
  return apiRequest<T>(
    path,
    {
      method: "PATCH",
      body: body === undefined ? undefined : JSON.stringify(body),
    },
    fallbackMessage,
  );
}

export function apiDelete(path: string, fallbackMessage?: string) {
  return apiRequest<void>(path, { method: "DELETE" }, fallbackMessage);
}
