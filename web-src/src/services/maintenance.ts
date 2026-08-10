import { apiFetch } from '@/lib/apiClient'
import { ApiResponse, Maintenance } from '@/types/maintenance'


export async function getScheduledMaintenance(): Promise<Maintenance[]> {
  const res = await apiFetch('/api/v1/maintenance')
  if (!res.ok) throw new Error('Failed to fetch maintenance data')
  const data: ApiResponse<Maintenance, Maintenance> = await res.json()
  return data.body.scheduled ?? []
}

export async function getMaintenanceHistory(): Promise<Maintenance[]> {
    const res = await apiFetch('/api/v1/maintenance')
    if (!res.ok) throw new Error('Failed to fetch maintenance data')
    const data: ApiResponse<Maintenance, Maintenance> = await res.json()
    return data.body.history ?? []
}
