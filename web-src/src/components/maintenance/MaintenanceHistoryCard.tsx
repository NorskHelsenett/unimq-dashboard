import { Maintenance } from "@/types/maintenance"

function formatDateRange(start: string, end: string): string {
    // Expected data format: "2024-06-01 10:00:00"
    const fmt = (d: Date) =>
        `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
    const fmtTime = (d: Date) =>
        `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`

    const s = new Date(start.replace(' ', 'T'))
    const e = new Date(end.replace(' ', 'T'))
    const sameDay = s.getFullYear() === e.getFullYear() && s.getMonth() === e.getMonth() && s.getDate() === e.getDate()
    return sameDay ? `${fmt(s)} - ${fmtTime(e)}` : `${fmt(s)} - ${fmt(e)}`
}

export function MaintenanceHistoryCard({ maintenanceHistory }: { maintenanceHistory: Maintenance[] }) {
    
    return (
        <div className="bg-white rounded-lg shadow p-4 min-w-[300px] border border-border-card">
            <h3 className="text-lg font-semibold mb-2">Maintenance History</h3>
            <div className="overflow-y-auto max-h-64">
                <table className="w-full text-left border-collapse">
                    <thead>
                        <tr>
                            <th className="border-b border-border-card py-2 px-4 text-gray-500">Description</th>
                            <th className="border-b border-border-card py-2 px-4 text-gray-500">Date</th>
                            <th className="border-b border-border-card py-2 px-4 text-gray-500">Status</th>
                        </tr>
                    </thead>
                    <tbody>
                        {maintenanceHistory.map((maintenance) => (
                            <tr key={maintenance.id}>
                                <td className="border-b border-border-card py-2 px-4">{maintenance.description}</td>
                                <td className="border-b border-border-card py-2 px-4">{formatDateRange(maintenance.start, maintenance.end)}</td>
                                <td className="border-b border-border-card py-2 px-4">{maintenance.status}</td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>  
        </div>

    )
}