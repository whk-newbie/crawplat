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

export function listProjects() {
  return apiFetch<Project[]>('/projects')
}

export function createProject(input: CreateProjectInput) {
  return apiFetch<Project>('/projects', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}
