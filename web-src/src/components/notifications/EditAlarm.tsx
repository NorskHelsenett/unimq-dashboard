import { Switch } from "../ui/switch"
import { Pill } from "../ui/pill"
import { AlarmProps } from "@/types/notifications"
import { Button } from "../ui/button"
import { Input } from "../ui/input"
import { useState } from "react"
import { DropdownMenu } from "radix-ui"
import { DeleteItem } from "./DeleteItem"
import { AlarmLogSheet } from "./AlarmLogSheet"
import { Response } from "../ui/response"
import { toggleRule, updateRule, testRule } from '@/services/notifications'

export const EditAlarm = ({ alarm, vhost }: { alarm: AlarmProps, vhost: string }) => {
    const [deletingId, setDeletingId] = useState<string | null>(null)
    const alarmId = alarm.id
    const hasAlarmId = alarmId !== undefined && alarmId !== null
    const [showLogs, setShowLogs] = useState(false)
    const [updated, setUpdated] = useState(false)
    const [testResult, setTestResult] = useState<{ status: 'success' | 'error', message: string } | null>(null)
    const [updateResult, setUpdateResult] = useState<{ status: 'success' | 'error', message: string } | null>(null)
    const [updateSummary, setUpdateSummary] = useState("")
    const [threshold, setThreshold] = useState(alarm.threshold?.toString() ?? "")
    const [message, setMessage] = useState(alarm.message ?? "")
    const hasUnsavedChanges =
        threshold !== (alarm.threshold?.toString() ?? "") ||
        message !== (alarm.message ?? "")

    const redirectAfterDelete = () => {
        window.location.href = `/notifications?vhost=${encodeURIComponent(vhost)}`
    }

    const ensureOk = async (res: globalThis.Response, fallbackMessage: string) => {
        if (res.ok) return
        const errorText = await res.text().catch(() => "")
        throw new Error(errorText || fallbackMessage)
    }

    const secondaryActions = [
        { label: "Reset to default message", onClick: () => { resetToDefaultMessage() } },
        { label: "View history/logs", onClick: () => { setShowLogs(true) } },
        { label: "Delete alarm", onClick: () => { alarm.id && setDeletingId(alarm.id) } },
    ]

    const [enabledIds, setEnabledIds] = useState<Set<string>>(
                () => new Set(alarm.enabled && alarm.id ? [alarm.id] : [])
            )
    const toggleAlarm = (id: string) => {
        setEnabledIds(prev => {
            const next = new Set(prev)
            next.has(id) ? next.delete(id) : next.add(id)
            return next
        })
        toggleRule(vhost, id)
    }


    const updateAlarm = (id: string) => {
        const changes: string[] = []
        if (threshold !== (alarm.threshold?.toString() ?? "")) changes.push(`threshold → ${threshold}`)
        if (message !== (alarm.message ?? "")) changes.push(`message → "${message || "(default)"}"`)
        setUpdated(false)
        setUpdateResult(null)
        updateRule(vhost, id, Number(threshold), message)
            .then(async (res) => {
                await ensureOk(res, 'Failed to update alarm.')
                const summary = `Updated: ${changes.join(", ")}`
                setUpdateSummary(summary)
                setUpdateResult({ status: 'success', message: summary })
                setUpdated(true)
                setTimeout(() => window.location.reload(), 2500)
            })
            .catch((error: unknown) => {
                const errorMessage = error instanceof Error ? error.message : 'Failed to update alarm.'
                setUpdateSummary(errorMessage)
                setUpdateResult({ status: 'error', message: errorMessage })
                setUpdated(false)
            })
    }

    const resetToDefaultMessage = () => {
        if (!hasAlarmId) {
            return
        }

        setUpdated(false)
        setUpdateResult(null)
        updateRule(vhost, alarmId, Number(threshold), '')
            .then(async (res) => {
                await ensureOk(res, 'Failed to reset message to default.')
                const summary = `Updated: message → "(default)"`
                setUpdateSummary(summary)
                setUpdateResult({ status: 'success', message: summary })
                setUpdated(true)
                setTimeout(() => window.location.reload(), 2500)
            })
            .catch((error: unknown) => {
                const errorMessage = error instanceof Error ? error.message : 'Failed to reset message to default.'
                setUpdateSummary(errorMessage)
                setUpdateResult({ status: 'error', message: errorMessage })
                setUpdated(false)
            })
    }

    const testNotification = () => {
        if (!hasAlarmId) {
            return
        }

        testRule(vhost, alarmId)
            .then(result => {
                setTestResult({ status: result.success ? 'success' : 'error', message: result.message })
            })
            .catch(() => setTestResult({ status: 'error', message: 'Failed to reach server.' }))
    }

    return (
        <div className="mt-2">
        <DeleteItem alarm={alarm} vhost={vhost} open={deletingId !== null} onClose={() => setDeletingId(null)} onDeleted={redirectAfterDelete} />
        {hasAlarmId && (
            <AlarmLogSheet alarmId={alarmId} alarmName={alarm.name ?? ""} open={showLogs} onClose={() => setShowLogs(false)} />
        )}
        <Response onClose={() => setUpdated(false)} open={updated} status="success" message={`Alarm updated successfully!`} />
        <Response onClose={() => setTestResult(null)} open={testResult !== null} status={testResult?.status ?? 'success'} message={testResult?.message ?? ''} />
        <div className="border border-gray-200 bg-white rounded-lg overflow-hidden">
            {/* Header row */}
            <div className="flex items-center gap-3 px-5 py-4 border-b border-gray-100">
                <h3 className="text-lg font-semibold flex-1">{alarm.name}</h3>
                <label className="flex items-center gap-1.5 text-sm text-text-muted cursor-pointer select-none">
                    Activated
                    <Switch checked={hasAlarmId ? enabledIds.has(alarmId) : false} onCheckedChange={() => hasAlarmId && toggleAlarm(alarmId)}/>
                </label>
                <Pill variant={alarm.status === "ok" ? "lightGreen" : alarm.status === "firing" ? "destructive" : "secondary"}>
                    Status: {alarm.status}
                </Pill>
            </div>

            {/* Meta grid */}
            <div className="grid grid-cols-2 gap-x-8 gap-y-1 px-5 py-4 text-sm border-b border-gray-100">
                <div className="flex justify-between py-1.5">
                    <span className="text-text-muted">Type</span>
                    <span className="text-text-primary">{alarm.type}</span>
                </div>
                <div className="flex justify-between py-1.5">
                    <span className="text-text-muted">Vhost</span>
                    <span className="text-text-primary">{vhost}</span>
                </div>
                <div className="flex justify-between py-1.5">
                    <span className="text-text-muted">Last fired</span>
                    <span className="text-text-primary">
                        {alarm.last_fired
                            ? new Date(alarm.last_fired).toLocaleString('no-NO', { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' })
                            : "Never"}
                    </span>
                </div>
                <div className="flex justify-between py-1.5">
                    <span className="text-text-muted">Threshold</span>
                    <Input type="number" name="threshold" className="w-20 h-6 text-right" value={threshold} onChange={e => setThreshold(e.target.value)} />
                </div>
            </div>

            {/* Message editor */}
            <div className="px-5 py-4">
                <p className="text-md font-medium text-text-primary mb-1">Custom notification text</p>
                <p className="text-xs text-text-muted mb-2">Override the default message. Leave empty to use the default.</p>
                    <Input name="customMessage" className="flex-1" value={message} onChange={e => setMessage(e.target.value)} placeholder="Enter your custom message..." />
                {hasUnsavedChanges && (
                    <p className="text-xs text-amber-600 mt-2">You have unsaved changes. Click <strong>Save</strong> to apply them.</p>
                )}
                <div className="flex mt-2 justify-end gap-2"> 
                    <Button variant="orange" size="sm" onClick={() => updateAlarm(alarm.id!)}>Save</Button>
                    <div className="flex">
                        <Button variant="outline" size="sm" className="rounded-r-none border-r-0" onClick={testNotification}>Send test notification</Button>
                        <DropdownMenu.Root>
                            <DropdownMenu.Trigger asChild>
                                <Button variant="outline" size="sm" className="rounded-l-none px-2.5 tracking-widest">···</Button>
                            </DropdownMenu.Trigger>
                            <DropdownMenu.Portal>
                                <DropdownMenu.Content
                                    align="end"
                                    sideOffset={4}
                                    className="z-50 min-w-[10rem] overflow-hidden rounded-md border border-gray-200 bg-white shadow-md text-sm text-text-primary p-1"
                                    >
                                    {secondaryActions.map((action, index) => (
                                        <DropdownMenu.Item
                                        key={index}
                                        className={`flex cursor-pointer select-none items-center rounded px-3 py-2 outline-none hover:bg-gray-100 ${action.label === "Delete alarm" ? "text-destructive hover:bg-red-50" : ""}`}
                                        onSelect={action.onClick}
                                        >
                                            {action.label}
                                        </DropdownMenu.Item>
                                    ))}
                                </DropdownMenu.Content>
                            </DropdownMenu.Portal>
                        </DropdownMenu.Root>
                    </div>
                </div>
            </div>
        </div>
    </div>
    )
}