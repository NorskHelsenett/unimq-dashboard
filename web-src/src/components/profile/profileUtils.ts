export const DEFAULT_ROLES = new Set([
  'offline_access',
  'uma_authorization',
  'default-roles-unimq',
  'default-roles',
])

export function getInitials(name?: string): string {
  if (!name) return '?'
  return name.split(' ').filter(Boolean).slice(0, 2).map(n => n[0].toUpperCase()).join('')
}

export function formatAgo(ts: number): string {
  const diff = Math.floor(Date.now() / 1000 - ts)
  if (diff < 60) return `${diff}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
  return `${Math.floor(diff / 86400)}d ago`
}

export function formatIn(ts: number): string {
  const diff = Math.floor(ts - Date.now() / 1000)
  if (diff <= 0) return 'Expired'
  if (diff < 60) return `in ${diff}s`
  if (diff < 3600) return `in ${Math.floor(diff / 60)}m`
  if (diff < 86400) return `in ${Math.floor(diff / 3600)}h`
  return `in ${Math.floor(diff / 86400)}d`
}

export function formatDate(ts: number): string {
  return new Date(ts * 1000).toLocaleString('en-GB', {
    day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit',
  })
}

export function resolveIdProvider(iss?: string): string | undefined {
  if (!iss) return undefined
  if (iss.includes('realms') || iss.includes('keycloak')) return 'Keycloak'
  try { return new URL(iss).hostname } catch { return iss }
}
