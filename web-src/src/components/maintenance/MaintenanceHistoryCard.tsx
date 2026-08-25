import { Maintenance } from "@/types/maintenance"
import { upperCaseStatus } from "@/services/maintenance"
import { Pill } from "../ui/pill"
import { ChevronDown, ChevronUp, History } from "lucide-react"
import { useLocalStorage } from "@/hooks/useLocalStorage"
import { SectionCard, SectionCardHeader } from "../ui/section-card"

export function formatDateRange(start: string, end: string): string {
    const pad = (n: number) => String(n).padStart(2, '0')
    const fmtDate = (d: Date) =>
        `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
    const fmtTime = (d: Date) =>
        `${pad(d.getHours())}:${pad(d.getMinutes())}`

    const s = new Date(start)
    const e = new Date(end)
    const sameDay =
        s.getFullYear() === e.getFullYear() &&
        s.getMonth() === e.getMonth() &&
        s.getDate() === e.getDate()
    return sameDay ? `${fmtDate(s)} - ${fmtTime(e)}` : `${fmtDate(s)} - ${fmtDate(e)}`
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
                action={open ? <ChevronUp className="w-4 h-4 text-gray-400" /> : <ChevronDown className="w-4 h-4 text-gray-400" />}
                />
            </button>
            {open && (
                maintenanceHistory.length === 0 ? (
                    <p className="text-gray-500">No maintenance history available.</p>
                ) : (
                    <div className="overflow-y-auto max-h-64 text-sm">
                        <table className="w-full text-left border-collapse table-fixed">
                            <colgroup>
                                <col />
                                <col className="w-52" />
                                <col className="w-20" />
                                <col className="w-28" />
                                {/* spacer matches the Actions column width in the schedule table */}
                                <col className="w-32" />
                            </colgroup>
                            <thead>
                                <tr>
                                    <th className="border-b border-border-card py-2 px-4 text-xs text-gray-500">Description</th>
                                    <th className="border-b border-border-card py-2 px-4 text-xs text-gray-500">Date</th>
                                    <th className="border-b border-border-card py-2 px-4 text-xs text-gray-500">Duration</th>
                                    <th className="border-b border-border-card py-2 px-4 text-xs text-gray-500">Status</th>
                                    <th className="border-b border-border-card py-2 px-4" />
                                </tr>
                            </thead>
                            <tbody>
                                {maintenanceHistorySorted.map((maintenance) => (
                                    <tr key={maintenance.id} className="text-gray-500">
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