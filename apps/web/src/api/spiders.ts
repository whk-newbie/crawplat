import { apiFetch } from './client'

export type Spider = {
  id: string
  projectId: string
  name: string
  language: 'go' | 'python'
  runtime: 'docker' | 'host'
  image?: string
  command?: string[]
}

export type SpiderVersion = {
  id: string
  spiderId: string
  version: number
  registryAuthRef?: string
  image: string
  command: string[]
  createdAt: string
}

export type CreateSpiderInput = {
  projectId: string
  name: string
  language: 'go' | 'python'
  runtime: 'docker'
  image: string
  command: string[]
}

export type PaginatedSpiders = {
  items: Spider[]
  total: number
  limit: number
  offset: number
}

export function createSpider(input: CreateSpiderInput) {
  const { projectId, ...body } = input
  return apiFetch<Spider>(`/projects/${projectId}/spiders`, {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export function listSpiders(projectId: string) {
  return apiFetch<PaginatedSpiders>(`/projects/${projectId}/spiders`)
}

export function listSpiderVersions(spiderId: string) {
  return apiFetch<SpiderVersion[]>(`/spiders/${encodeURIComponent(spiderId)}/versions`)
}

export function createSpiderVersion(input: { spiderId: string; registryAuthRef?: string; image: string; command: string[] }) {
  const { spiderId, ...body } = input
  return apiFetch<SpiderVersion>(`/spiders/${encodeURIComponent(spiderId)}/versions`, {
    method: 'POST',
    body: JSON.stringify(body),
  })
}
