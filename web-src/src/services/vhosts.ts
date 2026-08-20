import { apiFetch } from '@/lib/apiClient'
import { VhostObj } from '@/types/vhosts'

export interface ApiResponse<T> {
  code: number
  message: string
  body: T
}

export async function getVhosts(): Promise<string[]> {
  const res = await apiFetch('/api/v1/vhosts')
  const data: ApiResponse<VhostObj[]> = await res.json()
  return (data.body ?? []).map(v => v.name)
}


export function getSelectedVhost(vhosts: string[]): string {
  const params = new URLSearchParams(window.location.search)
  const vhost = params.get('vhost')
  return (vhost && vhosts.includes(vhost)) ? vhost : (vhosts[0] ?? '')
}