import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '../index.css'
import { getPageData } from '@/lib/pageData'
import { Layout } from '@/components/layout/Layout'
import { LimitsCard } from '@/components/overview/LimitsCard'
import { QueueSizeInfoCard } from '@/components/overview/QueueSizeInfoCard'
import { QueuesCard } from '@/components/overview/QueuesCard'
import { ClusterResourceCard } from '@/components/overview/ClusterResourceCard'
import { VhostResourceCard } from '@/components/overview/VhostResourceCard'
import { SizeDistributionCard } from '@/components/overview/SizeDistributionCard'

interface Metrics {
  connections: number
  channels: number
  queues: number
  unacked: number
}

interface Limits {
  MaxConnections: number
  MaxQueues: number
}

interface IndexData {
  Vhosts: string[]
  Selected: string
  Metrics: Metrics | null
  Limits: Limits
}

const data = getPageData<IndexData>()

const root = document.getElementById('app')
if (!root) throw new Error('Missing #app mount point')

const MainPage = () => {
  const selected = data.Selected
  const metrics = data.Metrics
  
  return (
    <div className='text-text-primary text-base'>
      <h1 className='text-4xl mb-6'>{selected}</h1>
      <div className='flex gap-8 items-end flex-wrap'> 
        {metrics ? (
          <LimitsCard
            connections={metrics.connections}
            channels={metrics.channels}
            queues={metrics.queues}
            unacked={metrics.unacked}
            maxConnections={data.Limits.MaxConnections}
            maxQueues={data.Limits.MaxQueues}
          />
        ) : (
          <p className="text-sm text-text-muted">No metrics available.</p>
        )}
        <QueueSizeInfoCard />
        <QueuesCard vhost={selected} />
        <SizeDistributionCard vhost={selected} />
        <div className='flex gap-4'>
          <ClusterResourceCard />
          <VhostResourceCard vhost={selected} />
        </div>
      </div>
    </div>
  )
}

createRoot(document.getElementById('app')!).render(
  <StrictMode>
    <Layout Vhosts={data.Vhosts} Selected={data.Selected}>
      <MainPage />
    </Layout>
  </StrictMode>,
)