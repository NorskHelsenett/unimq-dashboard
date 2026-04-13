import { useState } from "react"
import { Input } from "../ui/input"
import { Selector, SelectorTrigger, SelectorContent, SelectorItem, SelectorValue, SelectLabel } from "../ui/selector"
import { Button } from "../ui/button"
import { Switch } from "../ui/switch"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter, DialogClose } from "../ui/dialog"

export interface AlarmProps {
    id?: string
    name?: string
    type?: string    
    queue_name?: string    
    threshold?: number
    message?: string
    enabled?: boolean
    status?: string //"firing", "ok", also based on enabled status
    last_fired?: string | null
    last_value?: number | null
}


const alarmDropdownOptions = [
    { value: 'connections', label: 'Connections' },
    { value: 'channels', label: 'Channels' },
    { value: 'queues', label: 'Queues' },
    { value: 'unacked', label: 'Unacknowledged Messages' },
    { value: 'queue_messages', label: 'Messages in Queue' },
    { value: 'queue_size', label: 'Queue Size' },
    { value: 'no_consumers', label: 'No Consumers' },
    { value: 'maintenance', label: 'Maintenance Message' },
]

// name, type, threshold, message, enabled, status, last_value
function ExistingAlarms({existingAlarms, vhost}: {existingAlarms: AlarmProps[], vhost: string}) {

    const alarms = existingAlarms || []
    const [disabledIds, setDisabledIds] = useState<Set<string>>(
        () => new Set(alarms.filter(a => !a.enabled && a.id).map(a => a.id!))
    )
    const [deletingId, setDeletingId] = useState<string | null>(null)
    const deletingAlarm = alarms.find(a => a.id === deletingId)
    const [showAll, setShowAll] = useState(false)
    const [showDeactivated, setShowDeactivated] = useState(false)
    const sortedAlarms = [...alarms].sort((a, b) => {
        const aDisabled = disabledIds.has(a.id!) ? 1 : 0
        const bDisabled = disabledIds.has(b.id!) ? 1 : 0
        return aDisabled - bDisabled
    })
    const visibleAlarms = showDeactivated ? sortedAlarms : sortedAlarms.filter(a => !disabledIds.has(a.id!))
    const deactivatedCount = alarms.filter(a => disabledIds.has(a.id!)).length

    const toggleAlarm = (id: string) => {
        setDisabledIds(prev => {
            const next = new Set(prev)
            next.has(id) ? next.delete(id) : next.add(id)
            return next
        })
    }

    return (
        <div className="mt-4">
            <Dialog open={deletingId !== null} onOpenChange={(open) => !open && setDeletingId(null)}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Delete alarm</DialogTitle>
                        <DialogDescription>
                            Are you sure you want to delete <span className="font-medium text-text-primary">{deletingAlarm?.name}</span>? This cannot be undone.
                        </DialogDescription>
                    </DialogHeader>
                    <DialogFooter>
                        <DialogClose asChild>
                            <Button variant="outline" className="bg-gray-100">Cancel</Button>
                        </DialogClose>
                        <Button variant="destructive" onClick={() => setDeletingId(null) /* TODO: call backend */}>
                            Delete
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
            {visibleAlarms.length > 0 ? (
                <div className="flex flex-col divide-y divide-gray-100 border border-gray-200 rounded-lg overflow-hidden">
                    {visibleAlarms.map(alarm => {
                        const isDisabled = disabledIds.has(alarm.id!)
                        return (
                        <div key={alarm.id} className="relative grid grid-cols-[1fr_14rem_auto] items-center gap-2 px-4 py-3 bg-white hover:bg-gray-50">
                            {isDisabled && (
                                <div className="absolute inset-0 bg-white/60 backdrop-grayscale pointer-events-none" />
                            )}
                            <div className={`flex items-center gap-3 transition-opacity ${isDisabled ? 'opacity-40 grayscale' : ''}`}>
                                <span className="relative flex size-3 shrink-0 items-center justify-center">
                                    <span className={`absolute inline-flex size-2 animate-ping rounded-full opacity-50 ${alarm.status === 'firing' ? 'bg-status-danger' : alarm.status === 'ok' ? 'bg-status-ok' : 'bg-status-warning'}`} />
                                    <span className={`relative inline-flex size-2 rounded-full ${alarm.status === 'firing' ? 'bg-status-danger' : alarm.status === 'ok' ? 'bg-status-ok' : 'bg-status-warning'}`} />
                                </span>
                                <div>
                                    <p className="text-sm font-medium text-text-primary">{alarm.name}</p>
                                    <p className="text-xs text-text-muted">{alarm.type} · Threshold: {alarm.threshold}</p>
                                    {alarm.status === 'firing' && alarm.last_value != null && (
                                        <p className="text-xs font-medium text-status-danger">⚠ Value: {alarm.last_value}</p>
                                    )}
                                    {alarm.status === 'ok' && alarm.last_value != null && (
                                        <p className="text-xs text-status-ok">✓ Value: {alarm.last_value}</p>
                                    )}
                                    {alarm.status === 'unknown' && (
                                        <p className="text-xs text-status-warning italic">Waiting for first check…</p>
                                    )}
                                </div>
                            </div>
                            <div className={`transition-opacity ${isDisabled ? 'opacity-40 grayscale' : ''}`}>
                                <p className="text-xs text-text-muted">Last fired</p>
                                <p className="text-xs font-medium text-text-secondary">
                                    {alarm.last_fired
                                        ? new Date(alarm.last_fired).toLocaleString('no-NO', { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' })
                                        : <span className="text-text-muted italic">Never</span>}
                                </p>
                            </div>
                            <div className="flex items-center gap-2 relative z-10">
                                <label className="flex items-center gap-2 cursor-pointer select-none">
                                    <span className="text-xs text-text-muted">Activated</span>
                                    <Switch
                                        checked={!isDisabled}
                                        onCheckedChange={() => alarm.id && toggleAlarm(alarm.id)}
                                    />
                                </label>
                                <Button variant="outline" size="xs" onClick={() => { window.location.href = `/notifications/rule?vhost=${encodeURIComponent(vhost)}&id=${alarm.id}` }}>Edit</Button>
                                <Button variant="destructive" size="xs" onClick={() => alarm.id && setDeletingId(alarm.id)}>Delete</Button>
                            </div>
                        </div>
                        )
                    })}
                </div>
            ) : (
                <p className="text-sm text-text-muted py-2">
                    {alarms.length === 0 ? 'No alarms configured.' : 'No active alarms.'}
                </p>
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
        </div>
    )
}


function AddAlarmForm({ selectedAlarm, onClose }: { selectedAlarm: string, onClose: () => void }) {
    if (!selectedAlarm) return null

    return(
        <div className="bg-gray-50 border border-gray-300 rounded-lg p-4">
            <h2 className="text-sm font-semibold mb-3">New alarm: {alarmDropdownOptions.find(o => o.value === selectedAlarm)?.label}</h2>

            {selectedAlarm === 'queue_messages' ? (
                <div className="grid grid-cols-3 gap-4 mb-3">
                    <div>
                        <p className="text-sm mb-1">Alarm name</p>
                        <Input placeholder="E.g. Queue backlog high" className="bg-white" />
                    </div>
                    <div>
                        <p className="text-sm mb-1">Queue name</p>
                        <Input placeholder="E.g. my.queue.name" className="bg-white" />
                    </div>
                    <div>
                        <p className="text-sm mb-1">Threshold</p>
                        <Input placeholder="E.g. 1000" className="bg-white" />
                    </div>
                </div>
            ) : selectedAlarm === 'maintenance' ? (
                <div className="mb-3">
                    <p className="text-sm mb-1">Alarm name</p>
                    <Input placeholder="E.g. Scheduled downtime" className="bg-white" />
                </div>
            ) : (
                <div className="grid grid-cols-2 gap-4 mb-3">
                    <div>
                        <p className="text-sm mb-1">Alarm name</p>
                        <Input placeholder="E.g. Connection close to limit" className="bg-white" />
                    </div>
                    <div>
                        <p className="text-sm mb-1">Threshold</p>
                        <Input placeholder="E.g. 250" className="bg-white" />
                    </div>
                </div>
            )}

            <div className="mb-3">
                <p className="text-sm mb-1">Notification text (optional)</p>
                <Input placeholder="Leave empty for default" className="bg-white" />
            </div>
            <div className="flex justify-between mt-2">
                <Button>Add alarm</Button>
                <Button variant="outline" onClick={onClose}>Cancel</Button>
            </div>
        </div>
    )
}



export function AlarmCard({ existingAlarms, vhost = '' }: { existingAlarms: AlarmProps[], vhost?: string }) {
    const [userSelectedAlarm, setUserSelectedAlarm] = useState('')

    return (
        <div className="bg-white rounded-lg shadow p-4 min-w-[300px] border border-border-card">
            <div className="flex items-center justify-between mb-2">
                <h3 className="text-lg font-semibold">Alarms</h3>
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
            </div>
            <p className="text-sm text-gray-600 mb-4 max-w-xl">
                Set up alarms to notify the team when thresholds are reached. 
                Alerts are sent only once per trigger and reset automatically when the value goes back below the threshold.
            </p>
            
            <AddAlarmForm selectedAlarm={userSelectedAlarm} onClose={() => setUserSelectedAlarm('')} />
            <ExistingAlarms existingAlarms={existingAlarms} vhost={vhost}/>
        </div>
    )
}