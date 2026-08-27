import type { ReactNode } from 'react'
import { cn } from '@/lib/utils'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { Network, GitBranch, LayoutList, Hourglass } from 'lucide-react'

interface LimitsCardProps {
  connections: number
  channels: number
  queues: number
  unacked: number
  maxConnections: number
  maxQueues: number
}

type StatusKey = 'danger' | 'warning' | 'neutral'

function limitColor(value: number, max: number): StatusKey {
  if (max <= 0) return 'neutral'
  if (value >= max * 0.8) return 'danger'
  if (value >= max * 0.5) return 'warning'
  return 'neutral'
}

const valueColor: Record<StatusKey, string> = {
  danger:  'text-red-600',
  warning: 'text-amber-600',
  neutral: 'text-text-primary',
}

interface StatCardProps {
  label: string
  tooltip: string
  value: number
  sub: string
  max?: number
  icon: ReactNode
  cardBg: string
  cardBorder: string
  iconBg: string
  iconColor: string
  subColor: string
}

function StatCard({ label, tooltip, value, sub, max, icon, cardBg, cardBorder, iconBg, iconColor, subColor }: StatCardProps) {
  const colorKey: StatusKey = max ? limitColor(value, max) : 'neutral'
  return (
    <div className={cn('rounded-xl border p-4', cardBg, cardBorder)}>
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-6">
          <div className={cn('w-9 h-9 rounded-lg flex items-center justify-center shrink-0', iconBg)}>
            <div className={iconColor}>{icon}</div>
          </div>
          <div className="flex flex-col gap-0.5 min-w-0">
            <span className="text-sm font-medium text-text-secondary">{label}</span>
            <div className={cn('text-2xl font-bold tabular-nums tracking-tight', valueColor[colorKey])}>
              {value}
            </div>
          </div>
        </div>
        <Tooltip>
          <TooltipTrigger asChild>
            <button
              type="button"
              aria-label={`Info about ${label}`}
              className="inline-flex items-center justify-center w-3.5 h-3.5 rounded-full bg-surface-card/70 text-text-muted text-[14px] cursor-default select-none"
            >
              ?
            </button>
          </TooltipTrigger>
          <TooltipContent>
            <p className="max-w-xs text-xs">{tooltip}</p>
          </TooltipContent>
        </Tooltip>
      </div>
      <div className="flex items-end justify-between gap-2">
        <div>
          <div className={cn('text-xs mt-0.5 font-medium', subColor)}>{sub}</div>
        </div>
      </div>
    </div>
  )
}

export function LimitsCard({ connections, channels, queues, unacked, maxConnections, maxQueues }: LimitsCardProps) {
  return (
    <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
      <StatCard
        label="Connections"
        tooltip={`Maks antall connections til en vhost er ${maxConnections}. Etter at dette antallet er nådd, vil det ikke lenger være mulig å opprette nye connections før du er under grensen igjen.`}
        value={connections}
        sub={`limit ${maxConnections}`}
        max={maxConnections > 0 ? maxConnections : undefined}
        icon={<Network className="w-4 h-4" />}
        cardBg="bg-surface-card" cardBorder="border-border-card"
        iconBg="bg-blue-100" iconColor="text-blue-500" subColor="text-blue-400"
      />
      <StatCard
        label="Channels"
        tooltip="Vi anbefaler å holde antallet channels per vhost under 1000."
        value={channels}
        sub="rec. <1000"
        icon={<GitBranch className="w-4 h-4" />}
        cardBg="bg-surface-card" cardBorder="border-border-card"
        iconBg="bg-violet-100" iconColor="text-violet-500" subColor="text-violet-400"
      />
      <StatCard
        label="Queues"
        tooltip={`Maks antall queues på en vhost er ${maxQueues}. Etter at dette antallet er nådd, vil det ikke lenger være mulig å opprette nye queues før du er under grensen igjen.`}
        value={queues}
        sub={`limit ${maxQueues}`}
        max={maxQueues > 0 ? maxQueues : undefined}
        icon={<LayoutList className="w-4 h-4" />}
        cardBg="bg-surface-card" cardBorder="border-border-card"
        iconBg="bg-orange-100" iconColor="text-orange-500" subColor="text-orange-400"
      />
      <StatCard
        label="Unacked messages"
        tooltip="Totalt antall unacked messages på vhosten. Meldinger som hentes, men ikke enda er konsumert, havner i «unacked state». Disse meldingene lagres i minne, det er derfor ikke ønskelig å ha for mange."
        value={unacked}
        sub="keep low"
        icon={<Hourglass className="w-4 h-4" />}
        cardBg="bg-surface-card" cardBorder="border-border-card"
        iconBg="bg-teal-100" iconColor="text-teal-500" subColor="text-teal-400"
      />
    </div>
  )
}

