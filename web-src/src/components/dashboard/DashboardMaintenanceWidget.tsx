import { Maintenance } from '@/types/maintenance'
import { SectionCard, SectionCardHeader } from '../ui/section-card'
import { Pill } from '../ui/pill'
import { CalendarClock, Wrench, ArrowRight } from 'lucide-react'
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
    <SectionCard accent={accent} className="min-w-0 h-full flex flex-col">
      <SectionCardHeader
        title="Maintenance"
        icon={
          <CalendarClock
            className={`w-4 h-4 ${hasInProgress ? 'text-amber-500' : 'text-blue-400'}`}
          />
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
      <p className="mt-auto pt-3">
        <a href="/maintenance" className={cn("text-submit-button text-xs mt-2 hover:font-semibold transition-colors inline-flex items-center gap-1 [text-decoration:none] hover:[text-decoration:none]")}>
          View maintenance schedule <ArrowRight className="w-3 h-3 inline ml-1" />
        </a>
      </p>
    </SectionCard>
  )
}
