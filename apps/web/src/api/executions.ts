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

export function listExecutions(input: { projectId: string; limit?: number; offset?: number }) {
  const limit = input.limit ?? 20
  const offset = input.offset ?? 0
  return apiFetch<PaginatedExecutions>(
    `/executions?projectId=${encodeURIComponent(input.projectId)}&limit=${limit}&offset=${offset}`
  )
}

export function getExecution(executionId: string) {
  return apiFetch<Execution>(`/executions/${executionId}`)
}

export function getExecutionLogs(executionId: string) {
  return apiFetch<ExecutionLog[]>(`/executions/${executionId}/logs`)
}
