export interface ApiResponse<Scheduled, History> {
    code: number
    message: string
    body: {
        Scheduled: Scheduled[]
        History: History[]
    }
}

export interface Maintenance {
    id: string
    description: string
    start: string
    end: string
    status: 'scheduled' | 'in_progress' | 'done' | 'skipped' | 'unknown'
    notified: boolean
    updated_by?: string
    updated_at?: string
    update_reason?: string
}

export interface UseMaintenanceHistoryResult {
  maintenanceHistory: Maintenance[]
  loading: boolean
}

export interface UseMaintenanceScheduleResult {
  maintenanceSchedule: Maintenance[]
  loading: boolean
  refetch: () => void
}