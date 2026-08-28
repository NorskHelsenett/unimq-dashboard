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

const progressBarColor: Record<StatusKey, string> = {
  danger: 'bg-red-500',
  warning: 'bg-amber-400',
  neutral: 'bg-gray-300',
}

function ProgressBar({ value, max, colorKey }: { value: number; max: number; colorKey: StatusKey }) {
  const pct = Math.min(100, Math.max(0, (value / max) * 100))
  return (
    <div className="flex items-center gap-2 mt-3">
      <div className="h-1 flex-1 rounded-full bg-black/10 dark:bg-white/10 overflow-hidden">
        <div
          className={cn('h-full rounded-full transition-[width] duration-300', progressBarColor[colorKey])}
          style={{ width: `${pct}%` }}
        />
      </div>
      <span className="text-xs text-text-muted tabular-nums w-9 text-right">{Math.round(pct)}%</span>
    </div>
  )
}

// Visualises unacked messages as a queue of blocks (soft cap: 100)
function QueueBar({ value }: { value: number }) {
  const BLOCKS = 10
  const SOFT_MAX = 100
  const filled = value > 0 ? Math.min(BLOCKS, Math.max(1, Math.ceil((value / SOFT_MAX) * BLOCKS))) : 0
  const blockColor =
    filled >= BLOCKS * 0.8 ? 'bg-red-400' :
    filled >= BLOCKS * 0.5 ? 'bg-amber-400' :
    'bg-teal-400'
  return (
    <div className="flex gap-0.5 mt-3">
      {Array.from({ length: BLOCKS }).map((_, i) => (
        <div
          key={i}
          className={cn('h-2 flex-1 rounded-[2px]', i < filled ? blockColor : 'bg-black/10 dark:bg-white/10')}
        />
      ))}
    </div>
  )
}

interface StatCardProps {
  label: string
  tooltip: string
  value: number
  sub: ReactNode
  max?: number
  icon: ReactNode
  cardBg: string
  cardBorder: string
  iconBg: string
  iconColor: string
  bar?: ReactNode
}

function StatCard({ label, tooltip, value, sub, max, icon, cardBg, cardBorder, iconBg, iconColor, bar }: StatCardProps) {
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
          <div className="text-sm mt-0.5 font-medium text-text-muted">{sub}</div>
        </div>
      </div>
      {bar}
    </div>
  )
}

export function LimitsCard({ connections, channels, queues, unacked, maxConnections, maxQueues }: LimitsCardProps) {
  const connColorKey = limitColor(connections, maxConnections)
  const queueColorKey = limitColor(queues, maxQueues)
  const channelColorKey = limitColor(channels, 1000)
  return (
    <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
      <StatCard
        label="Connections"
        tooltip={`Maks antall connections til en vhost er ${maxConnections}. Etter at dette antallet er nådd, vil det ikke lenger være mulig å opprette nye connections før du er under grensen igjen.`}
        value={connections}
        sub={<>of <span className="text-blue-400">{maxConnections}</span> allowed</>}
        max={maxConnections > 0 ? maxConnections : undefined}
        icon={<Network className="w-4 h-4" />}
        cardBg="bg-surface-card" cardBorder="border-border-card"
        iconBg="bg-blue-100" iconColor="text-blue-500"
        bar={maxConnections > 0 ? <ProgressBar value={connections} max={maxConnections} colorKey={connColorKey} /> : undefined}
      />
      <StatCard
        label="Channels"
        tooltip="Vi anbefaler å holde antallet channels per vhost under 1000."
        value={channels}
        sub={<>of <span className="text-violet-400">1000</span> recommended</>}
        icon={<GitBranch className="w-4 h-4" />}
        cardBg="bg-surface-card" cardBorder="border-border-card"
        iconBg="bg-violet-100" iconColor="text-violet-500"
        bar={<ProgressBar value={channels} max={1000} colorKey={channelColorKey} />}
      />
      <StatCard
        label="Queues"
        tooltip={`Maks antall queues på en vhost er ${maxQueues}. Etter at dette antallet er nådd, vil det ikke lenger være mulig å opprette nye queues før du er under grensen igjen.`}
        value={queues}
        sub={<>of <span className="text-orange-400">{maxQueues}</span> allowed</>}
        max={maxQueues > 0 ? maxQueues : undefined}
        icon={<LayoutList className="w-4 h-4" />}
        cardBg="bg-surface-card" cardBorder="border-border-card"
        iconBg="bg-orange-100" iconColor="text-orange-500"
        bar={maxConnections > 0 ? <ProgressBar value={queues} max={maxQueues} colorKey={queueColorKey} /> : undefined}
      />
      <StatCard
        label="Unacked messages"
        tooltip="Totalt antall unacked messages på vhosten. Meldinger som hentes, men ikke enda er konsumert, havner i «unacked state». Disse meldingene lagres i minne, det er derfor ikke ønskelig å ha for mange."
        value={unacked}
        sub="keep low"
        icon={<Hourglass className="w-4 h-4" />}
        cardBg="bg-surface-card" cardBorder="border-border-card"
        iconBg="bg-teal-100" iconColor="text-teal-500"
        bar={<QueueBar value={unacked} />}
      />
    </div>
  )
}

