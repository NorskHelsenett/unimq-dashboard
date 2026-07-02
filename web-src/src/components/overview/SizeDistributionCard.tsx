// import { useState, useEffect } from 'react'
// import { Sparkline } from '@/components/charts/Sparkline'
// import { Skeleton } from '@/components/ui/skeleton'
// import { fmtBytes, fmtRate } from '@/lib/format'
// import { cn } from '@/lib/utils'
// import { Tooltip, TooltipContent, TooltipTrigger } from '../ui/tooltip'

// interface QueueDetail {
//     name: string
//     messages: number
//     message_bytes: number
//     history: number[]
//     consumers: number
//     publish_rate: number
//     deliver_rate: number
//     redeliver_rate: number
//     messages_unacknowledged: number
// }

// interface QueuesCardProps {
//     vhost: string
// }

// const cellPaddingStyling = "py-2.5 px-3"
// const subHeaderCellPaddingStyling = "px-3"
// const monoFontStyling = "text-center text-sm font-mono tabular-nums"



// function QueueRow({ queue, vhost }: { queue: QueueDetail; vhost: string }) {
//     const noConsumer = queue.messages > 0 && queue.consumers === 0
//     const href = `/queue?vhost=${encodeURIComponent(vhost)}&name=${encodeURIComponent(queue.name)}`

//     return (
//         <tr className="border-b last:border-0 hover:bg-gray-50">
//             <td className={cellPaddingStyling}>
//                 <a href={href} className="font-medium text-orange-500 hover:text-orange-600 hover:underline text-center block w-full">
//                     {queue.name}
//                 </a>
//             </td>
//             <td className={cn("border-l", monoFontStyling, cellPaddingStyling, queue.messages === 0 ? 'text-text-muted' : 'text-text-primary')}>
//                 {queue.messages}
//             </td>
//             <td className={cn("w-64", cellPaddingStyling)}>
//                 <Sparkline data={queue.history} />
//             </td>
//             <td className={cn("border-l", monoFontStyling, cellPaddingStyling, queue.message_bytes === 0 ? 'text-text-muted' : 'text-text-primary')}>
//                 {fmtBytes(queue.message_bytes)}
//             </td>
//             <td className={cn(monoFontStyling, cellPaddingStyling, queue.messages_unacknowledged === 0 ? 'text-text-muted' : 'text-text-primary')}>
//                 {queue.messages_unacknowledged}
//             </td>
//             <td className={cn("border-l", monoFontStyling, cellPaddingStyling)}>
//                 {noConsumer ? (
//                     <Tooltip>
//                         <TooltipTrigger asChild>
//                             <span className="inline-flex items-center gap-1 rounded bg-status-warning-bg-subtle border border-status-warning-border px-1.5 py-0.5 text-xs text-status-warning font-mono tabular-nums cursor-default">
//                                 0
//                                 <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="lucide lucide-triangle-alert-icon lucide-triangle-alert">
//                                     <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3"/>
//                                     <path d="M12 9v4"/>
//                                     <path d="M12 17h.01"/>
//                                 </svg>
//                             </span>
//                         </TooltipTrigger>
//                         <TooltipContent>
//                             <p className="max-w-xs text-xs">This queue has messages but no consumers. Messages are piling up with nothing reading them.</p>
//                         </TooltipContent>
//                     </Tooltip>
//                 ) : (
//                     <span className={cn(monoFontStyling, queue.consumers === 0 ? 'text-text-muted' : 'text-text-primary')}>{queue.consumers}</span>
//                 )}
//             </td>
//             <td className={cn('border-l', monoFontStyling, cellPaddingStyling, queue.publish_rate === 0 ? 'text-text-muted' : 'text-text-primary')}>
//                 {fmtRate(queue.publish_rate)}
//             </td>
//             <td className={cn(monoFontStyling, cellPaddingStyling, queue.deliver_rate === 0 ? 'text-text-muted' : 'text-text-primary')}>
//                 {fmtRate(queue.deliver_rate)}
//             </td>
//             <td className={cn(monoFontStyling, cellPaddingStyling, queue.redeliver_rate === 0 ? 'text-text-muted' : 'text-text-primary')}>
//                 {fmtRate(queue.redeliver_rate)}
//             </td>
//         </tr>
//     )
// }

