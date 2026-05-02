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
