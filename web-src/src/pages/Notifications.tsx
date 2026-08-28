import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '../index.css'
import { RequireAuth } from '@/auth/RequireAuth'
import { Layout } from '@/components/layout/Layout'
import { AlarmCard } from '@/components/notifications/AlarmCard'
import { RecipientCard } from '@/components/notifications/RecipientCard'
import { useVhostNotification } from '@/hooks/useVhostNotification'
import { LiveDataWidget } from '@/components/dashboard/LiveDataWidget'


const Notifications = () => {
  const { selected, notification, loading } = useVhostNotification()
  const rules = notification?.Rules ?? []
  const recipients = notification?.Recipients ?? []

  return (
    <Layout>
      {loading ? (
        <div className="p-8 text-text-muted">Loading...</div>
      ) : (
            <div className="space-y-6">
                <div>
                  <h1 className='text-3xl tracking-tight'>Notifications</h1>
                </div>
     
          <div className='mx-auto flex flex-col mr-20 gap-4 pt-4'>
            <AlarmCard existingAlarms={rules} vhost={selected} />
            <RecipientCard existingRecipients={recipients} vhost={selected} />
          </div>
        </div>
      )}
    </Layout>
  )
}

const root = document.getElementById('app')
if (!root) throw new Error('Missing #app mount point')

createRoot(root).render(
  <StrictMode>
    <RequireAuth>
      <Notifications />
    </RequireAuth>
  </StrictMode>,
)