// function HeaderCell({ children, className, colSpan }: { children?: React.ReactNode; className?: string; colSpan?: number }) {
//     return (
//         <th colSpan={colSpan} className={cn('py-2 text-xs font-bold uppercase tracking-wide text-text bg-brand/30', className)}>
//             {children}
//         </th>
//     )
// }

// function SubHeaderCell({ children, className }: { children?: React.ReactNode; className?: string }) {
//     return (
//         <th className={cn('py-1 text-[10px] font-semibold text-text bg-brand/15', className)}>
//             {children}
//         </th>
//     )
// }

// export function SizeDistributionCard({ vhost }: QueuesCardProps) {
//     const [data, setData] = useState<QueueDetail[] | null>(null)
//     const [error, setError] = useState(false)

//     useEffect(() => {
//         const load = () =>
//             fetch(`/api/queues?vhost=${encodeURIComponent(vhost)}`)
//                 .then((r) => {
//                     if (!r.ok) throw new Error()
//                     return r.json() as Promise<QueueDetail[]>
//                 })
//                 .then(setData)
//                 .catch(() => setError(true))

//         load()
//         const id = setInterval(load, 10_000)
//         return () => clearInterval(id)
//     }, [vhost])

//     function updateSizeDist(queues) {
//             const section = document.getElementById('size-dist-section');
//             if (!section) return;

//             const counts = [0, 0, 0, 0, 0, 0, 0, 0, 0];
//             let totalMessages = 0, totalBytes = 0;

//             for (const q of queues) {
//                 if (!q.messages || q.messages <= 0) continue;
//                 const bytes = q.message_bytes || 0;
//                 totalMessages += q.messages;
//                 totalBytes += bytes;
//                 const avgSize = bytes / q.messages;
//                 if (avgSize <= bucketLimits[4]) {
//                     counts[getBucket(avgSize)] += q.messages;
//                 }
//             }

//             if (totalMessages === 0) { section.style.display = 'none'; return; }
//             section.style.display = '';

//             const avgBytes = totalBytes / totalMessages;
//             document.getElementById('size-avg-value').innerHTML = fmtAvgSize(avgBytes);

//             const pcts = counts.map(c => totalMessages > 0 ? c / totalMessages * 100 : 0);
//             const maxPct = Math.max(...pcts, 0.1);

//             document.getElementById('size-dist-bars').innerHTML = pcts.map((pct, i) => {
//                 const barH = Math.round((pct / maxPct) * 60);
//                 const color = bucketColors[i];
//                 return `<div class="size-bucket">
//                     <div class="bucket-pct" style="color:${color}">${pct.toFixed(1)}%</div>
//                     <div class="bucket-bar-wrap">
//                         <div class="bucket-bar" style="height:${barH}px;background:${color}"></div>
//                     </div>
//                     <div class="bucket-label">${bucketLabels[i]}</div>
//                 </div>`;
//             }).join('');
//         }

//         function fetchQueues() {
//             if (!currentVhost) return;
//             fetch('/api/queues?vhost=' + encodeURIComponent(currentVhost))
//                 .then(r => r.json())
//                 .then(queues => {
//                     const tbody = document.getElementById('queues-body');
//                     if (!queues || queues.length === 0) {
//                         tbody.innerHTML = '<tr><td colspan="9" class="empty">Ingen køer på denne vhosten</td></tr>';
//                         updateSizeDist([]);
//                         return;
//                     }
//                     tbody.innerHTML = queues.map(q => `
//                         <tr>
//                             <td class="queue-name"><a href="/queue?vhost=${encodeURIComponent(currentVhost)}&name=${encodeURIComponent(q.name)}">${q.name}</a></td>
//                             <td class="msg-count">${q.messages}</td>
//                             <td class="sparkline-cell">${sparkline(q.history)}</td>
//                             <td class="queue-size">${fmtBytes(q.message_bytes)}</td>
//                             <td>${q.messages_unacknowledged}</td>
//                             <td>${q.messages > 0 && q.consumers === 0
//                             ? '<span class="no-consumer">0 &#9888;</span>'
//                             : q.consumers}</td>
//                             <td>${fmtRate(q.publish_rate)}</td>
//                             <td>${fmtRate(q.deliver_rate)}</td>
//                             <td>${fmtRate(q.redeliver_rate)}</td>
//                         </tr>`).join('');
//                     updateSizeDist(queues);
//                 })
//                 .catch(() => {
//                     document.getElementById('queues-body').innerHTML =
//                         '<tr><td colspan="9" class="error">Kunne ikke hente kødata</td></tr>';
//                 });
//         }
//         fetchQueues();



