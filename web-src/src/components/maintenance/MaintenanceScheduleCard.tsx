import { Maintenance } from "@/types/maintenance"
import { formatDateRange, durationInMinutes } from "./MaintenanceHistoryCard"
import { Pill } from "../ui/pill"
import { addMaintenance, upperCaseStatus } from "@/services/maintenance"
import { Button } from "../ui/button"
import { Response } from "../ui/response"
import { Tooltip, TooltipContent, TooltipTrigger } from "../ui/tooltip"
import { Pencil, CalendarClock } from "lucide-react"
import { useState } from "react"
import { DeleteMaintenance } from "./DeleteMaintenance"

function AddMaintenanceForm({onClose, onError} : { onClose: () => void, onError: (msg: string) => void }) {
    const [validationError, setValidationError] = useState<string | null>(null)

    const pad = (n: number) => String(n).padStart(2, '0')
    const now = new Date()
    const minNow = `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}T${pad(now.getHours())}:${pad(now.getMinutes())}`

    return(
        <form onSubmit={(e) => {
            e.preventDefault()
            const fd = new FormData(e.currentTarget)
            const start = fd.get("start") as string
            const end = fd.get("end") as string
            if (new Date(start) <= new Date()) {
                setValidationError("Start date must be in the future.")
                return
            }
            if (new Date(end) <= new Date(start)) {
                setValidationError("End date must be after the start date.")
                return
            }
            setValidationError(null)
            const toDateTime = (v: string) => v.replace("T", " ") + (v.length === 16 ? ":00" : "")
            addMaintenance({
                description: fd.get("description") as string,
                start: toDateTime(start),
                end: toDateTime(end),
            }).then(() => onClose()).catch((err) => onError(err?.message ?? "Failed to add maintenance."))
        }} className="bg-white border border-blue-200 rounded-lg p-4 mb-4 shadow">
            {validationError && (
                <p className="text-destructive text-xs mb-2">{validationError}</p>
            )}
            <div className="grid grid-cols-[1fr_auto_auto_auto] gap-2 items-end">
                <div className="flex flex-col gap-1">
                <label className="text-sm font-medium">New maintenance</label>
                <input name="description" placeholder="Description" required className="border rounded px-2 py-1.5 text-sm" />
                </div>
                <div className="flex flex-col gap-1">
                    <label className="text-xs text-gray-500">Start</label>
                    <input name="start" type="datetime-local" required min={minNow} className="border rounded px-2 py-1.5 text-sm" />
                </div>
                <div className="flex flex-col gap-1">
                    <label className="text-xs text-gray-500">End</label>
                    <input name="end" type="datetime-local" required min={minNow} className="border rounded px-2 py-1.5 text-sm" />
                </div>
                <Button type="submit" variant="orange" size="sm">Save</Button>
            </div>
        </form>
    )
}

