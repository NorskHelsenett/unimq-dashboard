import { useState, useEffect } from 'react'
import { getCheckerStatus, type CheckerStatus } from '@/services/checkerStatus'

function toTimeAgo(iso: string): string {
  const s = Math.floor((Date.now() - new Date(iso).getTime()) / 1000)
  if (s < 5)   return 'just now'
  if (s < 60)  return `${s}s ago`
  const m = Math.floor(s / 60)
  if (m < 60)  return `${m}m ago`
  return `${Math.floor(m / 60)}h ago`
}

export function useCheckerStatus() {
  const [status, setStatus] = useState<CheckerStatus | null>(null)
  const [timeAgo, setTimeAgo] = useState('')

  useEffect(() => {
    getCheckerStatus().then(setStatus)
    const poll = setInterval(() => getCheckerStatus().then(setStatus), 30_000)
    return () => clearInterval(poll)
  }, [])

  useEffect(() => {
    if (!status?.last_checked) return
    const update = () => setTimeAgo(toTimeAgo(status.last_checked!))
    update()
    const tick = setInterval(update, 5_000)
    return () => clearInterval(tick)
  }, [status?.last_checked])

  return { status, timeAgo }
}
