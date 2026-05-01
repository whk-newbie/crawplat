import { loadPersistedToken } from '../stores/auth'

const defaultBaseURL = '/api/v1'

function resolveBaseURL() {
  const envValue = import.meta.env.VITE_API_BASE_URL as string | undefined
  return (envValue && envValue.trim()) || defaultBaseURL
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers)
  if (!headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }
  const token = loadPersistedToken()
  if (token && !headers.has('Authorization')) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const response = await fetch(`${resolveBaseURL()}${path}`, {
    ...init,
    headers,
  })

  if (!response.ok) {
    const message = await response.text()
    throw new Error(message || `request failed: ${response.status}`)
  }

  if (response.status === 204) {
    return undefined as T
  }

  return response.json() as Promise<T>
}
