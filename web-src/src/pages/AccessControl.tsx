import '../index.css'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RequireAuth } from '@/auth/RequireAuth'
import { Layout } from '@/components/layout/Layout'
import { AclCard } from '@/components/access-control/AclCard'
import { useACLs } from '@/hooks/useAcls'
import { useVhost } from '@/hooks/useVhost'


export function AccessControl() {
    const {acls, loading, error, refetch } = useACLs() 
    const {vhosts, loading: vhostLoading } = useVhost()


    return (
        <Layout>
           {loading || vhostLoading ? <p>Loading...</p> : error ? <p>Error: {error.message}</p> : <AclCard acls={acls} vhosts={vhosts} onRefresh={refetch} />}
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