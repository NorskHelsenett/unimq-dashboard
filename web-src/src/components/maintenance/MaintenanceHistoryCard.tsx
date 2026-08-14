import { Maintenance } from "@/types/maintenance"
import { upperCaseStatus } from "@/services/maintenance"
import { Pill } from "../ui/pill"

export function formatDateRange(start: string, end: string): string {
    const pad = (n: number) => String(n).padStart(2, '0')
    const fmtDate = (d: Date) =>
        `${d.getUTCFullYear()}-${pad(d.getUTCMonth() + 1)}-${pad(d.getUTCDate())} ${pad(d.getUTCHours())}:${pad(d.getUTCMinutes())}`
    const fmtTime = (d: Date) =>
        `${pad(d.getUTCHours())}:${pad(d.getUTCMinutes())}`

    const s = new Date(start)
    const e = new Date(end)
    const sameDay =
        s.getUTCFullYear() === e.getUTCFullYear() &&
        s.getUTCMonth() === e.getUTCMonth() &&
        s.getUTCDate() === e.getUTCDate()
    return sameDay ? `${fmtDate(s)} - ${fmtTime(e)}` : `${fmtDate(s)} - ${fmtDate(e)}`
}

export function MaintenanceHistoryCard({ maintenanceHistory }: { maintenanceHistory: Maintenance[] }) {
    const maintenanceHistorySorted = [...maintenanceHistory].sort((a, b) => new Date(b.start).getTime() - new Date(a.start).getTime())
    const maintenanceHistoryDataLength = maintenanceHistory.length
    if (maintenanceHistoryDataLength === 0) {
        return (
            <div className="bg-white rounded-lg shadow p-4 min-w-[300px] border border-border-card">
                <h3 className="text-lg font-semibold mb-2">Maintenance history</h3>
                <p className="text-gray-500">No maintenance history available.</p>
            </div>
        )
    }

    return (
        <div className="bg-white rounded-lg shadow p-4 min-w-[300px] border border-border-card">
            <h3 className="text-lg font-semibold mb-2">Maintenance history</h3>
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
                        {maintenanceHistorySorted.map((maintenance) => (
                            <tr key={maintenance.id}>
                                <td className="border-b border-border-card py-2 px-4">{maintenance.description}</td>
                                <td className="border-b border-border-card py-2 px-4">{formatDateRange(maintenance.start, maintenance.end)}</td>
                                <td className="border-b border-border-card py-2 px-4">
                                <Pill variant={maintenance.status === 'done' ? 'lightGreen' : maintenance.status === 'skipped' ? 'amber' : 'lightBlue'} className="border-none py-1 px-2 text-sm">
                                    {upperCaseStatus(maintenance.status)}
                                </Pill>
                            </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>  
        </div>

    )
}