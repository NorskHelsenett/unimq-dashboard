import { cn } from '@/lib/utils'
import { useCheckerStatus } from '@/hooks/useCheckerStatus'

interface LiveDataWidgetProps {
  className?: string
}

export function LiveDataWidget({ className }: LiveDataWidgetProps) {
  const { status, timeAgo } = useCheckerStatus()

  const isFresh =
    status?.last_checked != null &&
    Date.now() - new Date(status.last_checked).getTime() < status.interval_s * 1500

  return (
    <div className={cn('flex items-center gap-2 mt-1', className)}>
      {/* Light status chip */}
      <div className="inline-flex items-center gap-2 bg-surface-card border border-border-card rounded-lg px-3 py-1.5 shadow-sm">
        <span
          className={cn(
            'w-2 h-2 rounded-full shrink-0 self-start mt-[4px]',
            !status
              ? 'bg-text-muted'
              : isFresh
              ? 'bg-green-500 animate-pulse'
              : 'bg-yellow-400',
          )}
        />
        <span className="text-xs text-text-muted">
          {!status ? (
            <span className="text-text-muted">Awaiting first check…</span>
          ) : (
            <span className="flex flex-col leading-tight">
              <span>Synced</span>
              <span className="text-text-primary font-medium mt-1">{timeAgo}</span>
            </span>
          )}
        </span>
        {status?.runtime_ms != null && (
          <>
            <div className="w-px h-3 bg-border-card shrink-0" />
            <span className="text-xs font-mono text-green-600">
              {status.runtime_ms}ms
            </span>
          </>
        )}
      </div>
    </div>
  )
}
