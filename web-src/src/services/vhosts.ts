import { apiFetch } from '@/lib/apiClient'
import { VhostObj, Limits } from '@/types/vhosts'
import { VhostResources } from '@/types/clusterStats'

export interface ApiResponse<T> {
  code: number
  message: string
  body: T
}

export async function getVhosts(signal?: AbortSignal): Promise<string[]> {
  const res = await apiFetch('/api/v1/vhosts', { signal })
  if (!res.ok) {
    throw new Error(`Vhost request failed (${res.status})`)
  }
  const data: ApiResponse<VhostObj[]> = await res.json()
  return (data.body ?? []).map(v => v.name)
}

export async function getVhostUsage(vhost: string, signal?: AbortSignal): Promise<VhostResources> {
  const res = await apiFetch(`/api/v1/vhosts/${encodeURIComponent(vhost)}/usage`, { signal })
  if (!res.ok) {
    throw new Error(`Vhost usage request failed (${res.status})`)
  }
  const data: ApiResponse<VhostResources> = await res.json()
  return data.body
}


export function getSelectedVhost(vhosts: string[]): string {
  const params = new URLSearchParams(window.location.search)
  const vhost = params.get('vhost')
  return (vhost && vhosts.includes(vhost)) ? vhost : (vhosts[0] ?? '')
}

export async function getVhostLimits(vhost: string): Promise<Limits> {
  const res = await apiFetch(`/api/v1/vhosts/${encodeURIComponent(vhost)}/limits`)
  if (!res.ok) {
    throw new Error(`Vhost limits request failed (${res.status})`)
  }
  const data: ApiResponse<Limits> = await res.json()
  return data.body
}