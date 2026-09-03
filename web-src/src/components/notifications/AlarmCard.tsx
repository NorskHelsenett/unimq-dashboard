import React, { useState } from "react"
import { Input } from "../ui/input"
import { Selector, SelectorTrigger, SelectorContent, SelectorItem, SelectorValue, SelectLabel } from "../ui/selector"
import { Button } from "../ui/button"
import { Switch } from "../ui/switch"
import { DeleteItem } from "./DeleteItem"
import { AlarmProps, alarmDropdownOptions } from '@/types/notifications'
import { toggleRule, addRule } from '@/services/notifications'
import { SectionCard, SectionCardHeader } from "../ui/section-card"
import { StatusDot } from "../ui/status-dot"
import { Pill } from "../ui/pill"
import { Bell, Pencil } from "lucide-react"
import { Tooltip, TooltipContent, TooltipTrigger } from "../ui/tooltip"
import { cn } from "@/lib/utils"

function ExistingAlarms({existingAlarms, vhost, disabledIds, onToggle}: {
    existingAlarms: AlarmProps[]
    vhost: string
    disabledIds: Set<string>
    onToggle: (id: string) => void
}) {
    const alarms = existingAlarms || []
    const [deletingId, setDeletingId] = useState<string | null>(null)
    const deletingAlarm = alarms.find(a => a.id === deletingId)
    const [showDeactivated, setShowDeactivated] = useState(false)
    const sortedAlarms = [...alarms].sort((a, b) => {
        const aDisabled = disabledIds.has(a.id!) ? 1 : 0
        const bDisabled = disabledIds.has(b.id!) ? 1 : 0
        return aDisabled - bDisabled
    })
    const visibleAlarms = showDeactivated ? sortedAlarms : sortedAlarms.filter(a => !disabledIds.has(a.id!))
    const deactivatedCount = alarms.filter(a => disabledIds.has(a.id!)).length

    const toggleAlarm = (id: string) => {
        onToggle(id)
        toggleRule(vhost, id)
    }

    return (
        <div className="mt-2">
            <DeleteItem alarm={deletingAlarm} vhost={vhost} open={deletingId !== null} onClose={() => setDeletingId(null)} />
            {alarms.length === 0 ? (
                <p className="text-sm text-text-muted py-2">No alarms configured.</p>
            ) : (
                <>
                    {visibleAlarms.length > 0 ? (
                        <div className="overflow-y-auto max-h-64">
                            <table className="w-full text-left border-collapse">
                                <thead>
                                    <tr>
                                        <th className="border-b border-border-card py-2 px-4 text-xs text-text-muted">Alarm</th>
                                        <th className="border-b border-border-card py-2 px-4 text-xs text-text-muted">Last fired</th>
                                        <th className="border-b border-border-card py-2 pl-1 pr-2 text-xs text-text-muted">Status</th>
                                        <th className="border-b border-border-card py-2 pl-1 pr-1 text-xs text-text-muted">Active</th>
                                        <th className="border-b border-border-card py-2 pl-6 pr-4 text-xs text-text-muted text-right">Actions</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {visibleAlarms.map((alarm, i) => {
                                        const isDisabled = disabledIds.has(alarm.id!)
                                        const prevIsDisabled = i > 0 && disabledIds.has(visibleAlarms[i - 1].id!)
                                        const showDivider = isDisabled && !prevIsDisabled && i > 0
                                        const pillVariant =
                                            alarm.status === 'firing'  ? 'red' :
                                            alarm.status === 'fired'   ? 'amber' :
                                            alarm.status === 'ok'      ? 'lightGreen' :
                                            alarm.status === 'inactive' ? 'gray' :
                                            'lightBlue' // active / unknown — pending first check
                                        const fadedCell = isDisabled ? 'opacity-40 grayscale' : ''
                                        return (
                                            <React.Fragment key={alarm.id}>
                                            {showDivider && (
                                                <tr key={`divider-${i}`}>
                                                    <td colSpan={5} className="py-1 px-4">
                                                        <div className="flex items-center gap-2 text-xs text-text-muted">
                                                            <div className="flex-1 border-t border-dashed border-border-card" />
                                                            <span>Deactivated</span>
                                                            <div className="flex-1 border-t border-dashed border-border-card" />
                                                        </div>
                                                    </td>
                                                </tr>
                                            )}
                                            <tr
                                                className={cn('cursor-pointer',
                                                    isDisabled ? '' :
                                                    alarm.status === 'firing' ? 'bg-red-50 hover:bg-red-100' :
                                                    'hover:bg-surface-page'
                                                )}
                                                onClick={() => { window.location.href = `/notifications/rule?vhost=${encodeURIComponent(vhost)}&id=${alarm.id}` }}
                                            >
                                                <td className={cn("border-b border-border-card py-2 px-4 text-sm font-medium", fadedCell)}>
                                                    <span className="flex items-center gap-2">
                                                        <StatusDot color={alarm.status === 'firing' ? 'danger' : alarm.status === 'fired' ? 'warning' : alarm.status === 'ok' ? 'ok' : 'warning'} />
                                                        <div>
                                                            <p className="text-sm font-medium text-text-primary">{alarm.name}</p>
                                                            <p className="text-xs text-text-muted">
                                                                {alarm.type}{alarm.threshold != null ? ` · ${alarm.threshold}` : ''}
                                                            </p>
                                                            {alarm.status === 'firing' && alarm.last_value != null && (
                                                                <p className="text-xs font-medium text-status-danger">⚠ {alarm.last_value}</p>
                                                            )}
                                                            {alarm.status === 'ok' && alarm.last_value != null && (
                                                                <p className="text-xs text-status-ok">✓ {alarm.last_value}</p>
                                                            )}
                                                        </div>
                                                    </span>
                                                </td>
                                                <td className={cn("border-b border-border-card py-2 px-4 text-sm", fadedCell)}>
                                                    {alarm.last_fired
                                                        ? new Date(alarm.last_fired).toLocaleString('no-NO', { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' })
                                                        : <span className="text-text-muted italic text-xs">Never</span>}
                                                </td>
                                                <td className={cn("border-b border-border-card py-2 pl-1 pr-2", fadedCell)}>
                                                    <Pill variant={pillVariant} className="border-none px-2 text-xs">
                                                        {alarm.status === 'firing'   ? 'Firing' :
                                                         alarm.status === 'fired'    ? 'Fired' :
                                                         alarm.status === 'ok'       ? 'Healthy' :
                                                         alarm.status === 'inactive' ? 'Inactive' :
                                                         alarm.status === 'active'   ? 'Pending' :
                                                         alarm.status === 'unknown'  ? 'Pending' :
                                                         'Unknown'}
                                                    </Pill>
                                                </td>
                                                <td className="border-b border-border-card py-2 pl-1 pr-1">
                                                    <Switch
                                                        checked={!isDisabled}
                                                        onCheckedChange={() => alarm.id && toggleAlarm(alarm.id)}
                                                    />
                                                </td>
                                                <td className="border-b border-border-card py-2 pl-6 pr-4" onClick={e => e.stopPropagation()}>
                                                    <div className="flex items-center justify-end gap-2">
                                                        <Tooltip>
                                                            <TooltipTrigger asChild>
                                                                <Button
                                                                    variant="ghost"
                                                                    size="sm"
                                                                    className="p-0 text-text-muted hover:text-text-secondary"
                                                                    onClick={() => { window.location.href = `/notifications/rule?vhost=${encodeURIComponent(vhost)}&id=${alarm.id}` }}
                                                                >
                                                                    <Pencil className="h-4 w-4" />
                                                                </Button>
                                                            </TooltipTrigger>
                                                            <TooltipContent>Edit</TooltipContent>
                                                        </Tooltip>
                                                        <Button variant="destructive" size="xs" onClick={() => alarm.id && setDeletingId(alarm.id)}>Delete</Button>
                                                    </div>
                                                </td>
                                            </tr>
                                            </React.Fragment>
                                        )
                                    })}
                                </tbody>
                            </table>
                        </div>
                    ) : (
                        <p className="text-sm text-text-muted py-2">No active alarms.</p>
                    )}
                    {deactivatedCount > 0 && (
                        <div className="flex justify-center mt-2">
                            <button
                                onClick={() => setShowDeactivated(prev => !prev)}
                                className="text-xs text-text-muted hover:text-text-primary transition-colors"
                            >
                                {showDeactivated ? 'Hide deactivated alarms ↑' : `Show ${deactivatedCount} deactivated alarms ↓`}
                            </button>
                        </div>
                    )}
                </>
            )}
        </div>
    )
}


function AddAlarmForm({ selectedAlarm, vhost, onClose }: { selectedAlarm: string, vhost: string, onClose: () => void }) {
    if (!selectedAlarm) return null

    const hasQueue = ['queue_messages', 'queue_size', 'no_consumer'].includes(selectedAlarm)
    const hasThreshold = !['maintenance', 'no_consumer'].includes(selectedAlarm)
    const label = alarmDropdownOptions.find(o => o.value === selectedAlarm)?.label ?? selectedAlarm

    return (
        <form onSubmit={(e) => {
            e.preventDefault()
            const fd = new FormData(e.currentTarget)
            addRule(vhost, {
                name: fd.get('name') as string,
                type: selectedAlarm,
                queue_name: (fd.get('queue_name') as string) || undefined,
                threshold: fd.get('threshold') ? Number(fd.get('threshold')) : undefined,
                message: (fd.get('message') as string) || '',
            }).then(() => window.location.reload())
        }} className="bg-surface-card border border-blue-200 rounded-lg p-4 mb-4 shadow">
            <div className="flex flex-wrap items-end gap-2">
                <div className="flex flex-col gap-1 min-w-[180px] flex-1">
                    <label className="text-sm font-medium">{label}</label>
                    <Input name="name" placeholder="Alarm name" required />
                </div>
                {hasQueue && (
                    <div className="flex flex-col gap-1">
                        <label className="text-xs text-text-muted">Queue name</label>
                        <Input name="queue_name" placeholder="my.queue.name" required />
                    </div>
                )}
                {hasThreshold && (
                    <div className="flex flex-col gap-1">
                        <label className="text-xs text-text-muted">Threshold</label>
                        <Input name="threshold" type="number" placeholder="e.g. 1000" required />
                    </div>
                )}
                <Button type="submit" variant="orange" size="sm">Save</Button>
                <Button type="button" variant="ghost" size="sm" onClick={onClose}>Cancel</Button>
            </div>
            <div className="mt-2 flex flex-col gap-1">
                <label className="text-xs text-text-muted">Message (optional)</label>
                <Input name="message" placeholder="Leave empty for default" />
            </div>
        </form>
    )
}



export function AlarmCard({ existingAlarms, vhost = '' }: { existingAlarms: AlarmProps[], vhost?: string }) {
    const alarms = existingAlarms || []
    const [userSelectedAlarm, setUserSelectedAlarm] = useState('')
    const [disabledIds, setDisabledIds] = useState<Set<string>>(
        () => new Set(alarms.filter(a => !a.enabled && a.id).map(a => a.id!))
    )

    const toggleDisabled = (id: string) => {
        setDisabledIds(prev => {
            const next = new Set(prev)
            next.has(id) ? next.delete(id) : next.add(id)
            return next
        })
    }

    const hasFiring = alarms.some(a => a.status === 'firing' && !disabledIds.has(a.id!))
    const iconColor = hasFiring ? 'text-status-danger' : 'text-blue-400'

    return (
        <SectionCard accent={hasFiring ? 'danger' : 'blue'}>
            <SectionCardHeader
                title="Alarms"
                icon={<Bell className={`w-5 h-5 ${iconColor}`} />}
                action={
                    <Selector value={userSelectedAlarm} onValueChange={(val) => setUserSelectedAlarm(val)}>
                        <SelectorTrigger className="w-60 ml-3">
                            <SelectorValue placeholder="Add new alarm" />
                        </SelectorTrigger>
                        <SelectorContent>
                            <SelectLabel>Select type</SelectLabel>
                            {alarmDropdownOptions.map(option => (
                                <SelectorItem key={option.value} value={option.value}>
                                    {option.label}
                                </SelectorItem>
                            ))}
                        </SelectorContent>
                    </Selector>
                }
            />
            <p className="text-sm text-text-secondary mb-4 max-w-xl">
                Set up alarms to notify the team when thresholds are reached.
                Alerts are sent only once per trigger and reset automatically when the value goes back below the threshold.
            </p>
            <AddAlarmForm selectedAlarm={userSelectedAlarm} vhost={vhost} onClose={() => setUserSelectedAlarm('')} />
            {userSelectedAlarm && <hr className="border-border-card mb-4" />}
            <ExistingAlarms existingAlarms={existingAlarms} vhost={vhost} disabledIds={disabledIds} onToggle={toggleDisabled} />
        </SectionCard>
    )
}