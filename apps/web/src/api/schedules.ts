import { apiFetch } from './client'

export type Schedule = {
  id: string
  projectId: string
  spiderId: string
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
  name: string
  cronExpr: string
  enabled: boolean
  image: string
  command: string[]
  retryLimit: number
  retryDelaySeconds: number
}

export function listSchedules() {
  return apiFetch<Schedule[]>('/schedules')
}

export function createSchedule(input: CreateScheduleInput) {
  return apiFetch<Schedule>('/schedules', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}
