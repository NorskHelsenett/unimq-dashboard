import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '../index.css'
import { RequireAuth } from '@/auth/RequireAuth'
import { Layout } from '@/components/layout/Layout'
import { EditMaintenance } from '@/components/maintenance/EditMaintenance'
import { useEffect, useState } from 'react'
import { getMaintenanceAdmin } from '@/services/maintenance'
import { Maintenance } from '@/types/maintenance'

function EditMaintenancePage() {
    const id = new URLSearchParams(window.location.search).get('id')
    const [maintenance, setMaintenance] = useState<Maintenance | null>(null)
    const [loading, setLoading] = useState(true)
    const [notFound, setNotFound] = useState(false)

    useEffect(() => {
        if (!id) {
            setNotFound(true)
            setLoading(false)
            return
        }
        getMaintenanceAdmin()
            .then(entries => {
                const found = entries.find(e => e.id === id) ?? null
                setMaintenance(found)
                if (!found) setNotFound(true)
            })
            .catch(() => setNotFound(true))
            .finally(() => setLoading(false))
    }, [id])

    return (
        <Layout>
            <div className="max-w-4xl mx-auto">
                <a href="/maintenance" className="text-sm text-text-muted hover:text-text-primary mb-6 inline-block">
                    ← Back to maintenance
                </a>
                {loading ? (
                    <div className="p-8 text-text-muted">Loading…</div>
                ) : notFound || !maintenance ? (
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
