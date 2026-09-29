import { useVhost } from '@/hooks/useVhost'
import { VhostSelector } from './VhostSelector'
import { LiveDataWidget } from '../dashboard/LiveDataWidget'

export function TopBar() {
  const { vhosts, selected } = useVhost()

  if (vhosts.length === 0) return null

  return (
    <div className="flex items-center justify-end gap-3 px-6 pt-4 pb-3 border-b border-border-card/70">
      <VhostSelector Vhosts={vhosts} Selected={selected} />
      <LiveDataWidget />
    </div>
  )
}
