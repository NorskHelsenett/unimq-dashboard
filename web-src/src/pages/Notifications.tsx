import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '../index.css'
import { RequireAuth } from '@/auth/RequireAuth'
import { Layout } from '@/components/layout/Layout'
import { AlarmCard } from '@/components/notifications/AlarmCard'
import { RecipientCard } from '@/components/notifications/RecipientCard'
import { useVhostNotification } from '@/hooks/useVhostNotification'


const Notifications = () => {
  const { vhosts, selected, notification, loading } = useVhostNotification()
  const rules = notification?.Rules ?? []
  const recipients = notification?.Recipients ?? []

  return (
    <Layout Vhosts={vhosts} Selected={selected}>
      {loading ? (
        <div className="p-8 text-text-muted">Loading...</div>
      ) : (
        <div>
          <h1 className='text-4xl mb-6'>Notifications</h1>
          <div className='max-w-4xl mx-auto flex flex-col gap-4'>
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