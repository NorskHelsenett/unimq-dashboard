// import { createRoot } from 'react-dom/client'
// import '../index.css'

// function App() {
//   return <div />
// }

// const root = document.getElementById('app')
// if (!root) throw new Error('Missing #app mount point')
// createRoot(root).render(<App />)

import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '../index.css'
import { getPageData } from '@/lib/pageData'
import { Layout } from '@/components/layout/Layout'
import { ScheduledSection, HistoryTable } from '@/components/maintenance/MaintenanceCard'
import type { MaintenanceEntry } from '@/types/maintenance'
import { RequireAuth } from '@/auth/RequireAuth'

interface MaintenanceData {
    Vhosts: string[]
    Selected: string
    Scheduled: MaintenanceEntry[]
    History: MaintenanceEntry[]
}

const data = getPageData<MaintenanceData>()

const MaintenancePage = () => {
    return (
        <div>
            <h1 className="text-4xl mb-6">Maintenance</h1>
            <div className="max-w-4xl mx-auto flex flex-col gap-6">
                <div className="bg-white rounded-lg shadow p-6 border border-border-card">
                    <div className="flex items-center justify-between mb-4">
                        <h2 className="text-lg font-semibold">Scheduled</h2>
                        <a
                            href="/maintenance/admin"
                            className="text-sm text-gray-500 hover:text-gray-900 transition-colors"
                        >
                            Manage →
                        </a>
                    </div>
                    <ScheduledSection entries={data.Scheduled || []} />
                </div>

                <div className="bg-white rounded-lg shadow p-6 border border-border-card">
                    <h2 className="text-lg font-semibold mb-4">History</h2>
                    <HistoryTable entries={data.History || []} />
                </div>
            </div>
        </div>
    )
}

const root = document.getElementById('app')
if (!root) throw new Error('Missing #app mount point')

createRoot(root).render(
    <StrictMode>
        <Layout Vhosts={data.Vhosts} Selected={data.Selected}>
            <MaintenancePage />
        </Layout>
    </StrictMode>,
)