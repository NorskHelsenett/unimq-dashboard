import { Maintenance } from '@/types/maintenance'
import { SectionCard, SectionCardHeader } from '../ui/section-card'
import { Pill } from '../ui/pill'
import { CalendarClock, Wrench } from 'lucide-react'
import { cn } from '@/lib/utils'
import { formatDateRange } from '../maintenance/MaintenanceHistoryCard'

export function DashboardMaintenanceWidget({ schedule }: { schedule: Maintenance[] }) {
  const upcoming = [...schedule]
    .sort((a, b) => {
      if (a.status === 'in_progress' && b.status !== 'in_progress') return -1
      if (b.status === 'in_progress' && a.status !== 'in_progress') return 1
      return new Date(a.start).getTime() - new Date(b.start).getTime()
    })
    .slice(0, 5)

  const hasInProgress = schedule.some(m => m.status === 'in_progress')
  const accent = hasInProgress ? 'amber' : 'blue'

  return (
    <SectionCard accent={accent} className="min-w-0 h-full">
      <SectionCardHeader
        title="Maintenance"
        icon={
          <CalendarClock
            className={`w-4 h-4 ${hasInProgress ? 'text-amber-500' : 'text-blue-400'}`}
          />
        }
        action={
          hasInProgress ? (
            <span className="flex items-center gap-1.5 text-xs font-semibold text-amber-600 bg-amber-50 border border-amber-200 px-2 py-0.5 rounded-full">
              <span className="w-1.5 h-1.5 rounded-full bg-amber-500 animate-pulse shrink-0" />
              Live
            </span>
          ) : undefined
        }
      />
      {upcoming.length === 0 ? (
        <p className="text-sm text-text-muted">No maintenance scheduled.</p>
      ) : (
        <div>
          {upcoming.map(m => {
            const inProgress = m.status === 'in_progress'
            const pill = inProgress
              ? { variant: 'amber' as const, label: 'In progress' }
              : { variant: 'lightBlue' as const, label: 'Scheduled' }
            return (
              <div
                key={m.id}
                className={cn(
                  'py-2 border-b last:border-0 border-border-card',
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
              </div>
            )
          })}
        </div>
      )}
    </SectionCard>
  )
}
