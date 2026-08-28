import { useEffect, useState } from 'react'
import { useAuth } from 'react-oidc-context'
import { AlertTriangle, RefreshCw, LogIn } from 'lucide-react'

type BannerState = 'expiring' | 'expired' | null

export function SessionExpiryBanner() {
  const auth = useAuth()
  const [state, setState] = useState<BannerState>(auth.user?.expired ? 'expired' : null)
  const [refreshing, setRefreshing] = useState(false)

  useEffect(() => auth.events.addAccessTokenExpiring(() => setState(s => s !== 'expired' ? 'expiring' : s)), [auth.events])
  useEffect(() => auth.events.addAccessTokenExpired(() => setState('expired')), [auth.events])

  const handleRefresh = async () => {
    setRefreshing(true)
    try {
      await auth.signinSilent()
      setState(null)
    } catch {
      setState('expired')
    } finally {
      setRefreshing(false)
    }
  }

  if (!state) return null

  const expired = state === 'expired'

  return (
    <div className={`flex items-center justify-between gap-4 px-6 py-3 text-sm ${expired ? 'bg-destructive/10 border-b border-destructive/20 text-destructive' : 'bg-amber-500/10 border-b border-amber-500/20 text-amber-600'}`}>
      <span className="flex items-center gap-2 font-medium">
        <AlertTriangle size={15} className="shrink-0" />
        {expired
          ? 'Your session has expired. Please sign in again to continue.'
          : 'Your session is about to expire.'}
      </span>
      <div className="flex items-center gap-2 shrink-0">
        {!expired && (
          <button
            onClick={handleRefresh}
            disabled={refreshing}
            className="flex items-center gap-1.5 px-3 py-1 rounded-md border border-amber-500/40 hover:bg-amber-500/10 transition-colors disabled:opacity-50 text-xs font-medium"
          >
            <RefreshCw size={13} className={refreshing ? 'animate-spin' : ''} />
            Refresh session
          </button>
        )}
        <button
          onClick={() => auth.signinRedirect()}
          className={`flex items-center gap-1.5 px-3 py-1 rounded-md border transition-colors text-xs font-medium ${expired ? 'border-destructive/40 hover:bg-destructive/10' : 'border-amber-500/40 hover:bg-amber-500/10'}`}
        >
          <LogIn size={13} />
          Sign in again
        </button>
      </div>
    </div>
  )
}
