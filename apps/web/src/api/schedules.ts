import { apiFetch } from './client'

export type Schedule = {
  id: string
  projectId: string
  spiderId: string
  spiderVersion?: number
  name: string
  cronExpr: string
  enabled: boolean
  image: string
  command: string[]
  retryLimit: number
  retryDelaySeconds: number
}

export type CreateScheduleInput = {
  projectId: string
  spiderId: string
  spiderVersion?: number
  name: string
  cronExpr: string
  enabled: boolean
  image?: string
  command: string[]
  retryLimit: number
  retryDelaySeconds: number
}

export type PaginatedSchedules = {
  items: Schedule[]
  total: number
  limit: number
  offset: number
}

export function listSchedules() {
  return apiFetch<PaginatedSchedules>('/schedules')
}

export function createSchedule(input: CreateScheduleInput) {
  return apiFetch<Schedule>('/schedules', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}
