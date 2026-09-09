import { Maintenance } from '@/types/maintenance'
import { SectionCard, SectionCardHeader } from '../ui/section-card'
import { Pill } from '../ui/pill'
import { CalendarClock, CalendarDays, Wrench, ArrowRight } from 'lucide-react'
import { cn } from '@/lib/utils'
import { formatDateRange } from '../maintenance/MaintenanceHistoryCard'
import { Selector, SelectorContent, SelectorItem, SelectorTrigger, SelectorValue } from '../ui/selector'
import { useLocalStorage } from '@/hooks/useLocalStorage'

type MaintenanceRange = 'today' | 'week' | 'month'

const dateKeyInOslo = (date: Date) => {
  const parts = new Intl.DateTimeFormat('en-CA', {
    timeZone: 'Europe/Oslo',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).formatToParts(date).reduce<Record<string, string>>((result, part) => {
    result[part.type] = part.value
    return result
  }, {})
  return `${parts.year}-${parts.month}-${parts.day}`
}

const getRange = (range: MaintenanceRange) => {
  const now = new Date()
  const todayKey = dateKeyInOslo(now)
  const today = new Date(`${todayKey}T00:00:00Z`)

  if (range === 'today') return { start: todayKey, end: todayKey }

  if (range === 'week') {
    const dayOfWeek = today.getUTCDay() || 7
    const weekStart = new Date(today)
    weekStart.setUTCDate(today.getUTCDate() - dayOfWeek + 1)
    const weekEnd = new Date(weekStart)
    weekEnd.setUTCDate(weekStart.getUTCDate() + 6)
    return { start: dateKeyInOslo(weekStart), end: dateKeyInOslo(weekEnd) }
  }

  const monthEnd = new Date(Date.UTC(today.getUTCFullYear(), today.getUTCMonth() + 1, 0))
  return { start: `${todayKey.slice(0, 7)}-01`, end: dateKeyInOslo(monthEnd) }
}

const rangeSelectorLabels: Record<MaintenanceRange, string> = {
  today: 'Today',
  week: 'This week',
  month: 'This month',
}

export function DashboardMaintenanceWidget({ schedule }: { schedule: Maintenance[] }) {
  const [range, setRange] = useLocalStorage<MaintenanceRange>('dashboard-maintenance-range', 'today')
  const selectedRange = getRange(range)
  const filteredSchedule = schedule.filter(maintenance => {
    const start = dateKeyInOslo(new Date(maintenance.start))
    const end = dateKeyInOslo(new Date(maintenance.end))
    return start <= selectedRange.end && end >= selectedRange.start
  })
  const upcoming = [...filteredSchedule]
    .sort((a, b) => {
      if (a.status === 'in_progress' && b.status !== 'in_progress') return -1
      if (b.status === 'in_progress' && a.status !== 'in_progress') return 1
      return new Date(a.start).getTime() - new Date(b.start).getTime()
    })
    .slice(0, 5)
  const hasMoreMaintenance = filteredSchedule.length > upcoming.length

  const hasInProgress = schedule.some(m => m.status === 'in_progress')
  const accent = hasInProgress ? 'amber' : 'blue'

  return (
    <SectionCard accent={accent} className="min-w-0 h-full flex flex-col">
      <SectionCardHeader
        title="Maintenance"
        icon={
          <CalendarClock
            className={`w-4 h-4 ${hasInProgress ? 'text-amber-500' : 'text-blue-400'}`}
          />
        }
        action={
          <Selector value={range} onValueChange={value => setRange(value as MaintenanceRange)}>
            <SelectorTrigger className="w-36 border-border-card bg-surface-page px-3 font-medium text-xs text-text-primary shadow-sm hover:border-blue-300 hover:bg-surface-card focus-visible:border-blue-400 focus-visible:ring-blue-200">
              <span className="flex min-w-0 items-center gap-2">
                <CalendarDays className="h-4 w-4 shrink-0 text-text-muted" />
                <SelectorValue />
              </span>
            </SelectorTrigger>
            <SelectorContent>
              <SelectorItem value="today">{rangeSelectorLabels.today}</SelectorItem>
              <SelectorItem value="week">{rangeSelectorLabels.week}</SelectorItem>
              <SelectorItem value="month">{rangeSelectorLabels.month}</SelectorItem>
            </SelectorContent>
          </Selector>
        }
      />
      {upcoming.length === 0 ? (
        <p className="text-sm text-text-muted">No maintenance in this period.</p>
      ) : (
        <div>
          {upcoming.map(m => {
            const inProgress = m.status === 'in_progress'
            const pill = inProgress
              ? { variant: 'amber' as const, label: 'In progress' }
              : { variant: 'lightBlue' as const, label: 'Scheduled' }
            return (
              <a
                key={m.id}
                href={`/maintenance/edit?id=${encodeURIComponent(m.id ?? '')}`}
                className={cn(
                  'block py-2 px-1 -mx-1 rounded border-b last:border-0 border-border-card [text-decoration:none] hover:bg-surface-page transition-colors',
                  inProgress && 'rounded-lg px-2 -mx-2 mb-1 bg-maintenance-in-progress-bg hover:bg-maintenance-in-progress-bg-hover border-amber-200'
                )}
              >
                <div className="flex items-start justify-between gap-2 mb-0.5">
                  <span className={cn('text-sm leading-tight', inProgress ? 'font-semibold text-amber-900' : 'font-medium')}>
                    {inProgress && <Wrench className="w-3 h-3 inline mr-1.5 text-amber-500 shrink-0" />}
                    {m.description}
                  </span>
                  <Pill variant={pill.variant} className="shrink-0">{pill.label}</Pill>
                </div>
                <p className={cn('text-xs', inProgress ? 'text-amber-600' : 'text-text-muted')}>
                  {formatDateRange(m.start, m.end)}
                </p>
              </a>
            )
          })}
        </div>
      )}
      {hasMoreMaintenance && (
        <p className="pt-3 text-xs text-text-muted">
          Showing {upcoming.length} of {filteredSchedule.length} maintenance entries for {rangeSelectorLabels[range].toLowerCase()}.
        </p>
      )}
      <p className="mt-auto pt-3">
        <a href="/maintenance" className={cn("text-submit-button text-xs mt-2 hover:font-semibold transition-colors inline-flex items-center gap-1 [text-decoration:none] hover:[text-decoration:none]")}>
          View maintenance schedule <ArrowRight className="w-3 h-3 inline ml-1" />
        </a>
      </p>
    </SectionCard>
  )
}