//     return (
//         <div className='w-full'>
//             <h2 className="text-lg font-semibold text-text-primary mb-3">Size distribution</h2>
//             <div className="rounded-lg border bg-surface-card overflow-hidden w-full">
                
//             </div>
//         </div>
//     )
// }

import { useState, useEffect, useMemo } from 'react'
import { fmtBytes } from '@/lib/format'
import { cn } from '@/lib/utils'
 
interface QueueDetail {
    name: string
    messages: number
    message_bytes: number
    history: number[]
    consumers: number
    publish_rate: number
    deliver_rate: number
    redeliver_rate: number
    messages_unacknowledged: number
}
 
interface SizeDistributionCardProps {
    vhost: string
}

const CHART_ITEMS = [
    {
        label: "0-100B",
        limit: 100,
        textColor: "text-emerald-600",
        barColor: "bg-emerald-500",
        borderColor: "border-emerald-500"
    },
    {
        label: "100-500B",
        limit: 500,
        textColor: "text-green-600",
        barColor: "bg-green-500",
        borderColor: "border-green-500"
    },
    {
        label: "500B-1KB",
        limit: 1024,
        textColor: "text-lime-600",
        barColor: "bg-lime-500",
        borderColor: "border-lime-500"
    },
    {
        label: "1-5KB",
        limit: 5120,
        textColor: "text-amber-600",
        barColor: "bg-amber-500",
        borderColor: "border-amber-500"
    },
    {
        label: "5-10KB",
        limit: 10240,
        textColor: "text-amber-600",
        barColor: "bg-amber-500",
        borderColor: "border-amber-500"
    },
    {
        label: "10-50KB",
        limit: 51200,
        textColor: "text-orange-600",
        barColor: "bg-orange-500",
        borderColor: "border-orange-500"
    },
    {
        label: "50-100KB",
        limit: 102400,
        textColor: "text-orange-600",
        barColor: "bg-orange-500",
        borderColor: "border-orange-500"
    },
    {
        label: "100-500KB",
        limit: 512000,
        textColor: "text-red-600",
        barColor: "bg-red-500",
        borderColor: "border-red-500"
    },
    {
        label: "500KB-1MB",
        limit: 1048576,
        textColor: "text-red-600",
        barColor: "bg-red-500",
        borderColor: "border-red-500"
    },
]
 
const BUCKET_LABELS = [
    '0-100B', '100-500B', '500B-1KB',
    '1-5KB', '5-10KB', '10-50KB',
    '50-100KB', '100-500KB', '500KB-1MB',
]
 
const BUCKET_LIMITS = [100, 500, 1024, 5120, 10240, 51200, 102400, 512000, 1048576]
 
const BUCKET_TEXT_COLORS = [
    'text-emerald-600', 'text-green-600', 'text-lime-600',
    'text-amber-600', 'text-amber-600',
    'text-orange-600', 'text-orange-600',
    'text-red-600', 'text-red-600',
]
 
const BUCKET_BAR_COLORS = [
    'bg-green-500', 'bg-green-500', 'bg-lime-500',
    'bg-amber-400', 'bg-amber-400',
    'bg-orange-500', 'bg-orange-500',
    'bg-red-500', 'bg-red-500',
]
 
const BUCKET_BORDER_COLORS = [
    'border-green-300', 'border-green-300', 'border-lime-300',
    'border-amber-300', 'border-amber-300',
    'border-orange-300', 'border-orange-300',
    'border-red-300', 'border-red-300',
]
 
