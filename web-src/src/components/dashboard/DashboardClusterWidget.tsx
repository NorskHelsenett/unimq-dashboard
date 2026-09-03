import { ClusterStats } from '@/types/clusterStats'
import { SectionCard, SectionCardHeader } from '../ui/section-card'
import { GaugeChart } from '../charts/GaugeChart'
import { RatioChart } from '../charts/RatioChart'
import { Server } from 'lucide-react'
import { convertBytes } from '@/lib/bytes'

export function DashboardClusterWidget({
  clusters,
  vhost,
}: {
  clusters: ClusterStats | null
  vhost: string
}) {
  if (!clusters) {
    return (
      <SectionCard accent="none" className="min-w-0 h-full">
        <p className="text-sm text-text-muted">Cluster data unavailable.</p>
      </SectionCard>
    )
  }

  const { total_mem_used, total_mem_limit, total_disk_free, min_disk_limit } = clusters
  const vhostResource = (clusters.vhost_resources ?? []).find(v => v.name === vhost)

  return (
    <SectionCard accent="none" className="min-w-0 h-full">
      <SectionCardHeader
        title="Cluster Resources"
        icon={<Server className="w-4 h-4 text-text-muted" />}
      />
      <div className="flex flex-wrap gap-6">
        <div>
          <p className="text-xs text-text-muted uppercase tracking-wide mb-2">Cluster</p>
          <div className="flex flex-wrap gap-4">
            <GaugeChart
              title="Memory"
              usage={total_mem_used}
              max={total_mem_limit}
              labelText={`${convertBytes(total_mem_used)}/${convertBytes(total_mem_limit)}`}
              fontSize={14}
            />
            <RatioChart
              title="Disk"
              description={
                <div className="text-xs text-text-muted space-y-0.5">
                  <p>Lower limit: {convertBytes(min_disk_limit)}</p>
                  <p>Free: {convertBytes(total_disk_free)}</p>
                </div>
              }
              free={total_disk_free}
              limit={min_disk_limit}
            />
          </div>
        </div>
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
