import { apiFetch } from './client'

export type Project = {
  id: string
  code: string
  name: string
}

export type CreateProjectInput = {
  code: string
  name: string
}

export type PaginatedProjects = {
  items: Project[]
  total: number
  limit: number
  offset: number
}

export function listProjects() {
  return apiFetch<PaginatedProjects>('/projects')
}

export function createProject(input: CreateProjectInput) {
  return apiFetch<Project>('/projects', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}
