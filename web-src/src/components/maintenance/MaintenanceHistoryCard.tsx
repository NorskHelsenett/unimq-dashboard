import { Maintenance } from "@/types/maintenance"
import { upperCaseStatus } from "@/services/maintenance"
import { Pill } from "../ui/pill"
import { ChevronDown, ChevronUp, History } from "lucide-react"
import { useLocalStorage } from "@/hooks/useLocalStorage"
import { SectionCard, SectionCardHeader } from "../ui/section-card"

export function formatDateRange(start: string, end: string): string {
    const formatDateTime = new Intl.DateTimeFormat('no-NO', {
        timeZone: 'Europe/Oslo',
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        hourCycle: 'h23',
    })
    const formatTime = new Intl.DateTimeFormat('no-NO', {
        timeZone: 'Europe/Oslo',
        hour: '2-digit',
        minute: '2-digit',
        hourCycle: 'h23',
    })

    const s = new Date(start)
    const e = new Date(end)
    const sDate = formatDateTime.format(s)
    const eDate = formatDateTime.format(e)
    const sameDay = sDate.slice(0, 10) === eDate.slice(0, 10)
    return sameDay ? `${sDate} - ${formatTime.format(e)}` : `${sDate} - ${eDate}`
}

export const durationInMinutes = (start: string, end: string) => {
    const startDate = new Date(start)
    const endDate = new Date(end)
    return Math.round((endDate.getTime() - startDate.getTime()) / (1000 * 60))
}

export function MaintenanceHistoryCard({ maintenanceHistory }: { maintenanceHistory: Maintenance[] }) {
    const maintenanceHistorySorted = [...maintenanceHistory].sort((a, b) => new Date(b.start).getTime() - new Date(a.start).getTime())
    const [open, setOpen] = useLocalStorage("maintenance-history-open", true)

    return (
        <SectionCard accent="green">
        <button
                onClick={() => setOpen(o => !o)}
                className="w-full text-left"
            >
            <SectionCardHeader
                title="Maintenance history"
                icon={<History className="w-4 h-4 text-green-400" />}
                action={open ? <ChevronUp className="w-4 h-4 text-text-muted" /> : <ChevronDown className="w-4 h-4 text-text-muted" />}
                />
            </button>
            {open && (
                maintenanceHistory.length === 0 ? (
                    <p className="text-text-muted">No maintenance history available.</p>
                ) : (
                    <div className="overflow-y-auto max-h-64 text-sm">
                        <table className="w-full text-left border-collapse">
                            <thead>
                                <tr>
                                    <th className="border-b border-border-card py-2 px-4 text-xs text-text-muted">Description</th>
                                    <th className="border-b border-border-card py-2 px-4 text-xs text-text-muted">Date</th>
                                    <th className="border-b border-border-card py-2 px-4 text-xs text-text-muted">Duration</th>
                                    <th className="border-b border-border-card py-2 px-4 text-xs text-text-muted">Status</th>
                                    <th className="border-b border-border-card py-2 px-4" />
                                </tr>
                            </thead>
                            <tbody>
                                {maintenanceHistorySorted.map((maintenance) => (
                                    <tr key={maintenance.id} className="text-text-muted">
                                        <td className="border-b font-medium border-border-card py-2 px-4">{maintenance.description}</td>
                                        <td className="border-b border-border-card py-2 px-4">{formatDateRange(maintenance.start, maintenance.end)}</td>
                                        <td className="border-b border-border-card py-2 px-4">{durationInMinutes(maintenance.start, maintenance.end)} min</td>
                                        <td className="border-b border-border-card py-2 px-4">
                                            <Pill variant={maintenance.status === 'done' ? 'lightGreen' : maintenance.status === 'skipped' ? 'amber' : 'lightBlue'} className="border-none px-2 text-xs">
                                                {upperCaseStatus(maintenance.status)}
                                            </Pill>
                                        </td>
                                        <td className="border-b border-border-card py-2 px-4" />
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                )
            )}
        </SectionCard>
    )
}