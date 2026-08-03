import { useEffect, useState } from "react"
import { Sheet, SheetContent, SheetHeader, SheetTitle } from "../ui/sheet"
import { AlertTriangle, CheckCircle2 } from "lucide-react"
import { getAlarmLogs} from '@/services/notifications'
import type { LogEntry } from '@/types/notifications'

export const AlarmLogSheet = ({
    alarmId,
    alarmName,
    alarmType,
    open,
    onClose,
}: {
    alarmId: string
    alarmName: string
    alarmType?: string
    open: boolean
    onClose: () => void
}) => {
    const [entries, setEntries] = useState<LogEntry[]>([])
    const [loading, setLoading] = useState(false)

    useEffect(() => {
        if (!open) return
        setLoading(true)
        getAlarmLogs(alarmId, alarmType)
            .then(data => setEntries(data))
            .catch(() => setEntries([]))
            .finally(() => setLoading(false))
    }, [open, alarmId, alarmType])

    return (
        <Sheet open={open} onOpenChange={o => !o && onClose()}>
            <SheetContent side="right" className="w-full sm:max-w-md flex flex-col">
                <SheetHeader>
                    <SheetTitle>History for {alarmName}</SheetTitle>
                </SheetHeader>
                <div className="flex-1 overflow-y-auto px-4 pb-4">
                    {loading && (
                        <p className="text-sm text-text-muted mt-4">Loading...</p>
                    )}
                    {!loading && entries.length === 0 && (
                        <p className="text-sm text-text-muted mt-4">No events recorded yet.</p>
                    )}
                    {!loading && entries.length > 0 && (
                        <ol className="mt-4 space-y-3">
                            {entries.map((e, i) => {
                                const isFired = e.event === "fired"
                                return isFired ? (
                                    <li key={i} className="flex gap-3 rounded-lg border border-orange-200 bg-orange-50 px-4 py-3">
                                        <AlertTriangle className="mt-0.5 shrink-0 text-orange-500" size={18} />
                                        <div>
                                            <p className="text-sm font-semibold text-orange-800">Fired</p>
                                            <p className="text-xs text-orange-600 mt-0.5">
                                                {new Date(e.ts).toLocaleString("no-NO", {
                                                    day: "2-digit", month: "short", year: "numeric",
                                                    hour: "2-digit", minute: "2-digit",
                                                })}
                                            </p>
                                            {e.value !== undefined && (
                                                <p className="text-xs text-orange-700 mt-1">
                                                    Value <span className="font-semibold">{e.value}</span> exceeded threshold of <span className="font-semibold">{e.threshold}</span>
                                                </p>
                                            )}
                                        </div>
                                    </li>
                                ) : (
                                    <li key={i} className="flex gap-3 rounded-lg border border-gray-200 bg-gray-50 px-4 py-2.5">
                                        <CheckCircle2 className="mt-0.5 shrink-0 text-gray-400" size={15} />
                                        <div>
                                            <p className="text-xs font-medium text-gray-500">Resolved</p>
                                            <p className="text-xs text-gray-400">
                                                {new Date(e.ts).toLocaleString("no-NO", {
                                                    day: "2-digit", month: "short", year: "numeric",
                                                    hour: "2-digit", minute: "2-digit",
                                                })}
                                                {e.value !== undefined && <> · back to {e.value}</>}
                                            </p>
                                        </div>
                                    </li>
                                )
                            })}
                        </ol>
                    )}
                </div>
            </SheetContent>
        </Sheet>
    )
}
