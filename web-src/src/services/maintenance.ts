import { apiFetch } from '@/lib/apiClient'
import { ApiResponse, Maintenance } from '@/types/maintenance'

interface AdminApiResponse {
    code: number
    message: string
    body: {
        Entries: Maintenance[]
    }
}


export async function getScheduledMaintenance(): Promise<Maintenance[]> {
  const res = await apiFetch('/api/v1/maintenance')
  if (!res.ok) throw new Error('Failed to fetch maintenance data')
  const data: ApiResponse<Maintenance, Maintenance> = await res.json()
  return data.body.Scheduled ?? []
}

export async function getMaintenanceHistory(): Promise<Maintenance[]> {
    const res = await apiFetch('/api/v1/maintenance')
    if (!res.ok) throw new Error('Failed to fetch maintenance data')
    const data: ApiResponse<Maintenance, Maintenance> = await res.json()
    return data.body.History ?? []
}

export async function addMaintenance({
  description,
  start,
  end
}: {
  description: string,
  start: string,
  end: string
}): Promise<void> {
  await apiFetch(
    '/api/v1/maintenance',
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ description, start, end })
    }
  )
}

export interface MaintenanceEditLog {
    id: string
    maintenance_id: string
    description: string
    start: string
    end: string
    reason: string
    updated_by: string
    updated_at: string
}

export async function getMaintenanceEditLogs(maintenanceId: string): Promise<MaintenanceEditLog[]> {
    const res = await apiFetch(`/api/v1/maintenance/${maintenanceId}/logs`)
    if (!res.ok) throw new Error('Failed to fetch maintenance edit logs')
    const data: { code: number; message: string; body: { logs: MaintenanceEditLog[] } } = await res.json()
    return data.body.logs ?? []
}

export async function getMaintenanceAdmin(): Promise<Maintenance[]> {
    const res = await apiFetch('/api/v1/maintenance/admin')
    if (!res.ok) throw new Error('Failed to fetch maintenance data')
    const data: AdminApiResponse = await res.json()
    return data.body.Entries ?? []
}

export async function updateMaintenance({
    id,
    description,
    start,
    end,
    reason,
    updated_by,
}: {
    id: string
    description: string
    start: string
    end: string
    reason: string
    updated_by: string
}): Promise<void> {
    const res = await apiFetch(
        `/api/v1/maintenance/${id}`,
        {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ description, start, end, reason, updated_by }),
        }
    )
    if (!res.ok) {
        const text = await res.text().catch(() => '')
        throw new Error(text || 'Failed to update maintenance.')
    }
}

export async function deleteMaintenance(id: string): Promise<void> {
  await apiFetch(
    `/api/v1/maintenance/${id}`,
    {
      method: 'DELETE'
    }
  )
}


export async function updateMaintenanceStatus(id: string, status: 'scheduled' | 'in_progress' | 'done' | 'skipped'): Promise<void> {
    const res = await apiFetch(
        `/api/v1/maintenance/${id}`,
        {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ status }),
        }
    )
    if (!res.ok) {
        const text = await res.text().catch(() => '')
        throw new Error(text || 'Failed to update maintenance status.')
    }
}

export const upperCaseStatus = (status: string): string => {
    switch (status) {
        case 'scheduled': return 'Scheduled'
        case 'in_progress': return 'In Progress'
        case 'done':      return 'Done'
        case 'skipped':   return 'Skipped'
        default:          return status
    }
}