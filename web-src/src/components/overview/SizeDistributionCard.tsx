import { useMemo } from 'react'
import { fmtBytes } from '@/lib/format'
import { cn } from '@/lib/utils'
import type { SizeDistributionCardProps, QueueDetail } from '@/types/queues'
import { SectionCard, SectionCardHeader } from '@/components/ui/section-card'
import { BarChart3 } from 'lucide-react'

const BUCKETS = [
    { label: '0-100B',    limit: 100,     barColor: 'bg-emerald-500', dotColor: 'bg-emerald-500', textColor: 'text-emerald-600' },
    { label: '100-500B',  limit: 500,     barColor: 'bg-green-500',   dotColor: 'bg-green-500',   textColor: 'text-green-600'   },
    { label: '500B-1KB',  limit: 1024,    barColor: 'bg-lime-500',    dotColor: 'bg-lime-500',    textColor: 'text-lime-600'    },
    { label: '1-5KB',     limit: 5120,    barColor: 'bg-yellow-500',  dotColor: 'bg-yellow-500',  textColor: 'text-yellow-600'  },
    { label: '5-10KB',    limit: 10240,   barColor: 'bg-amber-500',   dotColor: 'bg-amber-500',   textColor: 'text-amber-600'   },
    { label: '10-50KB',   limit: 51200,   barColor: 'bg-orange-500',  dotColor: 'bg-orange-500',  textColor: 'text-orange-600'  },
    { label: '50-100KB',  limit: 102400,  barColor: 'bg-red-400',     dotColor: 'bg-red-400',     textColor: 'text-red-500'     },
    { label: '100-500KB', limit: 512000,  barColor: 'bg-red-600',     dotColor: 'bg-red-600',     textColor: 'text-red-700'     },
    { label: '500KB+',    limit: Infinity, barColor: 'bg-rose-700',   dotColor: 'bg-rose-700',    textColor: 'text-rose-800'    },
]

function getBucketIndex(avgSize: number): number {
    for (let i = 0; i < BUCKETS.length - 1; i++) {
        if (avgSize < BUCKETS[i].limit) return i
    }
    return BUCKETS.length - 1
}

function computeDist(queues: QueueDetail[]) {
    const counts = new Array<number>(BUCKETS.length).fill(0)
    let totalMessages = 0
    let totalBytes = 0

    for (const q of queues) {
        if (!q.messages || q.messages <= 0) continue
        const bytes = q.message_bytes ?? 0
        totalMessages += q.messages
        totalBytes += bytes
        counts[getBucketIndex(bytes / q.messages)] += q.messages
    }

    return {
        percentages: counts.map(c => (totalMessages > 0 ? (c / totalMessages) * 100 : 0)),
        totalMessages,
        avgBytes: totalMessages > 0 ? totalBytes / totalMessages : 0,
    }
}

export function SizeDistributionCard({ queues }: SizeDistributionCardProps) {
    const dist = useMemo(() => (queues ? computeDist(queues) : null), [queues])

    const nonZero = dist
        ? BUCKETS.map((b, i) => ({ ...b, pct: dist.percentages[i] })).filter(b => b.pct > 0)
        : []

    return (
        <SectionCard accent="none" className="min-w-0 h-full">
            <SectionCardHeader
                title="Message Size Distribution"
                icon={<BarChart3 className="w-4 h-4 text-gray-400" />}
                action={
                    dist && dist.totalMessages > 0 ? (
                        <span className="text-sm font-bold tabular-nums text-text-primary">
                            avg. {fmtBytes(dist.avgBytes)}
                        </span>
                    ) : undefined
                }
            />

            {!dist || dist.totalMessages === 0 ? (
                <p className="text-sm text-text-muted">No messages present.</p>
            ) : (
                <div className="space-y-4">
                    {/* Stacked horizontal bar */}
                    <div className="h-3 rounded-full overflow-hidden flex bg-gray-100 gap-px">
                        {BUCKETS.map((b, i) => {
                            const pct = dist.percentages[i]
                            if (pct < 0.05) return null
                            return (
                                <div
                                    key={b.label}
                                    className={cn('h-full transition-all', b.barColor)}
                                    style={{ width: `${pct}%` }}
                                    title={`${b.label}: ${pct.toFixed(1)}%`}
                                />
                            )
                        })}
                    </div>

                    {/* Legend — only non-zero buckets */}
                    <div className="flex flex-wrap gap-x-4 gap-y-1.5">
                        {nonZero.map(b => (
                            <div key={b.label} className="flex items-center gap-1.5">
                                <div className={cn('w-2 h-2 rounded-sm shrink-0', b.dotColor)} />
                                <span className="text-xs text-text-muted">{b.label}</span>
                                <span className={cn('text-xs font-semibold tabular-nums', b.textColor)}>
                                    {b.pct.toFixed(1)}%
                                </span>
                            </div>
                        ))}
                    </div>
                </div>
            )}
        </SectionCard>
    )
}
