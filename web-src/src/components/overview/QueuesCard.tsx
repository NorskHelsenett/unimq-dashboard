import { useState, useEffect } from 'react'
import { Sparkline } from '@/components/charts/Sparkline'
import { Skeleton } from '@/components/ui/skeleton'
import { fmtBytes, fmtRate } from '@/lib/format'
import { cn } from '@/lib/utils'
import { Tooltip, TooltipContent, TooltipTrigger } from '../ui/tooltip'

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

interface QueuesCardProps {
    vhost: string
}

const cellPaddingStyling = "py-2.5 px-3"
const subHeaderCellPaddingStyling = "px-3"
const monoFontStyling = "text-center text-sm font-mono tabular-nums"



function QueueRow({ queue, vhost }: { queue: QueueDetail; vhost: string }) {
    const noConsumer = queue.messages > 0 && queue.consumers === 0
    const href = `/queue?vhost=${encodeURIComponent(vhost)}&name=${encodeURIComponent(queue.name)}`

    return (
        <tr className="border-b last:border-0 hover:bg-gray-50">
            <td className={cellPaddingStyling}>
                <a href={href} className="font-medium text-orange-500 hover:text-orange-600 hover:underline text-center block w-full">
                    {queue.name}
                </a>
            </td>
            <td className={cn("border-l", monoFontStyling, cellPaddingStyling, queue.messages === 0 ? 'text-text-muted' : 'text-text-primary')}>
                {queue.messages}
            </td>
            <td className={cn("w-64", cellPaddingStyling)}>
                <Sparkline data={queue.history} />
            </td>
            <td className={cn("border-l", monoFontStyling, cellPaddingStyling, queue.message_bytes === 0 ? 'text-text-muted' : 'text-text-primary')}>
                {fmtBytes(queue.message_bytes)}
            </td>
            <td className={cn(monoFontStyling, cellPaddingStyling, queue.messages_unacknowledged === 0 ? 'text-text-muted' : 'text-text-primary')}>
                {queue.messages_unacknowledged}
            </td>
            <td className={cn("border-l", monoFontStyling, cellPaddingStyling)}>
                {noConsumer ? (
                    <Tooltip>
                        <TooltipTrigger asChild>
                            <span className="inline-flex items-center gap-1 rounded bg-status-warning-bg-subtle border border-status-warning-border px-1.5 py-0.5 text-xs text-status-warning font-mono tabular-nums cursor-default">
                                0
                                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="lucide lucide-triangle-alert-icon lucide-triangle-alert">
                                    <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3"/>
                                    <path d="M12 9v4"/>
                                    <path d="M12 17h.01"/>
                                </svg>
                            </span>
                        </TooltipTrigger>
                        <TooltipContent>
                            <p className="max-w-xs text-xs">This queue has messages but no consumers. Messages are piling up with nothing reading them.</p>
                        </TooltipContent>
                    </Tooltip>
                ) : (
                    <span className={cn(monoFontStyling, queue.consumers === 0 ? 'text-text-muted' : 'text-text-primary')}>{queue.consumers}</span>
                )}
            </td>
            <td className={cn('border-l', monoFontStyling, cellPaddingStyling, queue.publish_rate === 0 ? 'text-text-muted' : 'text-text-primary')}>
                {fmtRate(queue.publish_rate)}
            </td>
            <td className={cn(monoFontStyling, cellPaddingStyling, queue.deliver_rate === 0 ? 'text-text-muted' : 'text-text-primary')}>
                {fmtRate(queue.deliver_rate)}
            </td>
            <td className={cn(monoFontStyling, cellPaddingStyling, queue.redeliver_rate === 0 ? 'text-text-muted' : 'text-text-primary')}>
                {fmtRate(queue.redeliver_rate)}
            </td>
        </tr>
    )
}

function HeaderCell({ children, className, colSpan }: { children?: React.ReactNode; className?: string; colSpan?: number }) {
    return (
        <th colSpan={colSpan} className={cn('py-2 text-xs font-bold uppercase tracking-wide text-text bg-brand/30', className)}>
            {children}
        </th>
    )
}

function SubHeaderCell({ children, className }: { children?: React.ReactNode; className?: string }) {
    return (
        <th className={cn('py-1 text-[10px] font-semibold text-text bg-brand/15', className)}>
            {children}
        </th>
    )
}

