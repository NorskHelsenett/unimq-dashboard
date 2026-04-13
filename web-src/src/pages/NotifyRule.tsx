import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '../index.css'
import { getPageData } from '@/lib/pageData'
import { Layout } from '@/components/layout/Layout'
import { EditAlarm } from '@/components/notifications/EditAlarm'
import { AlarmProps } from '@/components/notifications/AlarmCard'

interface NotifyRuleData {
  Vhost: string
  Rule: AlarmProps
  Msg: string
}

const data = getPageData<NotifyRuleData>()

const root = document.getElementById('app')
if (!root) throw new Error('Missing #app mount point')

createRoot(root).render(
  <StrictMode>
    <Layout Vhosts={[data.Vhost]} Selected={data.Vhost}>
      <div className="max-w-4xl mx-auto">
        <a href={`/notifications?vhost=${encodeURIComponent(data.Vhost)}`} className="text-sm text-text-muted hover:text-text-primary mb-4 inline-block">← Back to alarms</a>
        <EditAlarm alarm={data.Rule} />
      </div>
    </Layout>
  </StrictMode>,
)