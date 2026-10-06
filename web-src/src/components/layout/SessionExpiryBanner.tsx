import { useEffect, useState } from 'react'
import { AlertTriangle, RefreshCw, LogIn } from 'lucide-react'
import { useSession, useSessionRefresh } from '@/auth/RequireAuth'
import { startLogin } from '@/lib/apiClient'

/** How long before expiry the warning appears. */
const WARN_WITHIN_SECONDS = 5 * 60

/** How often the remaining time is recomputed. */
const TICK_MS = 15_000

function secondsUntil(unixSeconds: number): number {
  return Math.floor(unixSeconds - Date.now() / 1000)
}

function formatRemaining(seconds: number): string {
  if (seconds >= 120) return `${Math.floor(seconds / 60)} minutes`
  if (seconds >= 60) return 'a minute'
  return 'less than a minute'
}

/**
 * Warns the user before their session lapses.
 *
 * The expiry comes from the "exp" claim the backend reports on /v1/profile —
 * the frontend no longer holds a token to inspect. RequireAuth renews the
 * session automatically a minute before it lapses, so this banner is the
 * fallback for when that failed, or when the tab was suspended through it.
 */
export function SessionExpiryBanner() {
  const { expires_at: expiresAt } = useSession()
  const refresh = useSessionRefresh()
  const [remaining, setRemaining] = useState(() => secondsUntil(expiresAt))
  const [refreshing, setRefreshing] = useState(false)
  const [failed, setFailed] = useState(false)

  useEffect(() => {
    setFailed(false)
    setRemaining(secondsUntil(expiresAt))
    const id = setInterval(() => setRemaining(secondsUntil(expiresAt)), TICK_MS)
    return () => clearInterval(id)
  }, [expiresAt])

  const handleRefresh = async () => {
    setRefreshing(true)
    try {
      // On success RequireAuth swaps in a new profile, which re-runs the effect
      // above and hides the banner.
      if (!(await refresh())) setFailed(true)
    } finally {
      setRefreshing(false)
    }
  }

  // A missing or far-off expiry means there is nothing to say.
  if (!expiresAt || remaining > WARN_WITHIN_SECONDS) return null

  const expired = remaining <= 0

  return (
    <div className={`flex items-center justify-between gap-4 px-6 py-3 text-sm ${expired ? 'bg-destructive/10 border-b border-destructive/20 text-destructive' : 'bg-amber-500/10 border-b border-amber-500/20 text-amber-600'}`}>
      <span className="flex items-center gap-2 font-medium">
        <AlertTriangle size={15} className="shrink-0" />
        {expired
          ? 'Your session has expired. Please sign in again to continue.'
          : failed
            ? 'Your session could not be renewed. Please sign in again.'
            : `Your session expires in ${formatRemaining(remaining)}.`}
      </span>
      <div className="flex items-center gap-2 shrink-0">
        {!expired && !failed && (
          <button
            onClick={() => void handleRefresh()}
            disabled={refreshing}
            className="flex items-center gap-1.5 px-3 py-1 rounded-md border border-amber-500/40 hover:bg-amber-500/10 transition-colors disabled:opacity-50 text-xs font-medium"
          >
            <RefreshCw size={13} className={refreshing ? 'animate-spin' : ''} />
            Refresh session
          </button>
        )}
        <button
          onClick={startLogin}
          className={`flex items-center gap-1.5 px-3 py-1 rounded-md border transition-colors text-xs font-medium ${expired || failed ? 'border-destructive/40 hover:bg-destructive/10' : 'border-amber-500/40 hover:bg-amber-500/10'}`}
        >
          <LogIn size={13} />
          Sign in again
        </button>
      </div>
    </div>
  )
}
