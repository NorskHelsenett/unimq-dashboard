import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '../index.css'
import { RequireAuth } from '@/auth/RequireAuth'
import { Layout } from '@/components/layout/Layout'
import { EditAlarm } from '@/components/notifications/EditAlarm'
import { useVhostNotification, useVhostNotificationById } from '@/hooks/useVhostNotification'

const NotificationRule = () => {
  const params = new URLSearchParams(window.location.search)
  const ruleId = params.get('id') ?? ''
  const { selected, loading: vhostsLoading } = useVhostNotification()
  const vhost = params.get('vhost') || selected
  const { alarm, loading: alarmLoading } = useVhostNotificationById(vhost, ruleId)
  const loading = vhostsLoading || alarmLoading

  return (
    <Layout>
      <div className="max-w-4xl mx-auto">
        <a href={`/notifications?vhost=${encodeURIComponent(selected)}`} className="text-sm text-text-muted hover:text-text-primary mb-4 inline-block">
        ← Back to alarms
        </a>
        {loading ? (
          <div className="p-8 text-text-muted">Loading...</div>
        ) : alarm ? (
          <EditAlarm alarm={alarm} vhost={vhost} />
        ) : (
          <p className="text-sm text-text-muted">Alarm not found.</p>
        )}
      </div>
    </Layout>
  )
}

const root = document.getElementById('app')
if (!root) throw new Error('Missing #app mount point')

createRoot(root).render(
  <StrictMode>
    <RequireAuth>
      <NotificationRule />
    </RequireAuth>
  </StrictMode>,
) 
