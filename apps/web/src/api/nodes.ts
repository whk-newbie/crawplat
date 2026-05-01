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

export function listNodes() {
  return apiFetch<NodeSummary[]>('/nodes')
}

type NodeDetailResponse = {
  node: NodeSummary
  heartbeatHistory: Array<{
    seenAt: string
    capabilities?: string[]
  }>
  recentExecutions: NodeRecentExecution[]
}

function buildFallbackDetail(node: NodeSummary): NodeDetail {
  return {
    ...node,
    heartbeats: node.lastSeenAt ? [{ seenAt: node.lastSeenAt, capabilities: node.capabilities }] : [],
    recentExecutions: [],
  }
}

export async function getNodeDetail(nodeId: string) {
  try {
    const detail = await apiFetch<NodeDetailResponse>(`/nodes/${encodeURIComponent(nodeId)}`)
    return {
      ...detail.node,
      heartbeats: detail.heartbeatHistory,
      recentExecutions: detail.recentExecutions,
    }
  } catch (err) {
    const message = err instanceof Error ? err.message : ''
    const shouldFallback = /404|not found/i.test(message)
    if (!shouldFallback) {
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
