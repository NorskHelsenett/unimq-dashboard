import { useState, useEffect } from 'react'
import { getVhostNotification, getSelectedVhost } from '@/services/notifications'
import type { VhostNotification } from '@/types/notifications'
import { getVhosts } from '@/services/vhosts'

interface UseVhostNotificationResult {
  vhosts: string[]
  selected: string
  notification: VhostNotification | null
  loading: boolean
}

export function useVhostNotification(): UseVhostNotificationResult {
  const [vhosts, setVhosts] = useState<string[]>([])
  const [selected, setSelected] = useState<string>('')
  const [notification, setNotification] = useState<VhostNotification | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    getVhosts()
      .then(names => {
        setVhosts(names)
        const sel = getSelectedVhost(names)
        setSelected(sel)
        return getVhostNotification(sel)
      })
      .then(data => setNotification(data))
      .finally(() => setLoading(false))
  }, [])

  return { vhosts, selected, notification, loading }
}
