import { apiFetch } from './client'

export type NodeSummary = {
  id: string
  name: string
  status: string
  capabilities: string[]
  lastSeenAt: string
}

export type NodeHeartbeat = {
  seenAt: string
  status?: string
  capabilities?: string[]
}

export type NodeRecentExecution = {
  id: string
  spiderId?: string
  status: string
  startedAt?: string
  finishedAt?: string
}

export type NodeDetail = NodeSummary & {
  heartbeats: NodeHeartbeat[]
  recentExecutions: NodeRecentExecution[]
}

export type NodeSession = {
  startedAt: string
  endedAt?: string
  durationSeconds?: number
  heartbeatCount?: number
}

export type NodeSessionsSummary = {
  totalSessions: number
  totalHeartbeatCount: number
  totalOnlineDurationSeconds: number
}

export type NodeDetailQuery = {
  executionLimit?: number
  executionOffset?: number
  executionStatus?: string
  executionFrom?: string
  executionTo?: string
}

export type NodeSessionsQuery = {
  limit?: number
  gapSeconds?: number
}

export function listNodes() {
  return apiFetch<NodeSummary[]>('/nodes')
}

type NodeDetailResponse = {
  node: NodeSummary
  heartbeatHistory: Array<{
    seenAt: string
    status?: string
    capabilities?: string[]
  }>
  recentExecutions: NodeRecentExecution[]
}

type NodeSessionsResponse =
  | NodeSession[]
  | {
      sessions?: NodeSession[]
      summary?: Partial<NodeSessionsSummary> & {
        sessionCount?: number
        totalDurationSeconds?: number
      }
    }

export type NodeSessionsResult = {
  sessions: NodeSession[]
  summary: NodeSessionsSummary
}

function buildFallbackDetail(node: NodeSummary): NodeDetail {
  return {
    ...node,
    heartbeats: node.lastSeenAt ? [{ seenAt: node.lastSeenAt, capabilities: node.capabilities }] : [],
    recentExecutions: [],
  }
}

function shouldFallbackByError(err: unknown) {
  const message = err instanceof Error ? err.message : ''
  return /404|not found/i.test(message)
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

export async function getNodeDetail(nodeId: string, query: NodeDetailQuery = {}) {
  const requestQuery = buildQueryString({
    executionLimit: query.executionLimit ? String(query.executionLimit) : '',
    executionOffset: query.executionOffset ? String(query.executionOffset) : '',
    executionStatus: query.executionStatus ?? '',
    executionFrom: query.executionFrom ?? '',
    executionTo: query.executionTo ?? '',
  })

  try {
    const detail = await apiFetch<NodeDetailResponse>(`/nodes/${encodeURIComponent(nodeId)}${requestQuery}`)
    return {
      ...detail.node,
      heartbeats: detail.heartbeatHistory,
      recentExecutions: detail.recentExecutions,
    }
  } catch (err) {
    if (!shouldFallbackByError(err)) {
      throw err
    }

    const nodes = await listNodes()
    const node = nodes.find((item) => item.id === nodeId)
    if (!node) {
      throw err
    }
    return buildFallbackDetail(node)
  }
}

export async function getNodeSessions(nodeId: string, query: NodeSessionsQuery = {}) {
  const requestQuery = buildQueryString({
    limit: query.limit ? String(query.limit) : '',
    gapSeconds: query.gapSeconds ? String(query.gapSeconds) : '',
  })

  try {
    const response = await apiFetch<NodeSessionsResponse>(
      `/nodes/${encodeURIComponent(nodeId)}/sessions${requestQuery}`,
    )
    if (Array.isArray(response)) {
      return {
        sessions: response,
        summary: {
          totalSessions: response.length,
          totalHeartbeatCount: 0,
          totalOnlineDurationSeconds: 0,
        },
      }
    }
    return {
      sessions: response.sessions ?? [],
      summary: {
        totalSessions: response.summary?.totalSessions ?? response.summary?.sessionCount ?? 0,
        totalHeartbeatCount: response.summary?.totalHeartbeatCount ?? 0,
        totalOnlineDurationSeconds: response.summary?.totalOnlineDurationSeconds ?? response.summary?.totalDurationSeconds ?? 0,
      },
    }
  } catch (err) {
    if (shouldFallbackByError(err)) {
      return {
        sessions: [],
        summary: {
          totalSessions: 0,
          totalHeartbeatCount: 0,
          totalOnlineDurationSeconds: 0,
        },
      }
    }
    throw err
  }
}
