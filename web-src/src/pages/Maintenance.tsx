import '../index.css'
import { createRoot } from 'react-dom/client'
import { StrictMode } from 'react'
import { RequireAuth } from '@/auth/RequireAuth'
import { Layout } from '@/components/layout/Layout'
import { MaintenanceHistoryCard } from '@/components/maintenance/MaintenanceHistoryCard'
import { useHistoricMaintenance, useScheduledMaintenance } from '@/hooks/useMaintenance'
import { useIndex } from '@/hooks/useIndex'
import { MaintenanceScheduleCard } from '@/components/maintenance/MaintenanceScheduleCard'

function MaintenancePage() {

  const { maintenanceHistory, loading } = useHistoricMaintenance()
  const { maintenanceSchedule, loading: loadingSchedule, refetch } = useScheduledMaintenance()

  const { Vhosts, Selected } = useIndex()
  return (
    <Layout Vhosts={Vhosts} Selected={Selected} >
          {loading || loadingSchedule ? (
            <div className="p-8 text-text-muted">Loading...</div>
          ) : (
            <div>
              <h1 className='text-4xl mb-6'>Maintenance</h1>
              <div className='max-w-4xl mx-auto flex flex-col gap-4'>
                <MaintenanceScheduleCard maintenanceSchedule={maintenanceSchedule} onRefresh={refetch} />
                <MaintenanceHistoryCard maintenanceHistory={maintenanceHistory} />
              </div>
            </div>
          )}
    </Layout>
  )
}

const root = document.getElementById('app')
if (!root) throw new Error('Missing #app mount point')

createRoot(root).render(
  <StrictMode>
    <RequireAuth>
      <MaintenancePage />
    </RequireAuth>
  </StrictMode>,
)