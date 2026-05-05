import { apiFetch } from './client'

export type MonitorOverview = {
  generatedAt?: string
  timestamp?: string
  [key: string]: unknown
}

export function getMonitorOverview() {
  return apiFetch<MonitorOverview>('/monitor/overview')
}
