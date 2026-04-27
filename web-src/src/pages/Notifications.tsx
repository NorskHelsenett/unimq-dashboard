import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '../index.css'
import { getPageData } from '@/lib/pageData'
import { Layout } from '@/components/layout/Layout'
import { AlarmCard, AlarmProps } from '@/components/notifications/AlarmCard'
import { RecipientCard, RecipientsProps } from '@/components/notifications/RecipientCard'


interface NotificationData {
  Vhosts: string[]
  Selected: string
  Recipients: RecipientsProps[]
  Rules: AlarmProps[]
}

const data = getPageData<NotificationData>()

const root = document.getElementById('app')
if (!root) throw new Error('Missing #app mount point')

const NotificationsPage = () => {

  return(
    <div>
      <h1 className='text-4xl mb-6'>Notifications</h1>
      <div className='max-w-4xl mx-auto flex flex-col gap-4'>
        <AlarmCard existingAlarms={data.Rules || []} vhost={data.Selected} />
        <RecipientCard existingRecipients={data.Recipients || null} vhost={data.Selected} />
      </div>    
    </div>
  )
}

createRoot(document.getElementById('app')!).render(
  <StrictMode>
    <Layout Vhosts={data.Vhosts} Selected={data.Selected}>
        <NotificationsPage />
    </Layout>
  </StrictMode>,
)