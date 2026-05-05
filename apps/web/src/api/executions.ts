import { apiFetch } from './client'

export type ExecutionLog = {
  id: string
  executionId: string
  message: string
  createdAt: string
}

export type Execution = {
  id: string
  projectId: string
  spiderId: string
  nodeId?: string
  status: string
  triggerSource: string
  image: string
  command: string[]
  errorMessage?: string
  createdAt: string
  startedAt?: string
  finishedAt?: string
  logs?: ExecutionLog[]
}

export type PaginatedExecutions = {
  items: Execution[]
  total: number
  limit: number
  offset: number
}

export type ListExecutionsParams = {
  limit?: number
  offset?: number
  status?: string
}

export type CreateExecutionInput = {
  projectId: string
  spiderId: string
  image: string
  command: string[]
}

export function listExecutions(params?: ListExecutionsParams) {
  const qs = new URLSearchParams()
  if (params?.limit != null) qs.set('limit', String(params.limit))
  if (params?.offset != null) qs.set('offset', String(params.offset))
  if (params?.status) qs.set('status', params.status)
  const query = qs.toString()
  return apiFetch<PaginatedExecutions>(`/executions${query ? '?' + query : ''}`)
}

export function createExecution(input: CreateExecutionInput) {
  return apiFetch<Execution>('/executions', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function getExecution(executionId: string) {
  return apiFetch<Execution>(`/executions/${executionId}`)
}

export function getExecutionLogs(executionId: string) {
  return apiFetch<ExecutionLog[]>(`/executions/${executionId}/logs`)
}
