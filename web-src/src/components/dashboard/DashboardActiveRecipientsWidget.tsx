import { VhostNotification } from '@/types/notifications'
import { SectionCard, SectionCardHeader } from '../ui/section-card'
import { StatusDot } from '../ui/status-dot'
import { ArrowRight, Users } from 'lucide-react'
import { cn } from '@/lib/utils'

const TYPE_LABELS: Record<string, string> = {
  teams: 'Teams',
  slack: 'Slack',
  webhook: 'Webhook',
}

export function DashboardActiveRecipientsWidget({
  notification,
}: {
  notification: VhostNotification | null
}) {
  const recipients = notification?.Recipients ?? []

  return (
    <SectionCard accent="blue" className="min-w-0 h-full flex flex-col">
      <SectionCardHeader
        title="Notification recipients"
        icon={<Users className="w-4 h-4 text-blue-400" />}
      />
      {recipients.length === 0 ? (
        <p className="text-sm text-text-muted">No recipients configured.</p>
      ) : (
        <div>
          {recipients.map(r => (
            <div
              key={r.id}
              className="flex items-center justify-between py-1.5 border-b last:border-0 border-border-card"
            >
              <span className="flex items-center gap-2 text-sm">
                <StatusDot color="blue" pulse={false} />
                <span className="font-medium">{r.name}</span>
              </span>
              <span className="text-xs text-text-muted">
                {TYPE_LABELS[r.type] ?? r.type}
              </span>
            </div>
          ))}
        </div>
      )}
      <p className="mt-auto pt-3">
        <a href={`/notifications?vhost=${encodeURIComponent(notification?.Name ?? '')}`} className={cn("text-submit-button text-xs hover:font-semibold transition-colors inline-flex items-center gap-1 [text-decoration:none] hover:[text-decoration:none]")}>
          View all recipients <ArrowRight className="w-3 h-3 inline ml-1" />
        </a>
      </p>
    </SectionCard>
  )
}
