import { apiFetch } from './client'

export type LoginInput = {
  username: string
  password: string
}

export type RegisterInput = {
  username: string
  password: string
}

export type AuthResponse = {
  token: string
}

export function login(input: LoginInput) {
  return apiFetch<AuthResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function register(input: RegisterInput) {
  return apiFetch<AuthResponse>('/auth/register', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}
