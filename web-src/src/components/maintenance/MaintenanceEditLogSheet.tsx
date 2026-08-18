import { useEffect, useState } from "react"
import { Sheet, SheetContent, SheetHeader, SheetTitle } from "../ui/sheet"
import { Pencil } from "lucide-react"
import { getMaintenanceEditLogs, MaintenanceEditLog } from "@/services/maintenance"

const fmtUtc = (iso: string) => {
    const d = new Date(iso)
    const pad = (n: number) => String(n).padStart(2, "0")
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export function MaintenanceEditLogSheet({
    maintenanceId,
    description,
    open,
    onClose,
}: {
    maintenanceId: string
    description: string
    open: boolean
    onClose: () => void
}) {
    const [logs, setLogs] = useState<MaintenanceEditLog[]>([])
    const [loading, setLoading] = useState(false)

    useEffect(() => {
        if (!open) return
        setLoading(true)
        getMaintenanceEditLogs(maintenanceId)
            .then(data => setLogs([...data].reverse()))
            .catch(() => setLogs([]))
            .finally(() => setLoading(false))
    }, [open, maintenanceId])

    return (
        <Sheet open={open} onOpenChange={o => !o && onClose()}>
            <SheetContent side="right" className="w-full sm:max-w-md flex flex-col">
                <SheetHeader>
                    <SheetTitle>Changelog — {description}</SheetTitle>
                </SheetHeader>
                <div className="flex-1 overflow-y-auto px-4 pb-4">
                    {loading && <p className="text-sm text-text-muted mt-4">Loading…</p>}
                    {!loading && logs.length === 0 && (
                        <p className="text-sm text-text-muted mt-4">No edits recorded yet.</p>
                    )}
                    {!loading && logs.length > 0 && (
                        <ol className="mt-4 space-y-3">
                            {logs.map(log => (
                                <li key={log.id} className="flex gap-3 rounded-lg border border-gray-200 bg-gray-50 px-4 py-3">
                                    <Pencil className="mt-0.5 shrink-0 text-gray-400" size={15} />
                                    <div className="min-w-0 flex-1">
                                        <div className="flex items-center justify-between gap-2">
                                            <p className="text-sm font-medium text-text-primary truncate">{log.updated_by}</p>
                                            <p className="text-xs text-text-muted shrink-0">{fmtUtc(log.updated_at)}</p>
                                        </div>
                                        <p className="text-xs text-text-muted mt-1 italic">"{log.reason}"</p>
                                        <div className="mt-2 text-xs text-gray-500 space-y-0.5">
                                            <p><span className="font-medium">Description:</span> {log.description}</p>
                                            <p><span className="font-medium">Start:</span> {fmtUtc(log.start)}</p>
                                            <p><span className="font-medium">End:</span> {fmtUtc(log.end)}</p>
                                        </div>
                                    </div>
                                </li>
                            ))}
                        </ol>
                    )}
                </div>
            </SheetContent>
        </Sheet>
    )
}
