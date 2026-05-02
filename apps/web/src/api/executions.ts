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
  executionStatus?: string
  executionTriggerSource?: string
  executionFrom?: string
  executionTo?: string
}) {
  const params = new URLSearchParams({
    projectId: input.projectId,
    limit: String(input.limit ?? 20),
    offset: String(input.offset ?? 0),
  })
  if (input.executionStatus?.trim()) {
    params.set('executionStatus', input.executionStatus.trim())
  }
  if (input.executionTriggerSource?.trim()) {
    params.set('executionTriggerSource', input.executionTriggerSource.trim())
  }
  if (input.spiderId?.trim()) {
    params.set('spiderId', input.spiderId.trim())
  }
  if (input.nodeId?.trim()) {
    params.set('nodeId', input.nodeId.trim())
  }
  if (input.executionFrom?.trim()) {
    params.set('executionFrom', input.executionFrom.trim())
  }
  if (input.executionTo?.trim()) {
    params.set('executionTo', input.executionTo.trim())
  }
  return apiFetch<PaginatedExecutions>(`/executions?${params.toString()}`)
}

export function getExecution(executionId: string) {
  return apiFetch<Execution>(`/executions/${executionId}`)
}

export function getExecutionLogs(executionId: string) {
  return apiFetch<ExecutionLog[]>(`/executions/${executionId}/logs`)
}
