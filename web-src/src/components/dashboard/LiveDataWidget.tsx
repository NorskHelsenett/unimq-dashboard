import { cn } from '@/lib/utils'
import { useCheckerStatus } from '@/hooks/useCheckerStatus'

interface LiveDataWidgetProps {
  vhost?: string
  className?: string
}

export function LiveDataWidget({ vhost, className }: LiveDataWidgetProps) {
  const { status, timeAgo } = useCheckerStatus()

  const isFresh =
    status?.last_checked != null &&
    Date.now() - new Date(status.last_checked).getTime() < status.interval_s * 1500

  return (
    <div className={cn('flex items-center gap-2 mt-1', className)}>
      {/* Dark status chip */}
      <div className="inline-flex items-center gap-2 bg-gray-900 rounded-lg px-3 py-1.5 shadow-sm">
        <span
          className={cn(
            'w-1.5 h-1.5 rounded-full shrink-0',
            !status
              ? 'bg-gray-600'
              : isFresh
              ? 'bg-green-400 animate-pulse'
              : 'bg-yellow-400',
          )}
        />
        <span className="text-xs text-gray-300">
          {!status ? (
            <span className="text-gray-500">Awaiting first check…</span>
          ) : (
            <>
              Synced{' '}
              <span className="text-white font-medium">{timeAgo}</span>
            </>
          )}
        </span>
        {status?.runtime_ms != null && (
          <>
            <div className="w-px h-3 bg-gray-700 shrink-0" />
            <span className="text-xs font-mono text-green-400">
              {status.runtime_ms}ms
            </span>
          </>
        )}
      </div>

      {/* Vhost pill */}
      {vhost && (
        <span className="font-mono font-medium text-text-secondary bg-white/80 px-1.5 py-0.5 rounded text-xs border border-gray-200">
          {vhost}
        </span>
      )}
    </div>
  )
}
