import { useState, useEffect } from 'react'
import { getMaintenanceHistory, getScheduledMaintenance, updateMaintenanceStatus as apiUpdateStatus } from '@/services/maintenance'
import type { Maintenance } from '@/types/maintenance'
import { UseMaintenanceHistoryResult, UseMaintenanceScheduleResult } from '@/types/maintenance'


export function useHistoricMaintenance(): UseMaintenanceHistoryResult {
  const [maintenanceHistory, setMaintenanceHistory] = useState<Maintenance[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    getMaintenanceHistory()
      .then(data => setMaintenanceHistory(data))
      .finally(() => setLoading(false))
  }, [])

  return { maintenanceHistory, loading }
}

export function useScheduledMaintenance(): UseMaintenanceScheduleResult {
  const [maintenanceSchedule, setMaintenanceSchedule] = useState<Maintenance[]>([])
  const [loading, setLoading] = useState(true)
  const [tick, setTick] = useState(0)

  useEffect(() => {
    setLoading(true)
    getScheduledMaintenance()
      .then(data => setMaintenanceSchedule(data))
      .finally(() => setLoading(false))
  }, [tick])

  return { maintenanceSchedule, loading, refetch: () => setTick(t => t + 1) }
}

export function updateMaintenanceStatus(maintenanceId: string, status: 'scheduled' | 'done' | 'skipped'): Promise<void> {
  return apiUpdateStatus(maintenanceId, status)
}