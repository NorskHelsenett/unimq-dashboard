import { Clock, Calendar } from 'lucide-react'
import { getInitials, formatAgo, formatIn } from './profileUtils'
import { ProfileHeroCardProps } from '@/types/profile'

export function ProfileHeroCard({ name, email, verified, primaryRole, iat, exp }: ProfileHeroCardProps) {
  return (
    <div className="border border-border-card rounded-xl bg-surface-card p-6 flex flex-wrap items-center gap-6">
      <div className="size-20 rounded-full bg-brand/10 border-2 border-brand/20 flex items-center justify-center shrink-0">
        <span className="text-2xl font-bold text-brand">{getInitials(name)}</span>
      </div>
      <div className="min-w-0 flex-1">
        <h2 className="text-xl font-semibold text-text-primary">{name ?? '—'}</h2>
        <div className="flex items-center gap-2 mt-0.5 flex-wrap">
          {email && <span className="text-sm text-text-muted">{email}</span>}
          {verified && (
            <span className="px-1.5 py-0.5 rounded-full bg-status-ok/10 text-status-ok text-xs font-medium">✓ Verified</span>
          )}
        </div>
        {primaryRole && (
          <span className="inline-flex items-center gap-1 mt-1.5 px-2 py-0.5 rounded-full bg-amber-500/10 text-amber-500 text-xs font-medium capitalize">
            ★ {primaryRole}
          </span>
        )}
      </div>
      <div className="flex flex-wrap gap-10 shrink-0">
        {iat && (
          <div className="flex items-center gap-2">
            <Clock size={20} className="text-text-muted shrink-0" />
            <div>
              <p className="text-xs text-text-muted">Signed in</p>
              <p className="text-sm font-medium text-text-primary">{formatAgo(iat)}</p>
            </div>
          </div>
        )}
        {exp && (
          <div className="flex items-center gap-2">
            <Calendar size={20} className="text-text-muted shrink-0" />
            <div>
              <p className="text-xs text-text-muted">Session expires</p>
              <p className="text-sm font-medium text-text-primary">{formatIn(exp)}</p>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
