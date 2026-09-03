import { Shield } from 'lucide-react'
import { AccessPermissionsCardProps } from '@/types/profile'

export function AccessPermissionsCard({ displayRoles, resourceAccess, groups }: AccessPermissionsCardProps) {
  const hasContent = displayRoles.length > 0 || Object.keys(resourceAccess).length > 0 || groups.length > 0

  return (
    <div className="border border-border-card rounded-xl bg-surface-card">
      <div className="flex items-center gap-2 px-5 py-4 border-b border-border-card">
        <Shield size={16} className="text-text-muted" />
        <div>
          <h3 className="font-semibold text-text-primary text-sm">Access & permissions</h3>
          <p className="text-xs text-text-muted">Your roles and resource access.</p>
        </div>
      </div>
      <div className="p-5 space-y-5">
        {!hasContent && (
          <p className="text-sm text-text-muted">No permission data available.</p>
        )}
        {displayRoles.length > 0 && (
          <div>
            <p className="text-xs text-text-muted uppercase tracking-wide mb-2">Role</p>
            <div className="flex flex-wrap gap-2">
              {displayRoles.map(role => (
                <span key={role} className="px-2.5 py-1 rounded-full bg-amber-500/10 text-amber-500 text-xs font-medium capitalize">
                  ★ {role}
                </span>
              ))}
            </div>
          </div>
        )}
        {Object.keys(resourceAccess).length > 0 && (
          <div>
            <p className="text-xs text-text-muted uppercase tracking-wide mb-2">Resource access</p>
            <div className="border border-border-card rounded-lg overflow-hidden">
              <div className="grid grid-cols-2 px-3 py-2 bg-surface-page">
                <span className="text-xs text-text-muted">Resource</span>
                <span className="text-xs text-text-muted">Permission</span>
              </div>
              {Object.entries(resourceAccess).map(([resource, access]) => (
                <div key={resource} className="grid grid-cols-2 px-3 py-2.5 border-t border-border-card items-center">
                  <span className="text-sm text-text-primary">{resource}</span>
                  <div className="flex flex-wrap gap-1">
                    {access.roles.map(role => (
                      <span key={role} className="px-1.5 py-0.5 rounded bg-brand/10 text-brand text-xs">{role}</span>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
        {groups.length > 0 && (
          <div>
            <p className="text-xs text-text-muted uppercase tracking-wide mb-2">Groups</p>
            <div className="flex flex-wrap gap-2">
              {groups.map(group => (
                <span key={group} className="px-2.5 py-1 rounded-full border border-border-card text-xs text-text-muted">{group}</span>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
