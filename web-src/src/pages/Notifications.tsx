import { StrictMode, useState, useEffect } from 'react'
import { createRoot } from 'react-dom/client'
import '../index.css'
import { RequireAuth } from '@/auth/RequireAuth'
import { Layout } from '@/components/layout/Layout'
import { AlarmCard, AlarmProps } from '@/components/notifications/AlarmCard'
import { RecipientCard, RecipientsProps } from '@/components/notifications/RecipientCard'

interface ApiResponse<T> {
  code: number
  message: string
  body: T
}

interface VhostObj {
  name: string
}

interface VhostNotification {
  Name: string
  Recipients: RecipientsProps[]
  Rules: AlarmProps[]
  Notified: boolean
}

function getSelectedVhost(vhosts: string[]): string {
  const params = new URLSearchParams(window.location.search)
  const vhost = params.get('vhost')
  return (vhost && vhosts.includes(vhost)) ? vhost : (vhosts[0] ?? '')
}

const NotificationsApp = () => {
  const [vhosts, setVhosts] = useState<string[]>([])
  const [selected, setSelected] = useState<string>('')
  const [rules, setRules] = useState<AlarmProps[]>([])
  const [recipients, setRecipients] = useState<RecipientsProps[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch('/api/v1/vhosts')
      .then(r => r.json())
      .then((res: ApiResponse<VhostObj[]>) => {
        const names = (res.body ?? []).map(v => v.name)
        setVhosts(names)
        const sel = getSelectedVhost(names)
        setSelected(sel)
        return sel
      })
      .then(sel => fetch(`/api/v1/notifications/${encodeURIComponent(sel)}`))
      .then(r => r.ok ? r.json() : Promise.resolve({ body: { Rules: [], Recipients: [] } }))
      .then((res: ApiResponse<VhostNotification>) => {
        setRules(res.body?.Rules ?? [])
        setRecipients(res.body?.Recipients ?? [])
      })
      .finally(() => setLoading(false))
  }, [])

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
      <NotificationsApp />
    </RequireAuth>
  </StrictMode>,
)