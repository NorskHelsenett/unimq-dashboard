import { useState } from "react"
import { useAuth } from "react-oidc-context"
import { Maintenance } from "@/types/maintenance"
import { updateMaintenance } from "@/services/maintenance"
import { Button } from "../ui/button"
import { DeleteMaintenance } from "./DeleteMaintenance"
import { MaintenanceEditLogSheet } from "./MaintenanceEditLogSheet"
import { Response } from "../ui/response"

const toDatetimeLocal = (datetime: string) => {
    const d = new Date(datetime)
    const pad = (n: number) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

const toServerDateTime = (v: string) =>
    v.replace("T", " ") + (v.length === 16 ? ":00" : "")

const nowLabel = () => {
    const now = new Date()
    const pad = (n: number) => String(n).padStart(2, "0")
    return `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())} ${pad(now.getHours())}:${pad(now.getMinutes())}:${pad(now.getSeconds())}`
}

export function EditMaintenance({ maintenance }: { maintenance: Maintenance }) {
    const auth = useAuth()
    const userName = auth.user?.profile?.name ?? auth.user?.profile?.email ?? "Unknown"

    const currentVhost = new URLSearchParams(window.location.search).get('vhost')
    const vhostParam = currentVhost ? `?vhost=${encodeURIComponent(currentVhost)}` : ''

    const initialStart = toDatetimeLocal(maintenance.start)
    const initialEnd = toDatetimeLocal(maintenance.end)

    const [description, setDescription] = useState(maintenance.description)
    const [start, setStart] = useState(initialStart)
    const [end, setEnd] = useState(initialEnd)
    // track the last-saved values to detect actual changes
    const [savedDescription, setSavedDescription] = useState(maintenance.description)
    const [savedStart, setSavedStart] = useState(initialStart)
    const [savedEnd, setSavedEnd] = useState(initialEnd)

    const hasChanges = description !== savedDescription || start !== savedStart || end !== savedEnd

    const [reason, setReason] = useState("")
    const [editedAt, setEditedAt] = useState(nowLabel)
    const [error, setError] = useState<string | null>(null)
    const [saving, setSaving] = useState(false)
    const [success, setSuccess] = useState(false)
    const [showDelete, setShowDelete] = useState(false)
    const [showLogs, setShowLogs] = useState(false)

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault()
        setError(null)

        if (!hasChanges) {
            setError("No changes to save — update the description or schedule first.")
            return
        }
        if (!reason.trim()) {
            setError("Reason for change is required.")
            return
        }
        if (new Date(end) <= new Date(start)) {
            setError("End time must be after start time.")
            return
        }

        setSaving(true)
        updateMaintenance({
            id: maintenance.id,
            description,
            start: toServerDateTime(start),
            end: toServerDateTime(end),
            reason: reason.trim(),
            updated_by: userName,
        })
            .then(() => {
                setSavedDescription(description)
                setSavedStart(start)
                setSavedEnd(end)
                setReason("")
                setEditedAt(nowLabel())
                setSuccess(true)
            })
            .catch((err: unknown) => {
                setError(err instanceof Error ? err.message : "Failed to update maintenance.")
            })
            .finally(() => setSaving(false))
    }

    return (
        <div className="max-w-4xl">
            <Response
                open={success}
                status="success"
                message="Maintenance entry updated successfully."
                onClose={() => setSuccess(false)}
            />
            {showDelete && (
                <DeleteMaintenance
                    maintenance={maintenance}
                    open={true}
                    onClose={() => setShowDelete(false)}
                    onDeleted={() => { window.location.href = `/maintenance${vhostParam}` }}
                />
            )}
            <MaintenanceEditLogSheet
                maintenanceId={maintenance.id}
                description={maintenance.description}
                open={showLogs}
                onClose={() => setShowLogs(false)}
            />
            <h1 className="text-3xl font-semibold mb-6">Edit maintenance</h1>

            <form onSubmit={handleSubmit} className="grid grid-cols-2 gap-6">
                {/* Main fields */}
                <div className="bg-surface-card border border-border-card rounded-lg p-5 flex flex-col gap-4">
                    <h2 className="text-sm font-semibold text-text-muted uppercase tracking-wide">Details</h2>

                    <div className="flex flex-col gap-1">
                        <label className="text-xs text-text-muted">Description</label>
                        <input
                            value={description}
                            onChange={e => setDescription(e.target.value)}
                            required
                            className="border border-border-card rounded px-3 py-2 text-sm"
                            placeholder="Description"
                        />
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <div className="flex flex-col gap-1">
                            <label className="text-xs text-text-muted">Start</label>
                            <input
                                type="datetime-local"
                                value={start}
                                onChange={e => setStart(e.target.value)}
                                required
                                className="border border-border-card rounded px-3 py-2 text-sm dark:[color-scheme:dark]"
                            />
                        </div>
                        <div className="flex flex-col gap-1">
                            <label className="text-xs text-text-muted">End</label>
                            <input
                                type="datetime-local"
                                value={end}
                                onChange={e => setEnd(e.target.value)}
                                required
                                className="border border-border-card rounded px-3 py-2 text-sm dark:[color-scheme:dark]"
                            />
                        </div>
                    </div>
                </div>

                {/* Audit trail */}
                <div className="bg-surface-card border border-border-card rounded-lg p-5 flex flex-col gap-4">
                    <h2 className="text-sm font-semibold text-text-muted uppercase tracking-wide">Change record</h2>

                    <div className="grid grid-cols-2 gap-4">
                        <div className="flex flex-col gap-1">
                            <label className="text-xs text-text-muted">Edited by</label>
                            <input
                                readOnly
                                disabled
                                value={userName}
                                className="border border-border-card rounded px-3 py-2 text-sm bg-surface-page text-text-muted cursor-not-allowed"
                            />
                        </div>
                        <div className="flex flex-col gap-1">
                            <label className="text-xs text-text-muted">Edited at</label>
                            <input
                                readOnly
                                disabled
                                value={editedAt}
                                className="border border-border-card rounded px-3 py-2 text-sm bg-surface-page text-text-muted cursor-not-allowed"
                            />
                        </div>
                    </div>

                    <div className="flex flex-col gap-1">
                        <label className="text-xs text-text-muted">
                            Reason for change <span className="text-destructive">*</span>
                        </label>
                        <textarea
                            value={reason}
                            onChange={e => setReason(e.target.value)}
                            required
                            rows={3}
                            placeholder="Describe why this maintenance entry is being changed..."
                            className="border border-border-card rounded px-3 py-2 text-sm resize-none"
                        />
                    </div>
                </div>

                {error && <p className="text-destructive text-sm col-span-2">{error}</p>}

                <div className="flex items-center justify-between col-span-2">
                    <Button
                        type="button"
                        variant="destructive"
                        size="sm"
                        onClick={() => setShowDelete(true)}
                    >
                        Delete maintenance
                    </Button>
                    <div className="flex gap-2">
                        <Button type="button" variant="outline" size="sm" onClick={() => setShowLogs(true)}>
                            View edit history
                        </Button>
                        <Button type="submit" variant="orange" disabled={saving}>
                            {saving ? "Saving…" : "Save changes"}
                        </Button>
                    </div>
                </div>
            </form>
        </div>
    )
}
