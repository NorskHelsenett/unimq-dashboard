import '../index.css'
import { createRoot } from 'react-dom/client'
import { StrictMode } from 'react'
import { RequireAuth } from '@/auth/RequireAuth'
import { Layout } from '@/components/layout/Layout'
import { MaintenanceHistoryCard } from '@/components/maintenance/MaintenanceHistoryCard'
import { useMaintenance } from '@/hooks/useMaintenance'

function MaintenancePage() {

  const { maintenanceHistory, loading } = useMaintenance()

  const { vhosts, selected } = { vhosts: [], selected: '' }
  return (
    <Layout Vhosts={vhosts} Selected={selected} MaintenanceMode={true}>
          {loading ? (
            <div className="p-8 text-text-muted">Loading...</div>
          ) : (
            <div>
              <h1 className='text-4xl mb-6'>Maintenance</h1>
              <MaintenanceHistoryCard maintenanceHistory={maintenanceHistory} />
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