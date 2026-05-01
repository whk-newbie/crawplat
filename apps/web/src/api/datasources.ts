import { apiFetch } from './client'

export type Datasource = {
  id: string
  projectId: string
  name: string
  type: 'mongodb' | 'redis' | 'postgresql' | string
  readonly: boolean
  config: Record<string, string>
}

export type CreateDatasourceInput = {
  projectId: string
  name: string
  type: 'mongodb' | 'redis' | 'postgresql'
  config?: Record<string, string>
}

export type DatasourceTestResult = {
  datasourceId: string
  status: string
  message: string
}

export type DatasourcePreviewResult = {
  datasourceId: string
  datasourceType: string
  rows: Array<Record<string, string>>
}

export type PaginatedDatasources = {
  items: Datasource[]
  total: number
  limit: number
  offset: number
}

function buildQueryString(input: Record<string, string>) {
  const params = new URLSearchParams()
  for (const [key, value] of Object.entries(input)) {
    if (value !== '') {
      params.set(key, value)
    }
  }
  const query = params.toString()
  return query ? `?${query}` : ''
}

export function listDatasources(projectId?: string) {
  const query = buildQueryString({
    projectId: projectId?.trim() ?? '',
  })
  return apiFetch<PaginatedDatasources>(`/datasources${query}`)
}

export function createDatasource(input: CreateDatasourceInput) {
  return apiFetch<Datasource>('/datasources', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function testDatasource(id: string) {
  return apiFetch<DatasourceTestResult>(`/datasources/${encodeURIComponent(id)}/test`, {
    method: 'POST',
  })
}

export function previewDatasource(id: string) {
  return apiFetch<DatasourcePreviewResult>(`/datasources/${encodeURIComponent(id)}/preview`, {
    method: 'POST',
  })
}
