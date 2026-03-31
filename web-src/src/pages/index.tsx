import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '../index.css'
import { getPageData } from '@/lib/pageData'
import { Layout } from '@/components/layout/Layout'

interface IndexData {
  Vhosts: string[]
  Selected: string
}

const data = getPageData<IndexData>()

const root = document.getElementById('app')
if (!root) throw new Error('Missing #app mount point')

createRoot(document.getElementById('app')!).render(
  <StrictMode>
    <Layout Vhosts={data.Vhosts} Selected={data.Selected}>
      <p className="text-sm text-gray-500">Overview coming soon.</p>
    </Layout>
  </StrictMode>,
)