const BAR_MAX_HEIGHT = 60
 
function getBucketIndex(avgSize: number): number {
    for (let i = 0; i < BUCKET_LIMITS.length - 1; i++) {
        if (avgSize < BUCKET_LIMITS[i]) return i
    }
    return BUCKET_LIMITS.length - 1
}
 
function computeDist(queues: QueueDetail[]) {
    const counts = new Array<number>(BUCKET_LABELS.length).fill(0)
    let totalMessages = 0
    let totalBytes = 0
 
    for (const q of queues) {
        if (!q.messages || q.messages <= 0) continue
        const bytes = q.message_bytes ?? 0
        totalMessages += q.messages
        totalBytes += bytes
        const avgSize = bytes / q.messages
        if (avgSize <= BUCKET_LIMITS[4]) {
            counts[getBucketIndex(avgSize)] += q.messages
        }
    }
 
    const percentages = counts.map(c =>
        totalMessages > 0 ? (c / totalMessages) * 100 : 0
    )
 
    return {
        percentages,
        totalMessages,
        avgBytes: totalMessages > 0 ? totalBytes / totalMessages : 0,
    }
}
 
export function SizeDistributionCard({ vhost }: SizeDistributionCardProps) {
    const [data, setData] = useState<QueueDetail[] | null>(null)
 
    useEffect(() => {
        const load = () =>
            fetch(`/v1/vhosts/vhost=${encodeURIComponent(vhost)}/queues`)
                .then((r) => {
                    if (!r.ok) throw new Error()
                    return r.json() as Promise<QueueDetail[]>
                })
                .then(setData)
                .catch(() => {})
 
        load()
        const id = setInterval(load, 10_000)
        return () => clearInterval(id)
    }, [vhost])
 
    const dist = useMemo(() => (data ? computeDist(data) : null), [data])
 
    if (!dist || dist.totalMessages === 0) return (
        <div className='w-full'>
                <h2 className="text-lg font-semibold text-text-primary mb-3">Message size distribution</h2>
                <p>No messages present</p>
        </div>
    )
 
    const maxPct = Math.max(...dist.percentages, 0.1)
 
    return (
        <div className='w-full'>
            <h2 className="text-lg font-semibold text-text-primary mb-3">Message size distribution</h2>
            <div className="rounded-lg border bg-surface-card p-4 flex gap-8 items-center w-full">
                <div className='self-stretch border-r pr-8 flex flex-col justify-center'>
                    <p className="text-xs text-gray-500 mb-1">AVG. SIZE</p>
                    <p className="font-mono text-2xl font-semibold text-gray-900">{fmtBytes(dist.avgBytes)}</p>
                </div>
    
                <div className="flex-1">
                    <div className="flex gap-1">
                        {dist.percentages.map((pct, i) => {
                            const barHeight = Math.max(Math.round((pct / maxPct) * BAR_MAX_HEIGHT), 0)
                            const isEmpty = pct < 0.05
    
                            return (
                                <div key={CHART_ITEMS[i].label} className="flex-1 flex flex-col items-center gap-1">
                                    <span className={cn('text-xs font-mono', isEmpty ? 'text-gray-400' : CHART_ITEMS[i].textColor)}>
                                        {pct.toFixed(1)}%
                                    </span>
    
                                    <div className="w-full flex items-end" style={{ height: `${BAR_MAX_HEIGHT}px` }}>
                                        {isEmpty ? (
                                            <div className={cn('w-full border-b-2', CHART_ITEMS[i].borderColor)} />
                                        ) : (
                                            <div
                                                className={cn('w-full rounded-t-sm', CHART_ITEMS[i].barColor)}
                                                style={{ height: `${barHeight}px` }}
                                            />
                                        )}
                                    </div>
    
                                    <span className="text-gray-400 text-center leading-tight text-[10px]">
                                        {CHART_ITEMS[i].label}
                                    </span>
                                </div>
                            )
                        })}
                    </div>
                </div>
            </div>
        </div>
    )
}
