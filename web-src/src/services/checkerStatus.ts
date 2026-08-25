import { apiFetch } from '@/lib/apiClient'

export interface CheckerStatus {
  last_checked: string | null
  runtime_ms: number | null
  interval_s: number
}

export async function getCheckerStatus(): Promise<CheckerStatus | null> {
  const res = await apiFetch('/api/v1/status')
  if (!res.ok) return null
  const data = await res.json()
  return data.body ?? null
}
