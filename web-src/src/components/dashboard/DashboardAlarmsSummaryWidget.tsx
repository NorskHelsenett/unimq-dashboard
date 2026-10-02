import { VhostNotification, AlarmProps, Status } from '@/types/notifications'
import { SectionCard, SectionCardHeader } from '../ui/section-card'
import { Pill } from '../ui/pill'
import { StatusDot } from '../ui/status-dot'
import { cn } from '@/lib/utils'
import { Bell, AlertCircle, CheckCircle2, ArrowRight } from 'lucide-react'

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

function alarmStatusLabel(status: Status | undefined) {
  if (status === 'ok') return 'Healthy'
  if(status === 'firing') return 'Firing'
  if(status === 'fired') return 'Fired'
  if(status === 'active') return 'Active'
  if(status === 'inactive') return 'Inactive'
  return status || 'Unknown'
}

function AlarmStatusBanner({ alarms, expanded }: { alarms: AlarmProps[]; expanded: boolean }) {
  const firingCount = alarms.filter(a => a.status === 'firing').length
  const firedCount = alarms.filter(a => a.status === 'fired').length
  const okCount = alarms.filter(a => a.status === 'ok').length

  if (firingCount > 0) {
    return (
      <div className={cn("flex items-center rounded-lg bg-red-50 dark:bg-red-950/50 border border-red-200 dark:border-red-800/60 mb-2", expanded ? "gap-3 px-3 py-2.5" : "gap-2 px-2.5 py-2")}>
        <div className={cn("rounded-md bg-red-600 flex items-center justify-center shrink-0", expanded ? "w-8 h-8" : "w-7 h-7")}>
          <AlertCircle className={cn("text-white", expanded ? "w-4 h-4" : "w-3.5 h-3.5")} />
        </div>
        <div>
          <p className={cn("font-semibold leading-tight", expanded ? "text-sm" : "text-xs")}>{firingCount} alarm{firingCount !== 1 ? 's' : ''} firing</p>
          <p className={expanded ? "text-xs leading-tight" : "text-[11px] leading-tight"}>Needs your attention</p>
        </div>
      </div>
    )
  }

  if (firedCount > 0) {
    return (
      <div className={cn("flex items-center rounded-lg bg-amber-50 dark:bg-amber-950/40 border border-amber-200 dark:border-amber-800/50 mb-2", expanded ? "gap-3 px-3 py-2.5" : "gap-2 px-2.5 py-2")}>
        <div className={cn("rounded-md bg-amber-500 flex items-center justify-center shrink-0", expanded ? "w-8 h-8" : "w-7 h-7")}>
          <AlertCircle className={cn("text-white", expanded ? "w-4 h-4" : "w-3.5 h-3.5")} />
        </div>
        <div>
          <p className={cn("font-semibold leading-tight", expanded ? "text-sm" : "text-xs")}>{firedCount} alarm{firedCount !== 1 ? 's' : ''} fired</p>
          <p className={expanded ? "text-xs leading-tight" : "text-[11px] leading-tight"}>Check recent activity</p>
        </div>
      </div>
    )
  }

  return (
    <div className={cn("flex items-center rounded-lg bg-green-50 dark:bg-green-950/30 border border-green-200 dark:border-green-800/40 mb-2", expanded ? "gap-3 px-3 py-2.5" : "gap-2 px-2.5 py-2")}>
      <div className={cn("rounded-md bg-green-600 flex items-center justify-center shrink-0", expanded ? "w-8 h-8" : "w-7 h-7")}>
        <CheckCircle2 className={cn("text-white", expanded ? "w-4 h-4" : "w-3.5 h-3.5")} />
      </div>
      <div>
        <p className={cn("font-semibold leading-tight", expanded ? "text-sm" : "text-xs")}>All systems healthy</p>
        <p className={expanded ? "text-xs leading-tight" : "text-[11px] leading-tight"}>{okCount} of {alarms.length} alarms healthy</p>
      </div>
    </div>
  )
}

function AlarmRow({ alarm, vhost, expanded }: { alarm: AlarmProps; vhost: string; expanded: boolean }) {
  const dotColor = alarmDotColor(alarm.status)
  const pillVariant = alarmPillVariant(alarm.status)
  const pulse = alarm.status === 'firing'

  return (
    <a
      href={`/notifications/rule?id=${encodeURIComponent(alarm.id ?? '')}&vhost=${encodeURIComponent(vhost)}`}
      className={cn("flex items-center justify-between px-1 -mx-1 rounded border-b last:border-0 border-border-card [text-decoration:none] hover:bg-surface-page transition-colors", expanded ? "py-1.5" : "py-1")}
    >
      <span className={cn("flex items-center gap-1.5 min-w-0 leading-tight", expanded ? "text-sm" : "text-xs")}>
        <StatusDot color={dotColor} pulse={pulse} className="shrink-0" />
        <span className="truncate font-medium">{alarm.name || alarm.type}</span>
        {alarm.queue_name && (
          <span className={cn("text-text-muted truncate", expanded ? "text-xs" : "text-[11px]")}>({alarm.queue_name})</span>
        )}
      </span>
      <div className="flex items-center gap-1 shrink-0 ml-1.5">
        <Pill className={expanded ? "px-2 py-0.5 text-xs" : "px-1.5 py-0 text-[10px] leading-4"} variant={pillVariant}>{alarmStatusLabel(alarm.status)}</Pill>
      </div>
    </a>
  )
}

export function DashboardAlarmsSummaryWidget({
  notification,
  expanded = false,
}: {
  notification: VhostNotification | null
  expanded?: boolean
}) {
  const alarms = (notification?.Rules ?? []).filter(alarm => alarm.enabled !== false)
  const firingCount = alarms.filter(a => a.status === 'firing').length
  const firedCount = alarms.filter(a => a.status === 'fired').length

  const sorted = [...alarms].sort(
    (a, b) =>
      (STATUS_PRIORITY[a.status ?? ''] ?? 6) - (STATUS_PRIORITY[b.status ?? ''] ?? 6)
  )

  const accent =
    firingCount > 0 ? 'danger' : firedCount > 0 ? 'amber' : 'green'

  return (
    <SectionCard accent={accent} className="min-w-0 h-full flex flex-col">
      <SectionCardHeader
        title="Alarms"
        icon={<Bell className={`w-4 h-4 ${firingCount > 0 ? 'text-red-500' : firedCount > 0 ? 'text-amber-500' : 'text-green-500'}`} />}
      />
      {alarms.length === 0 ? (
        <p className="text-sm text-text-muted">No active alarms.</p>
      ) : (
        <>
          <AlarmStatusBanner alarms={alarms} expanded={expanded} />
          <div>
            {sorted.map(a => (
              <AlarmRow key={a.id} alarm={a} vhost={notification?.Name ?? ''} expanded={expanded} />
            ))}
          </div>
          <p className="mt-auto pt-3">
            <a href={`/notifications?vhost=${encodeURIComponent(notification?.Name ?? '')}`} className={cn("text-submit-button text-xs hover:font-semibold transition-colors inline-flex items-center gap-1 [text-decoration:none] hover:[text-decoration:none]")}>
              View all alarms <ArrowRight className="w-3 h-3 inline ml-1" />
            </a>
          </p>
        </>
      )}
    </SectionCard>
  )
}
