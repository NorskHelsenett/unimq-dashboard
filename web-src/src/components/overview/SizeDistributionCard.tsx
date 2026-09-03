import { useMemo } from 'react'
import { fmtBytes } from '@/lib/format'
import { cn } from '@/lib/utils'
import type { SizeDistributionCardProps, QueueDetail } from '@/types/queues'
 
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
 
export function SizeDistributionCard({ queues}: SizeDistributionCardProps) {
    const dist = useMemo(() => (queues ? computeDist(queues) : null), [queues])
 
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
