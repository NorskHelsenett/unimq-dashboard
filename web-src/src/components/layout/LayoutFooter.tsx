import { useCheckerStatus } from '@/hooks/useCheckerStatus'

export function LayoutFooter() {
  const { status } = useCheckerStatus()

  const isFresh =
    status?.last_checked != null &&
    Date.now() - new Date(status.last_checked).getTime() < status.interval_s * 1500

  const lastUpdated = status?.last_checked
    ? new Date(status.last_checked).toLocaleTimeString('en-GB')
    : null

  return (
    <footer className="flex items-center justify-between px-6 py-2 border-t border-border-card text-xs text-text-muted">
      <span className="w-32" />
      <span className="flex items-center gap-2">
        <span className={`w-1.5 h-1.5 rounded-full shrink-0 ${isFresh ? 'bg-green-500' : 'bg-text-muted'}`} />
        {isFresh ? 'All data is live and auto-refreshing' : 'Awaiting data refresh…'}
      </span>
      <span className="w-32 text-right">
        {lastUpdated && <>Last updated: {lastUpdated}</>}
      </span>
    </footer>
  )
}
