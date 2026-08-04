import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '../index.css'
import { RequireAuth } from '@/auth/RequireAuth'
import { Layout } from '@/components/layout/Layout'
import { EditAlarm } from '@/components/notifications/EditAlarm'
import { useVhostNotification } from '@/hooks/useVhostNotification'

const NotificationRule = () => {
  const { vhosts, selected, notification, loading } = useVhostNotification()
  const ruleId = new URLSearchParams(window.location.search).get('id')
  const alarm = notification?.Rules.find(r => r.id === ruleId) ?? null

  return (
    <Layout Vhosts={vhosts} Selected={selected}>
      <div className="max-w-4xl mx-auto">
        <a href={`/notifications?vhost=${encodeURIComponent(selected)}`} className="text-sm text-text-muted hover:text-text-primary mb-4 inline-block">← Back to alarms</a>
        {loading ? (
          <div className="p-8 text-text-muted">Loading...</div>
        ) : alarm ? (
          <EditAlarm alarm={alarm} vhost={selected} />
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
