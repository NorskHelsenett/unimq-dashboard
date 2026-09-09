import { useState, useEffect } from 'react'
import { getVhostNotification, getSelectedVhost, getVhostNotificationById } from '@/services/notifications'
import type { AlarmProps, VhostNotification } from '@/types/notifications'
import { getVhosts } from '@/services/vhosts'

interface UseVhostNotificationResult {
  vhosts: string[]
  selected: string
  notification: VhostNotification | null
  loading: boolean
}

interface UseVhostNotificationByIdResult {
  vhost: string
  ruleId: string
  alarm: AlarmProps | null
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

export function useVhostNotificationById(vhost: string, ruleId: string): UseVhostNotificationByIdResult {
  const [alarm, setAlarm] = useState<AlarmProps | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!vhost || !ruleId) {
      setAlarm(null)
      setLoading(false)
      return
    }

    setLoading(true)
    getVhostNotificationById(vhost, ruleId)
      .then(data => setAlarm(data))
      .finally(() => setLoading(false))
  }, [vhost, ruleId])

  return { vhost, ruleId, alarm, loading }
}
