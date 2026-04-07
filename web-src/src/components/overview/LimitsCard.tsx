// import { cn } from "@/lib/utils"
// import { Tooltip, TooltipContent, TooltipTrigger } from "../ui/tooltip"

// interface LimitsCardProps {
//     selected: string
//     connections: number
//     channels: number
//     queues: number
//     unacked: number
//     maxConnections: number
//     maxQueues: number
// }

// function limitColor(value: number, max: number): string {
//     if (value >= max) {
//         return "text-status-danger"
//     }
//     if (value >= max/2) {
//         return "text-status-warning"
//     }
//     return "text-status-ok"
// }

// interface RowProps {
//     label: string
//     tooltip: string
//     value: number
//     max?: number
// }

// const LimitRow = ({ label, tooltip, value, max}: RowProps) => {
//     const colorClass = max ? limitColor(value, max) : "text-text-primary"
//     const display = max ? `${value} of ${max}` : `${value}`

//     return (
//         <tr className="border-t border-border-card">
//             <td className="py-3 pr-4 text-sm text-text-secondary">
//                 <span className="flex items-center gap-1.5">
//                     {label}
//                     <Tooltip>
//                         <TooltipTrigger asChild>
//                             <span className="inline-flex items-center justify-center w-4 h-4 rounded-full bg-gray-100 text-gray-400 text-xs cursor-default select-none">
//                                 ?
//                             </span>
//                         </TooltipTrigger>
//                         <TooltipContent>
//                             <p className="max-w-xs text-xs">{tooltip}</p>
//                         </TooltipContent>
//                     </Tooltip>
//                 </span>
//             </td>
//             <td className={cn('py-3 text-right text-sm font-mono font-medium', colorClass)}>
//                 {display}
//             </td>
//         </tr>
//     )
// }

// export const LimitsCard = ({
//   selected,
//   connections,
//   channels,
//   queues,
//   unacked,
//   maxConnections,
//   maxQueues,
// }: LimitsCardProps) => {
//   return (
//     <div className="bg-surface-card border border-border-card rounded-lg p-6">
//       <h2 className="text-base font-semibold text-text-primary mb-4">
//         Limits — {selected}
//       </h2>
//       <table className="w-full">
//         <tbody>
//           <LimitRow
//             label="Connections"
//             tooltip="Max connections per vhost is 300. Once reached, new connections are rejected until the count drops below the limit."
//             value={connections}
//             max={maxConnections}
//           />
//           <LimitRow
//             label="Channels"
//             tooltip="We recommend keeping channels per vhost below 1000."
//             value={channels}
//           />
//           <LimitRow
//             label="Queues"
//             tooltip="Max queues per vhost is 150. Once reached, new queues cannot be created until the count drops below the limit."
//             value={queues}
//             max={maxQueues}
//           />
//           <LimitRow
//             label="Unacked messages"
//             tooltip="Messages delivered but not yet acknowledged. These are held in memory — a high count is undesirable."
//             value={unacked}
//           />
//         </tbody>
//       </table>
//     </div>
//   )
// }

import { cn } from '@/lib/utils'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'

interface LimitsCardProps {
  selected: string
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
              <span className="inline-flex items-center justify-center w-4 h-4 rounded-full
                               bg-gray-100 text-gray-400 text-xs cursor-default select-none normal-case tracking-normal">
                ?
              </span>
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

export function LimitsCard({
  selected,
  connections,
  channels,
  queues,
  unacked,
  maxConnections,
  maxQueues,
}: LimitsCardProps) {
  return (
    <div>
      <h2 className="text-base font-semibold text-text-primary mb-3">
        Limits
      </h2>
      <div className="grid grid-cols-2 gap-3">
        <MetricTile
          label="Connections"
          tooltip="Max connections per vhost is 300. Once reached, new connections are rejected until the count drops below the limit."
          value={connections}
          max={maxConnections}
          sub={`limit ${maxConnections}`}
        />
        <MetricTile
          label="Channels"
          tooltip="We recommend keeping channels per vhost below 1000."
          value={channels}
          sub="rec. <1000"
        />
        <MetricTile
          label="Queues"
          tooltip="Max queues per vhost is 150. Once reached, new queues cannot be created until the count drops below the limit."
          value={queues}
          max={maxQueues}
          sub={`limit ${maxQueues}`}
        />
        <MetricTile
          label="Unacked messages"
          tooltip="Messages delivered but not yet acknowledged. These are held in memory — a high count is undesirable."
          value={unacked}
          sub="keep low"
        />
      </div>
    </div>
  )
}