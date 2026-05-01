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

export type NodeDetailQuery = {
  executionLimit?: number
  executionOffset?: number
  executionStatus?: string
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
      return response
    }
    return response.sessions ?? []
  } catch (err) {
    if (shouldFallbackByError(err)) {
      return []
    }
    throw err
  }
}
