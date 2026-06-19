import { cn } from '@/lib/utils'
import type { MaintenanceEntry } from '@/types/maintenance'

type Status = MaintenanceEntry['status']

const STATUS_CONFIG: Record<Status, { label: string; className: string }> = {
    scheduled: {
        label: 'Scheduled',
        className: 'text-blue-700 bg-blue-50 border border-blue-200',
    },
    done: {
        label: 'Done',
        className: 'text-green-700 bg-green-50 border border-green-200',
    },
    skipped: {
        label: 'Skipped',
        className: 'text-gray-500 bg-gray-100 border border-gray-200',
    },
}

export function StatusBadge({ status }: { status: Status }) {
    const config = STATUS_CONFIG[status]
    return (
        <span className={cn('inline-flex items-center rounded px-2 py-0.5 text-xs font-medium', config.className)}>
            {config.label}
        </span>
    )
}