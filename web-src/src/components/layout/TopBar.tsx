import { useVhost } from '@/hooks/useVhost'
import { VhostSelector } from './VhostSelector'
import { LiveDataWidget } from '../dashboard/LiveDataWidget'

export function TopBar() {
  const { vhosts, selected, loading, error } = useVhost()

  return (
    <div className="flex flex-wrap items-center justify-end gap-3 px-6 pt-4 pb-3 border-b border-border-card/70">
      {error && <span role="alert" className="text-xs text-destructive">{error}</span>}
      <VhostSelector
        Vhosts={vhosts}
        Selected={selected}
        placeholder={loading ? 'Loading virtual hosts...' : error ? 'Vhosts unavailable' : 'No accessible vhosts'}
      />
      <LiveDataWidget />
    </div>
  )
}