export function MaintenanceScheduleCard({ maintenanceSchedule, onRefresh }: { maintenanceSchedule: Maintenance[], onRefresh: () => void }) {
    const maintenanceScheduleSorted = [...maintenanceSchedule].sort((a, b) => {
        const priority = (s: string) => s === 'in_progress' ? 0 : 1
        return priority(a.status) - priority(b.status) || new Date(a.start).getTime() - new Date(b.start).getTime()
    })
    const [showForm, setShowForm] = useState(false)
    const [deletingId, setDeletingId] = useState<string | null>(null)
    const [response, setResponse] = useState<{ open: boolean, status: 'success' | 'error', message: string }>({ open: false, status: 'success', message: '' })

    const maintenanceScheduleDataLength = maintenanceSchedule.length
    const hasInProgress = maintenanceSchedule.some(m => m.status === 'in_progress')
    const borderAccent = hasInProgress ? 'border-l-amber-500' : 'border-l-blue-400'
    const iconColor = hasInProgress ? 'text-amber-500' : 'text-blue-400'
    if (maintenanceScheduleDataLength === 0) {
        return (
            <div className="mt-2">
                <Response 
                    open={response.open} 
                    onClose={() => { setResponse(r => ({ ...r, open: false }))
                    if (response.status === 'success') 
                    onRefresh() }} 
                    status={response.status} 
                    message={response.message} />
                <div className={`bg-white rounded-lg shadow p-4 min-w-[300px] border border-border-card border-l-4 ${borderAccent}`}>
                    <div className="flex justify-between items-center mb-4">
                        <h3 className="text-lg font-semibold flex items-center gap-2">
                            <CalendarClock className={`w-5 h-5 ${iconColor}`} />
                            Upcoming maintenance
                        </h3>
                        <Button variant="orange" size="sm" onClick={() => setShowForm(!showForm)}>
                            {showForm ? "Cancel" : "Add maintenance"}
                        </Button>
                    </div>
                    {showForm && (
                        <>
                            <AddMaintenanceForm
                                onClose={() => { setShowForm(false); setResponse({ open: true, status: 'success', message: 'Maintenance schedule added successfully.' }) }}
                                onError={(msg) => { setShowForm(false); setResponse({ open: true, status: 'error', message: msg }) }}
                            />
                            <hr className="border-border-card mb-4" />
                        </>
                    )}
                    <p className="text-gray-500">No maintenance schedule available.</p>
                </div>
            </div>
        )
    }

    return (
        <div className="mt-2">
            {deletingId && (
                <DeleteMaintenance
                    maintenance={maintenanceSchedule.find(m => m.id === deletingId)!}
                    open={true}
                    onClose={() => setDeletingId(null)}
                    onDeleted={() => { setDeletingId(null); onRefresh() }}
                />
            )}
            <Response 
                open={response.open} 
                onClose={() => { setResponse(r => ({ ...r, open: false }))
                if (response.status === 'success') 
                onRefresh() }} 
                status={response.status} 
                message={response.message} />
            <div className={`bg-white rounded-lg shadow p-4 min-w-[300px] border border-border-card border-l-4 ${borderAccent}`}>
                <div className="flex justify-between items-center mb-4">
                    <h3 className="text-lg font-semibold flex items-center gap-2">
                        <CalendarClock className={`w-5 h-5 ${iconColor}`} />
                        Upcoming maintenance
                    </h3>
                    <Button variant="orange" size="sm" onClick={() => setShowForm(!showForm)}>
                        {showForm ? "Cancel" : "Add maintenance"}
                    </Button>
                </div>
                {showForm && (
                    <>
                        <AddMaintenanceForm
                            onClose={() => { setShowForm(false); setResponse({ open: true, status: 'success', message: 'Maintenance schedule added successfully.' }) }}
                            onError={(msg) => { setShowForm(false); setResponse({ open: true, status: 'error', message: msg }) }}
                        />
                        <hr className="border-border-card mb-4" />
                    </>
                )}
                <div className="overflow-y-auto max-h-64">
                    <table className="w-full text-left border-collapse table-fixed">
                        <colgroup>
                            <col />
                            <col className="w-52" />
                            <col className="w-20" />
                            <col className="w-28" />
                            <col className="w-32" />
                        </colgroup>
                        <thead>
                            <tr>
                                <th className="border-b border-border-card py-2 px-4 text-xs text-gray-500">Description</th>
                                <th className="border-b border-border-card py-2 px-4 text-xs text-gray-500">Date</th>
                                <th className="border-b border-border-card py-2 px-4 text-xs text-gray-500">Duration</th>
                                <th className="border-b border-border-card py-2 px-4 text-xs text-gray-500">Status</th>
                                <th className="border-b border-border-card py-2 px-4 text-xs text-gray-500 text-right">Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                        {maintenanceScheduleSorted.map((maintenance, i) => {
                            const prevStatus = maintenanceScheduleSorted[i - 1]?.status
                            const showDivider = i > 0 && maintenance.status !== 'in_progress' && prevStatus === 'in_progress'
                            return (
                                <>
                                {showDivider && (
                                    <tr key={`divider-${i}`}>
                                        <td colSpan={5} className="py-1 px-4">
                                            <div className="flex items-center gap-2 text-xs text-gray-400">
                                                <div className="flex-1 border-t border-dashed border-gray-300" />
                                                <span>Scheduled</span>
                                                <div className="flex-1 border-t border-dashed border-gray-300" />
                                            </div>
                                        </td>
                                    </tr>
                                )}
                                <tr key={maintenance.id} className={maintenance.status === 'in_progress' ? 'bg-amber-50 hover:bg-amber-100' : 'hover:bg-gray-50'}>
                                <td className="border-b border-border-card py-2 px-4 text-sm font-medium">
                                    <span className="flex items-center gap-2">
                                        {maintenance.status === 'in_progress' && (
                                            <span className="relative flex h-2 w-2 shrink-0">
                                                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-amber-400 opacity-75" />
                                                <span className="relative inline-flex rounded-full h-2 w-2 bg-amber-500" />
                                            </span>
                                        )}
                                        {maintenance.description}
                                    </span>
                                </td>
                                <td className="border-b border-border-card py-2 px-4 text-sm">{formatDateRange(maintenance.start, maintenance.end)}</td>
                                <td className="border-b border-border-card py-2 px-4 text-sm">
                                    {durationInMinutes(maintenance.start, maintenance.end)} min
                                </td>
                                <td className="border-b border-border-card py-2 px-4">
                                    <Pill variant={maintenance.status === 'done' ? 'lightGreen' : maintenance.status === 'in_progress' ? 'amber' : 'lightBlue'} className="border-none px-2 text-xs">
                                        {upperCaseStatus(maintenance.status)}
                                    </Pill>
                                </td>
                                <td className="border-b border-border-card py-2 px-4">
                                    <div className="flex items-center justify-end gap-2">
                                        <Tooltip>
                                            <TooltipTrigger asChild>
                                                <Button
                                                    variant="ghost"
                                                    size="sm"
                                                    className="p-0 text-gray-400 hover:text-gray-700"
                                                    onClick={() => window.location.href = `/maintenance/edit?id=${maintenance.id}`}
                                                >
                                                    <Pencil className="h-4 w-4" />
                                                </Button>
                                            </TooltipTrigger>
                                            <TooltipContent>Edit</TooltipContent>
                                        </Tooltip>
                                        <Button variant="destructive" size="xs" onClick={() => setDeletingId(maintenance.id)}>
                                            Delete
                                        </Button>
                                    </div>
                                </td>
                            </tr>
                                </>
                            )
                        })}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    )
}