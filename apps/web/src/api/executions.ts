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
  spiderVersion?: number
  registryAuthRef?: string
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

export type CreateExecutionInput = {
  projectId: string
  spiderId: string
  spiderVersion?: number
  registryAuthRef?: string
  image: string
  command: string[]
}

export type PaginatedExecutions = {
  items: Execution[]
  total: number
  limit: number
  offset: number
}

export function createExecution(input: CreateExecutionInput) {
  return apiFetch<Execution>('/executions', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function listExecutions(input: {
  projectId: string
  limit?: number
  offset?: number
  spiderId?: string
  nodeId?: string
  status?: string
  triggerSource?: string
  from?: string
  to?: string
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
}) {
  const params = new URLSearchParams({
    project_id: input.projectId,
    limit: String(input.limit ?? 20),
    offset: String(input.offset ?? 0),
    sort_by: input.sortBy ?? 'created_at',
    sort_order: input.sortOrder ?? 'desc',
  })
  if (input.status?.trim()) {
    params.set('status', input.status.trim())
  }
  if (input.triggerSource?.trim()) {
    params.set('trigger_source', input.triggerSource.trim())
  }
  if (input.spiderId?.trim()) {
    params.set('spider_id', input.spiderId.trim())
  }
  if (input.nodeId?.trim()) {
    params.set('node_id', input.nodeId.trim())
  }
  if (input.from?.trim()) {
    params.set('from', input.from.trim())
  }
  if (input.to?.trim()) {
    params.set('to', input.to.trim())
  }
  return apiFetch<PaginatedExecutions>(`/executions?${params.toString()}`)
}

export function getExecution(executionId: string) {
  return apiFetch<Execution>(`/executions/${executionId}`)
}

export function getExecutionLogs(executionId: string) {
  return apiFetch<ExecutionLog[]>(`/executions/${executionId}/logs`)
}
