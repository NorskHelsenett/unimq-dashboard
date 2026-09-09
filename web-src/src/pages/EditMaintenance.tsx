import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '../index.css'
import { RequireAuth } from '@/auth/RequireAuth'
import { Layout } from '@/components/layout/Layout'
import { EditMaintenance } from '@/components/maintenance/EditMaintenance'
import { useMaintenanceById } from '@/hooks/useMaintenance'

function EditMaintenancePage() {
    const id = new URLSearchParams(window.location.search).get('id')
    const { maintenance, loading } = useMaintenanceById(id ?? '')

    return (
        <Layout>
            <div className="max-w-4xl mx-auto">
                <a href="/maintenance" className="text-sm text-text-muted hover:text-text-primary mb-6 inline-block">
                    ← Back to maintenance
                </a>
                {loading ? (
                    <div className="p-8 text-text-muted">Loading…</div>
                ) : !maintenance ? (
                    <p className="text-sm text-text-muted">Maintenance entry not found.</p>
                ) : (
                    <EditMaintenance maintenance={maintenance} />
                )}
            </div>
        </Layout>
    )
}

const root = document.getElementById('app')
if (!root) throw new Error('Missing #app mount point')

createRoot(root).render(
    <StrictMode>
        <RequireAuth>
            <EditMaintenancePage />
        </RequireAuth>
    </StrictMode>,
)
