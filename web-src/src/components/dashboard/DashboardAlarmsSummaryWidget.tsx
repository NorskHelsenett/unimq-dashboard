import { VhostNotification, AlarmProps, Status } from '@/types/notifications'
import { SectionCard, SectionCardHeader } from '../ui/section-card'
import { Pill } from '../ui/pill'
import { StatusDot } from '../ui/status-dot'
import { cn } from '@/lib/utils'
import { Bell } from 'lucide-react'

const STATUS_PRIORITY: Record<string, number> = {
  firing: 0, fired: 1, active: 2, ok: 3, inactive: 4, unknown: 5, '': 6,
}

function alarmDotColor(status: Status | undefined): 'danger' | 'warning' | 'ok' | 'blue' {
  if (status === 'firing') return 'danger'
  if (status === 'fired') return 'warning'
  if (status === 'ok') return 'ok'
  return 'blue'
}

function alarmPillVariant(status: Status | undefined) {
  if (status === 'firing') return 'red' as const
  if (status === 'fired') return 'amber' as const
  if (status === 'ok') return 'lightGreen' as const
  if (status === 'inactive') return 'gray' as const
  return 'lightBlue' as const
}

function AlarmRow({ alarm }: { alarm: AlarmProps }) {
  const isDisabled = alarm.enabled === false
  const dotColor = alarmDotColor(alarm.status)
  const pillVariant = alarmPillVariant(alarm.status)
  const pulse = alarm.status === 'firing' && !isDisabled

  return (
    <div className={cn(
      "flex items-center justify-between py-1.5 border-b last:border-0 border-border-card",
      isDisabled && "opacity-50"
    )}>
      <span className="flex items-center gap-2 text-sm min-w-0">
        <StatusDot color={isDisabled ? 'blue' : dotColor} pulse={pulse} className="shrink-0" />
        <span className="truncate font-medium">{alarm.name || alarm.type}</span>
        {alarm.queue_name && (
          <span className="text-xs text-text-muted truncate">({alarm.queue_name})</span>
        )}
      </span>
      <div className="flex items-center gap-1.5 shrink-0 ml-2">
        {isDisabled ? (
          <span className="text-[10px] font-medium uppercase tracking-wide text-text-muted bg-gray-100 px-1.5 py-0.5 rounded">
            disabled
          </span>
        ) : (
          <Pill variant={pillVariant}>{alarm.status || 'unknown'}</Pill>
        )}
      </div>
    </div>
  )
}

export function DashboardAlarmsSummaryWidget({
  notification,
}: {
  notification: VhostNotification | null
}) {
  const alarms = notification?.Rules ?? []
  const firingCount = alarms.filter(a => a.status === 'firing').length
  const firedCount = alarms.filter(a => a.status === 'fired').length
  const okCount = alarms.filter(a => a.status === 'ok').length
  const disabledCount = alarms.filter(a => a.enabled === false).length

  const sorted = [...alarms].sort(
    (a, b) =>
      (STATUS_PRIORITY[a.status ?? ''] ?? 6) - (STATUS_PRIORITY[b.status ?? ''] ?? 6)
  )

  const accent =
    firingCount > 0 ? 'danger' : firedCount > 0 ? 'amber' : 'blue'

  return (
    <SectionCard accent={accent} className="min-w-0 h-full">
      <SectionCardHeader
        title="Alarms"
        icon={<Bell className="w-4 h-4 text-blue-400" />}
      />
      {alarms.length === 0 ? (
        <p className="text-sm text-text-muted">No alarms configured.</p>
      ) : (
        <>
          <div className="flex flex-wrap gap-2 mb-3">
            {firingCount > 0 && <Pill variant="red">{firingCount} firing</Pill>}
            {firedCount > 0 && <Pill variant="amber">{firedCount} fired</Pill>}
            {okCount > 0 && <Pill variant="lightGreen">{okCount} ok</Pill>}
            {disabledCount > 0 && <Pill variant="gray">{disabledCount} disabled</Pill>}
            {firingCount === 0 && firedCount === 0 && okCount === 0 && disabledCount === 0 && (
              <Pill variant="lightBlue">{alarms.length} active</Pill>
            )}
          </div>
          <div className="max-h-56 overflow-y-auto">
            {sorted.map(a => (
              <AlarmRow key={a.id} alarm={a} />
            ))}
          </div>
        </>
      )}
    </SectionCard>
  )
}
