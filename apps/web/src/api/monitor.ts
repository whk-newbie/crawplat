import { apiFetch } from './client'

export type MonitorOverview = {
  executions?: { total: number; pending: number; running: number; succeeded: number; failed: number }
  nodes?: { total: number; online: number; offline: number }
  generatedAt?: string
  timestamp?: string
  [key: string]: unknown
}

export type AlertRule = {
  id: string
  name: string
  ruleType: 'execution_failed' | 'node_offline'
  enabled: boolean
  webhookUrl: string
  cooldownSeconds: number
  timeoutSeconds: number
  offlineGraceSeconds: number
  createdAt: string
  updatedAt: string
}

export type AlertEvent = {
  id: string
  ruleId: string
  ruleType: 'execution_failed' | 'node_offline'
  entityType: string
  entityId: string
  dedupeKey: string
  payload: string
  deliveryStatus: string
  webhookStatusCode?: number
  errorMessage?: string
  createdAt: string
}

export type Paginated<T> = {
  items: T[]
  total: number
  limit: number
  offset: number
}

export function getMonitorOverview() {
  return apiFetch<MonitorOverview>('/monitor/overview')
}

export function listAlertRules() {
  return apiFetch<AlertRule[]>('/monitor/alerts/rules')
}

export function createAlertRule(input: {
  name: string
  ruleType: 'execution_failed' | 'node_offline'
  webhookUrl: string
  enabled?: boolean
  cooldownSeconds?: number
  timeoutSeconds?: number
  offlineGraceSeconds?: number
}) {
  return apiFetch<AlertRule>('/monitor/alerts/rules', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function updateAlertRule(id: string, patch: {
  name?: string
  enabled?: boolean
  webhookUrl?: string
  cooldownSeconds?: number
  timeoutSeconds?: number
  offlineGraceSeconds?: number
}) {
  return apiFetch<AlertRule>(`/monitor/alerts/rules/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    body: JSON.stringify(patch),
  })
}

export function listAlertEvents(limit = 20, offset = 0) {
  return apiFetch<Paginated<AlertEvent>>(`/monitor/alerts/events?limit=${limit}&offset=${offset}`)
}
