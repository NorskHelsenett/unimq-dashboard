import { Maintenance } from "@/types/maintenance"
import { formatDateRange } from "./MaintenanceHistoryCard"
import { Pill } from "../ui/pill"
import { addMaintenance, upperCaseStatus } from "@/services/maintenance"
import { Button } from "../ui/button"
import { Response } from "../ui/response"
import { Tooltip, TooltipContent, TooltipTrigger } from "../ui/tooltip"
import { Pencil } from "lucide-react"
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
        }} className="border border-border-card rounded-lg p-3 mb-3 bg-gray-50">
            {validationError && (
                <p className="text-destructive text-xs mb-2">{validationError}</p>
            )}
            <div className="grid grid-cols-[1fr_auto_auto_auto] gap-2 items-end">
                <input name="description" placeholder="Description" required className="border rounded px-2 py-1.5 text-sm" />
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
    const maintenanceScheduleSorted = [...maintenanceSchedule].sort((a, b) => new Date(a.start).getTime() - new Date(b.start).getTime())
    const [showForm, setShowForm] = useState(false)
    const [deletingId, setDeletingId] = useState<string | null>(null)
    const [response, setResponse] = useState<{ open: boolean, status: 'success' | 'error', message: string }>({ open: false, status: 'success', message: '' })

    const maintenanceScheduleDataLength = maintenanceSchedule.length
    if (maintenanceScheduleDataLength === 0) {
        return (
            <div className="bg-white rounded-lg shadow p-4 min-w-[300px] border border-border-card">
                <div className="flex justify-between items-center mb-2">
                    <h3 className="text-lg font-semibold mb-2">Maintenance schedule</h3>
                    <Button variant="orange" size="sm" className="mb-2 mr-2" onClick={() => setShowForm(!showForm)}>
                            {showForm ? "Cancel add" : "Add maintenance"}
                    </Button>
                </div>
                {showForm && (
                    <AddMaintenanceForm
                        onClose={() => { setShowForm(false); setResponse({ open: true, status: 'success', message: 'Maintenance schedule added successfully.' }) }}
                        onError={(msg) => { setShowForm(false); setResponse({ open: true, status: 'error', message: msg }) }}
                    />
                )}
                <p className="text-gray-500">No maintenance schedule available.</p>
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
            <div className="bg-white rounded-lg shadow p-4 min-w-[300px] border border-border-card">
                <div className="flex justify-between items-center mb-2">
                    <h3 className="text-lg font-semibold mb-2">Maintenance schedule</h3>
                    <div>
                        
                        <Button variant="orange" size="sm" className="mb-2 mr-2" onClick={() => setShowForm(!showForm)}>
                            {showForm ? "Cancel add" : "Add maintenance"}
                        </Button>
                    
                        {/* <Button variant="outline" size="sm" className="mb-2" onClick={() => setAdministrative(!administrative)}>
                            {administrative ? "Done" : "Administrate"}
                        </Button> */}
                    </div>
                </div>
                {showForm && (
                    <AddMaintenanceForm
                        onClose={() => { setShowForm(false); setResponse({ open: true, status: 'success', message: 'Maintenance schedule added successfully.' }) }}
                        onError={(msg) => { setShowForm(false); setResponse({ open: true, status: 'error', message: msg }) }}
                    />
                )}
                <div className="overflow-y-auto max-h-64">
                    <table className="w-full text-left border-collapse">
                        <thead>
                            <tr>
                                <th className="border-b border-border-card py-2 px-4 text-xs text-gray-500">Description</th>
                                <th className="border-b border-border-card py-2 px-4 text-xs text-gray-500">Date</th>
                                <th className="border-b border-border-card py-2 px-4 text-xs text-gray-500">Status</th>
                            </tr>
                        </thead>
                        <tbody>
                        {maintenanceScheduleSorted.map((maintenance) => (
                            <tr key={maintenance.id} className="hover:bg-gray-50">
                                <td className="border-b border-border-card py-2 px-4 text-sm font-medium ">{maintenance.description}</td>
                                <td className="border-b border-border-card py-2 px-4 text-sm">{formatDateRange(maintenance.start, maintenance.end)}</td>
                                <td className="border-b border-border-card py-2">
                                    <Pill variant={maintenance.status === 'done' ? 'lightGreen' : maintenance.status === 'skipped' ? 'amber' : 'lightBlue'} className="border-none px-2 text-xs">
                                        {upperCaseStatus(maintenance.status)}
                                    </Pill>
                                </td>
                                <td className="border-b border-border-card py-2 px-2">
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
                                </td>
                                    <td className="border-b border-border-card py-2">
                                        <Button variant="destructive" size="xs" onClick={() => setDeletingId(maintenance.id)}>
                                            Delete
                                        </Button>
                                    </td>
                                
                            </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    )
}