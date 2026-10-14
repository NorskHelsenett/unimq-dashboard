import { NodeStats } from '@/types/clusterStats'
import { SectionCard, SectionCardHeader } from '../ui/section-card'
import { GaugeChart } from '../charts/GaugeChart'
import { RatioChart } from '../charts/RatioChart'
import { Server } from 'lucide-react'
import { convertBytes } from '@/lib/bytes'
import { useVhostUsage } from '@/hooks/useVhostUsage'
import { useState } from 'react'

export function DashboardClusterWidget({
  clusters,
  vhost,
}: {
  clusters: NodeStats[] | null
  vhost: string
}) {
  const { usage: vhostResource } = useVhostUsage(vhost)
  const [selectedName, setSelectedName] = useState<string | null>(null)

  if (!clusters) {
    return (
      <SectionCard accent="none" className="min-w-0 h-full">
        <p className="text-sm text-text-muted">Cluster data unavailable.</p>
      </SectionCard>
    )
  }

  const node = clusters.find(n => n.name === selectedName) ?? clusters[0]

  return (
    <SectionCard accent="none" className="min-w-0 h-full">
      <SectionCardHeader
        title="Cluster Resources"
        icon={<Server className="w-4 h-4 text-text-muted" />}
      />
      {clusters.length > 1 && (
        <div role="tablist" className="flex gap-1 overflow-x-auto border-b border-border-card mb-4">
          {clusters.map(n => {
            const active = n.name === node.name
            return (
              <button
                key={n.name}
                type="button"
                role="tab"
                aria-selected={active}
                title={n.name}
                onClick={() => setSelectedName(n.name)}
                className={`shrink-0 max-w-48 truncate px-3 py-1.5 text-xs -mb-px border-b-2 transition-colors ${
                  active
                    ? 'border-text-primary text-text-primary'
                    : 'border-transparent text-text-muted hover:text-text-primary'
                }`}
              >
                {n.name}
              </button>
            )
          })}
        </div>
      )}
      <div className="flex flex-wrap gap-6">
        {!node && (
          <p className="text-sm text-text-muted">No nodes found.</p>
        )}
        {node && (
          <div>
            <p className="text-xs text-text-muted uppercase tracking-wide mb-2">{node.name}</p>
            <div className="flex flex-wrap gap-4">
              <GaugeChart
                title="Memory"
                usage={node.mem_used}
                max={node.mem_limit}
                labelText={`${convertBytes(node.mem_used)}/${convertBytes(node.mem_limit)}`}
                fontSize={14}
              />
              <RatioChart
                title="Disk"
                description={
                  <div className="text-xs text-text-muted space-y-0.5">
                    <p>Lower limit: {convertBytes(node.disk_free_limit)}</p>
                    <p>Free: {convertBytes(node.disk_free)}</p>
                  </div>
                }
                free={node.disk_free}
                limit={node.disk_free_limit}
              />
            </div>
          </div>
        )}
        <div>
          <p className="text-xs text-text-muted uppercase tracking-wide mb-2">Vhost</p>
          <div className="border border-border-card rounded-md divide-y divide-border-card text-sm">
            <div className="flex justify-between gap-8 px-3 py-2">
              <span className="text-text-muted">Messages (memory)</span>
              <span className="font-mono">{convertBytes(vhostResource?.message_bytes ?? 0)}</span>
            </div>
            <div className="flex justify-between gap-8 px-3 py-2">
              <span className="text-text-muted">Disk (persistent)</span>
              <span className="font-mono">{convertBytes(vhostResource?.disk_bytes ?? 0)}</span>
            </div>
          </div>
        </div>
      </div>
    </SectionCard>
  )
}
