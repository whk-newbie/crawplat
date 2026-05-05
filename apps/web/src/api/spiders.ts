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

export type CreateSpiderInput = {
  projectId: string
  name: string
  language: 'go' | 'python'
  runtime: 'docker'
  image: string
  command: string[]
}

export type SpiderVersion = {
  id: string
  spiderId: string
  version: string
  image: string
  isCurrent: boolean
}

export type CreateVersionInput = {
  version: string
  image: string
  isCurrent: boolean
}

export type RegistryAuthRef = {
  id: string
  projectId: string
  name: string
  registryUrl: string
}

export function createSpider(input: CreateSpiderInput) {
  const { projectId, ...body } = input
  return apiFetch<Spider>(`/projects/${projectId}/spiders`, {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export function listSpiders(projectId: string) {
  return apiFetch<Spider[]>(`/projects/${projectId}/spiders`)
}

export function createSpiderVersion(spiderId: string, input: CreateVersionInput) {
  return apiFetch<SpiderVersion>(`/spiders/${spiderId}/versions`, {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function listSpiderVersions(spiderId: string) {
  return apiFetch<SpiderVersion[]>(`/spiders/${spiderId}/versions`)
}

export function listRegistryAuthRefs(projectId: string) {
  return apiFetch<RegistryAuthRef[]>(`/projects/${projectId}/registry-auth-refs`)
}
