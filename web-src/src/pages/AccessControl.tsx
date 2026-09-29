import '../index.css'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RequireAuth } from '@/auth/RequireAuth'
import { Layout } from '@/components/layout/Layout'


export function AccessControl() {
    return (
        <Layout>
           <h1>Access Control</h1>
        </Layout>
    )
}
const root = document.getElementById('app')
if (!root) throw new Error('Missing #app mount point')
    
createRoot(root).render(
  <StrictMode>
    <RequireAuth>
      <AccessControl />
    </RequireAuth>
  </StrictMode>,
)