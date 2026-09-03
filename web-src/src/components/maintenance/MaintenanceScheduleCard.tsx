import React from "react"
import { Maintenance } from "@/types/maintenance"
import { formatDateRange, durationInMinutes } from "./MaintenanceHistoryCard"
import { Pill } from "../ui/pill"
import { addMaintenance, upperCaseStatus } from "@/services/maintenance"
import { Button } from "../ui/button"
import { Response } from "../ui/response"
import { Tooltip, TooltipContent, TooltipTrigger } from "../ui/tooltip"
import { Pencil, CalendarClock } from "lucide-react"
import { cn } from "@/lib/utils"
import { useState } from "react"
import { DeleteMaintenance } from "./DeleteMaintenance"
import { SectionCard, SectionCardHeader } from "../ui/section-card"
import { StatusDot } from "../ui/status-dot"

function AddMaintenanceForm({onClose, onCancel, onError} : { onClose: () => void, onCancel: () => void, onError: (msg: string) => void }) {
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
        }} className="bg-surface-card border border-blue-200 rounded-lg p-4 mb-4 shadow">
            {validationError && (
                <p className="text-destructive text-xs mb-2">{validationError}</p>
            )}
            <div className="grid grid-cols-[1fr_auto_auto] gap-2 items-end mb-2">
                <div className="flex flex-col gap-1">
                    <label className="text-sm font-medium">New maintenance</label>
                    <input name="description" placeholder="Description" required className="border rounded px-2 py-1.5 text-sm" />
                </div>
                <Button type="submit" variant="orange" size="sm">Save</Button>
                <Button type="button" variant="ghost" size="sm" onClick={onCancel}>Cancel</Button>
            </div>
            <div className="grid grid-cols-2 gap-2">
                <div className="flex flex-col gap-1">
                    <label className="text-xs text-text-muted">Start</label>
                    <input name="start" type="datetime-local" required min={minNow} className="border rounded px-2 py-1.5 text-sm dark:[color-scheme:dark]" />
                </div>
                <div className="flex flex-col gap-1">
                    <label className="text-xs text-text-muted">End</label>
                    <input name="end" type="datetime-local" required min={minNow} className="border rounded px-2 py-1.5 text-sm dark:[color-scheme:dark]" />
                </div>
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

    const hasInProgress = maintenanceSchedule.some(m => m.status === 'in_progress')
    const accent = hasInProgress ? 'amber' : 'blue'
    const iconColor = hasInProgress ? 'text-amber-500' : 'text-blue-400'

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
                onClose={() => {
                    setResponse(r => ({ ...r, open: false }))
                    if (response.status === 'success') onRefresh()
                }}
                status={response.status}
                message={response.message}
            />
            <SectionCard accent={accent as 'amber' | 'blue'}>
                <SectionCardHeader
                    title="Upcoming maintenance"
                    icon={<CalendarClock className={`w-5 h-5 ${iconColor}`} />}
                    action={
                        <Button variant="outline" size="sm" onClick={() => setShowForm(true)} disabled={showForm}>
                            Add maintenance
                        </Button>
                    }
                />
                {showForm && (
                    <>
                        <AddMaintenanceForm
                            onClose={() => { setShowForm(false); setResponse({ open: true, status: 'success', message: 'Maintenance schedule added successfully.' }) }}
                            onCancel={() => setShowForm(false)}
                            onError={(msg) => { setShowForm(false); setResponse({ open: true, status: 'error', message: msg }) }}
                        />
                        <hr className="border-border-card mb-4" />
                    </>
                )}
                {maintenanceSchedule.length === 0 ? (
                    <p className="text-text-muted">No maintenance schedule available.</p>
                ) : (
                    <div className="overflow-y-auto max-h-64">
                        <table className="w-full text-left border-collapse">
                            <thead>
                                <tr>
                                    <th className="border-b border-border-card py-2 px-4 text-xs text-text-muted">Description</th>
                                    <th className="border-b border-border-card py-2 px-4 text-xs text-text-muted">Date</th>
                                    <th className="border-b border-border-card py-2 px-4 text-xs text-text-muted">Duration</th>
                                    <th className="border-b border-border-card py-2 px-4 text-xs text-text-muted">Status</th>
                                    <th className="border-b border-border-card py-2 px-4 text-xs text-text-muted text-right">Actions</th>
                                </tr>
                            </thead>
                            <tbody>
                            {maintenanceScheduleSorted.map((maintenance, i) => {
                                const prevStatus = maintenanceScheduleSorted[i - 1]?.status
                                const showDivider = i > 0 && maintenance.status !== 'in_progress' && prevStatus === 'in_progress'
                                return (
                                    <React.Fragment key={maintenance.id}>
                                    {showDivider && (
                                        <tr key={`divider-${i}`}>
                                            <td colSpan={5} className="py-1 px-4">
                                                <div className="flex items-center gap-2 text-xs text-text-muted">
                                                    <div className="flex-1 border-t border-dashed border-border-card" />
                                                    <span>Scheduled</span>
                                                    <div className="flex-1 border-t border-dashed border-border-card" />
                                                </div>
                                            </td>
                                        </tr>
                                    )}
                                    <tr
                        className={cn('cursor-pointer', maintenance.status === 'in_progress' ? 'bg-maintenance-in-progress-bg hover:bg-maintenance-in-progress-bg-hover' : 'hover:bg-surface-page')}
                        onClick={() => {
                            const vhost = new URLSearchParams(window.location.search).get('vhost')
                            const vhostSuffix = vhost ? `&vhost=${encodeURIComponent(vhost)}` : ''
                            window.location.href = `/maintenance/edit?id=${maintenance.id}${vhostSuffix}`
                        }}
                    >
                                        <td className="border-b border-border-card py-2 px-4 text-sm font-medium">
                                            <span className="flex items-center gap-2">
                                                {maintenance.status === 'in_progress' && <StatusDot color="warning" />}
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
                                        <td className="border-b border-border-card py-2 px-4" onClick={e => e.stopPropagation()}>
                                            <div className="flex items-center justify-end gap-2">
                                                <Tooltip>
                                                    <TooltipTrigger asChild>
                                                        <Button
                                                            variant="ghost"
                                                            size="sm"
                                                            className="p-0 text-text-muted hover:text-text-secondary"
                                                            onClick={() => {
                                                                const vhost = new URLSearchParams(window.location.search).get('vhost')
                                                                const vhostSuffix = vhost ? `&vhost=${encodeURIComponent(vhost)}` : ''
                                                                window.location.href = `/maintenance/edit?id=${maintenance.id}${vhostSuffix}`
                                                            }}
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
                                    </React.Fragment>
                                )
                            })}
                            </tbody>
                        </table>
                    </div>
                )}
            </SectionCard>
        </div>
    )
}