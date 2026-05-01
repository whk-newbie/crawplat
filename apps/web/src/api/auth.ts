import { apiFetch } from './client'

export type LoginInput = {
  username: string
  password: string
}

export type LoginResponse = {
  token: string
}

export function login(input: LoginInput) {
  return apiFetch<LoginResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}
