import { Button } from '@/components/ui/button'
import { StatusBadge } from './StatusBadge'
import type { MaintenanceEntry } from '@/types/maintenance'
import { cellPaddingStyling, HeaderCell } from '../layout/TableCells'
import { cn } from '@/lib/utils'

function fmtWindow(start: string, end: string): string {
    const s = new Date(start)
    const e = new Date(end)
    const date = s.toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })
    const startTime = s.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' })
    const endTime = e.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' })
    return `${date}, ${startTime}–${endTime} UTC`
}

export function MaintenanceCard({ entry }: { entry: MaintenanceEntry }) {
    return (
        <div className="bg-white border border-gray-200 rounded-lg p-4 flex items-start justify-between gap-4">
            <div className="flex flex-col gap-1">
                <p className="text-sm font-medium text-gray-900">{entry.description}</p>
                <p className="text-xs text-gray-500 font-mono">{fmtWindow(entry.start, entry.end)}</p>
            </div>
            <StatusBadge status={entry.status} />
        </div>
    )
}

export function ScheduledSection({ entries }: { entries: MaintenanceEntry[] }) {
    if (entries.length === 0) {
        return <p className="text-sm text-gray-500 py-2">No maintenance scheduled.</p>
    }
    return (
        <div className="flex flex-col gap-3">
            {entries.map(entry => (
                <MaintenanceCard key={entry.id} entry={entry} />
            ))}
        </div>
    )
}

export function HistoryTable({ entries }: { entries: MaintenanceEntry[] }) {
    if (entries.length === 0) {
        return <p className="text-sm text-gray-500 py-2">No history yet.</p>
    }
    return (
        <div className="rounded-lg border bg-surface-card overflow-hidden w-full">
            <div className="overflow-x-auto">
                <table className="w-full border-collapse text-sm">
                    <thead>
                        <tr className="border-b border-gray-200">
                            <HeaderCell>Description</HeaderCell>
                            <HeaderCell>Window</HeaderCell>
                            <HeaderCell>Status</HeaderCell>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-100">
                        {entries.map(entry => (
                            <tr key={entry.id} className="border-b last:border-0 hover:bg-gray-50">
                                <td className={cn("text-gray-900", cellPaddingStyling)}>{entry.description}</td>
                                <td className={cn("text-gray-500 font-mono text-xs whitespace-nowrap", cellPaddingStyling)}>
                                    {fmtWindow(entry.start, entry.end)}
                                </td>
                                <td className={cn(cellPaddingStyling)}>
                                    <StatusBadge status={entry.status} />
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    )
}