import { pinia } from '../main'
import { useAuthStore } from '../stores/auth'

const defaultBaseURL = '/api/v1'

function resolveBaseURL() {
  const envValue = import.meta.env.VITE_API_BASE_URL as string | undefined
  return (envValue && envValue.trim()) || defaultBaseURL
}

function buildHeaders(init?: RequestInit): Record<string, string> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(init?.headers as Record<string, string> ?? {}),
  }
  const authStore = useAuthStore(pinia)
  if (authStore.token) {
    headers['Authorization'] = `Bearer ${authStore.token}`
  }
  return headers
}

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

const defaultErrorCodes: Record<number, string> = {
  400: 'errors.gateway.badRequest',
  401: 'errors.gateway.missingBearerToken',
  403: 'errors.gateway.invalidBearerToken',
  404: 'errors.gateway.notFound',
  429: 'errors.gateway.rateLimitExceeded',
  500: 'errors.gateway.upstreamServiceUnavailable',
  502: 'errors.gateway.upstreamServiceUnavailable',
  503: 'errors.gateway.upstreamServiceUnavailable',
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${resolveBaseURL()}${path}`, {
    ...init,
    headers: buildHeaders(init),
  })

  if (!response.ok) {
    let body: Record<string, unknown> = {}
    try {
      body = await response.json()
    } catch {
      // non-JSON response body, use status text
    }

    const message =
      (body.message as string) || `request failed: ${response.status}`
    const code =
      (body.code as string) ||
      (defaultErrorCodes[response.status] ?? 'errors.fallback')

    throw new ApiError(response.status, code, message)
  }

  if (response.status === 204) {
    return undefined as T
  }

  return response.json() as Promise<T>
}
