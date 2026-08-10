import { useState, useEffect } from 'react'
import { getMaintenanceHistory } from '@/services/maintenance'
import type { Maintenance } from '@/types/maintenance'

interface UseMaintenanceResult {
  maintenanceHistory: Maintenance[]
  loading: boolean
}

export function useMaintenance(): UseMaintenanceResult {
  const [maintenanceHistory, setMaintenanceHistory] = useState<Maintenance[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    getMaintenanceHistory()
      .then(data => setMaintenanceHistory(data))
      .finally(() => setLoading(false))
  }, [])

  return { maintenanceHistory, loading }
}