export function QueuesCard({ vhost }: QueuesCardProps) {
    const [data, setData] = useState<QueueDetail[] | null>(null)
    const [error, setError] = useState(false)

    useEffect(() => {
        const load = () =>
            //fetch(`/api/queues?vhost=${encodeURIComponent(vhost)}`)
            fetch(`/v1/vhosts/vhost=${encodeURIComponent(vhost)}/queues`)
                .then((r) => {
                    if (!r.ok) throw new Error()
                    return r.json() as Promise<QueueDetail[]>
                })
                .then(setData)
                .catch(() => setError(true))

        load()
        const id = setInterval(load, 10_000)
        return () => clearInterval(id)
    }, [vhost])

    return (
        <div className='w-full'>
            <h2 className="text-lg font-semibold text-text-primary mb-3">Queues</h2>
            <div className="rounded-lg border bg-surface-card overflow-hidden w-full">
                <div className="overflow-x-auto">
                    <table className="w-full border-collapse text-sm">
                        <thead>
                            <tr>
                                <HeaderCell>Queue</HeaderCell>
                                <HeaderCell className="border-l" colSpan={2}>Messages</HeaderCell>
                                <HeaderCell className="border-l" colSpan={2}>Storage</HeaderCell>
                                <HeaderCell className="border-l px-3">Consumers</HeaderCell>
                                <HeaderCell className="border-l" colSpan={3}>Rates</HeaderCell>
                            </tr>
                            <tr className="border-b">
                                <SubHeaderCell className={subHeaderCellPaddingStyling}>Name</SubHeaderCell>
                                <SubHeaderCell className={cn("border-l", subHeaderCellPaddingStyling)}>Count</SubHeaderCell>
                                <SubHeaderCell className={subHeaderCellPaddingStyling}>History</SubHeaderCell>
                                <SubHeaderCell className={cn("border-l", subHeaderCellPaddingStyling)}>Size</SubHeaderCell>
                                <SubHeaderCell className={subHeaderCellPaddingStyling}>Unacked</SubHeaderCell>
                                <SubHeaderCell className={cn("border-l", subHeaderCellPaddingStyling)}>Number</SubHeaderCell>
                                <SubHeaderCell className={cn("border-l", subHeaderCellPaddingStyling)}>Publish</SubHeaderCell>
                                <SubHeaderCell className={subHeaderCellPaddingStyling}>Deliver</SubHeaderCell>
                                <SubHeaderCell className={subHeaderCellPaddingStyling}>Redelivered</SubHeaderCell>
                            </tr>
                        </thead>
                        <tbody>
                            {error && (
                                <tr>
                                    <td colSpan={9} className="py-6 text-center text-sm text-status-danger">
                                        Failed to load queue data
                                    </td>
                                </tr>
                            )}
                            {!error && data === null && (
                                Array.from({ length: 3 }).map((_, i) => (
                                    <tr key={i} className="border-b last:border-0">
                                        <td className={cellPaddingStyling}><Skeleton className="h-4 w-28" /></td>
                                        <td className={cellPaddingStyling}><Skeleton className="h-4 w-8 ml-auto" /></td>
                                        <td className={cellPaddingStyling}><Skeleton className="h-4 w-16" /></td>
                                        <td className={cellPaddingStyling}><Skeleton className="h-4 w-12 ml-auto" /></td>
                                        <td className={cellPaddingStyling}><Skeleton className="h-4 w-8 ml-auto" /></td>
                                        <td className={cellPaddingStyling}><Skeleton className="h-4 w-6 ml-auto" /></td>
                                        <td className={cellPaddingStyling}><Skeleton className="h-4 w-12 ml-auto" /></td>
                                        <td className={cellPaddingStyling}><Skeleton className="h-4 w-12 ml-auto" /></td>
                                        <td className={cellPaddingStyling}><Skeleton className="h-4 w-12 ml-auto" /></td>
                                    </tr>
                                ))
                            )}
                            {!error && data !== null && data.length === 0 && (
                                <tr>
                                    <td colSpan={9} className="py-6 text-center text-sm text-text-muted">
                                        No queues on this vhost
                                    </td>
                                </tr>
                            )}
                            {!error && data !== null && data.map((queue) => (
                                <QueueRow key={queue.name} queue={queue} vhost={vhost} />
                            ))}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    )
}