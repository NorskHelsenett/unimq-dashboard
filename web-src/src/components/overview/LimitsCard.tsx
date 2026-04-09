import { cn } from '@/lib/utils'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'

interface LimitsCardProps {
  connections: number
  channels: number
  queues: number
  unacked: number
  maxConnections: number
  maxQueues: number
}

const bgMap: Record<string, string> = {
  'status-danger':  'bg-status-danger-bg border-status-danger-border',
  'status-warning': 'bg-status-warning-bg-subtle border-status-warning-border',
  'status-ok':      'bg-status-ok-bg-subtle border-status-ok-border',
  'text-text-primary': 'bg-surface-card border-border-card',
}

const textMap: Record<string, string> = {
  'status-danger':  'text-status-danger',
  'status-warning': 'text-status-warning',
  'status-ok':      'text-status-ok',
  'text-text-primary': 'text-text-primary',
}

function limitColor(value: number, max: number): string {
  if (value >= max / 5 * 4)       return 'status-danger'
  if (value >= max / 2)   return 'status-warning'
  return 'status-ok'
}

interface MetricTileProps {
  label: string
  tooltip: string
  value: number
  max?: number
  sub?: string
}

function MetricTile({ label, tooltip, value, max, sub }: MetricTileProps) {
  const colorKey = max ? limitColor(value, max) : 'text-text-primary'
  const bgClass   = bgMap[colorKey]
  const textClass = textMap[colorKey]

  return (
    <div className={cn('rounded-lg p-4 flex justify-between items-center border', bgClass)}>
      <div className="flex flex-col gap-0.5">
        <span className="flex items-center gap-1.5 text-xs text-text-muted uppercase tracking-wide">
          {label}
          <Tooltip>
            <TooltipTrigger asChild>
              <button
                type="button"
                aria-label={`Show information about ${label}`}
                className="inline-flex items-center justify-center w-4 h-4 rounded-full
                               bg-gray-100 text-gray-400 text-xs cursor-default select-none normal-case tracking-normal"
              >
                ?
              </button>
            </TooltipTrigger>
            <TooltipContent>
              <p className="max-w-xs text-xs">{tooltip}</p>
            </TooltipContent>
          </Tooltip>
        </span>
        {sub && <span className="text-xs text-text-muted">{sub}</span>}
      </div>
      <span className={cn('text-2xl font-mono font-semibold tabular-nums', textClass)}>
        {value}
      </span>
    </div>
  )
}

export function LimitsCard({ connections, channels, queues, unacked, maxConnections, maxQueues }: LimitsCardProps) {
  return (
    <div className='min-w-sm w-xl'>
      <h2 className="text-base font-semibold text-text-primary mb-3">
        Limits
      </h2>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <MetricTile
          label="Connections"
          tooltip={`Maks antall connections til en vhost er ${maxConnections}. Etter at dette antallet er nådd, vil det ikke lenger være mulig å opprette nye connections før du er under grensen igjen.`}
          value={connections}
          max={maxConnections}
          sub={`limit ${maxConnections}`}
        />
        <MetricTile
          label="Channels"
          tooltip="Vi anbefaler å holde antallet channels per vhost under 1000."
          value={channels}
          sub="rec. <1000"
        />
        <MetricTile
          label="Queues"
          tooltip={`Maks antall queues på en vhost er ${maxQueues}. Etter at dette antallet er nådd, vil det ikke lenger være mulig å opprette nye queues før du er under grensen igjen.`}
          value={queues}
          max={maxQueues}
          sub={`limit ${maxQueues}`}
        />
        <MetricTile
          label="Unacked messages"
          tooltip="Totalt antall unacked messages på vhosten. Meldinger som hentes, men ikke enda er konsumert, havner i «unacked state». Disse meldingene lagres i minne, det er derfor ikke ønskelig å ha for mange."
          value={unacked}
          sub="keep low"
        />
      </div>
    </div>
  )
}