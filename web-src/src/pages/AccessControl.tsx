import '../index.css'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RequireAuth } from '@/auth/RequireAuth'
import { Layout } from '@/components/layout/Layout'
import { AclCard } from '@/components/access-control/AclCard'
import { useACLs } from '@/hooks/useAcls'
import { useVhost } from '@/hooks/useVhost'
import { Button } from '@/components/ui/button'
import { CircleAlert, RefreshCw } from 'lucide-react'


export function AccessControl() {
    const {acls, loading, error, refetch } = useACLs() 
    const {vhosts, loading: vhostLoading } = useVhost()


    return (
        <Layout>
         {loading || vhostLoading ? <p>Loading...</p> : error ? (
           <div className="flex min-h-[50vh] items-center justify-center px-4">
             <section role="alert" className="w-full max-w-lg rounded-lg border border-destructive/20 bg-surface-card p-6 shadow-sm sm:p-8">
               <div className="mb-5 flex size-11 items-center justify-center rounded-md bg-destructive/10 text-destructive">
                 <CircleAlert aria-hidden="true" size={22} />
               </div>
               <p className="text-xs font-semibold uppercase text-text-muted">Access control</p>
               <h1 className="mt-2 text-xl font-semibold text-text-primary">We couldn’t load access rules</h1>
               <p className="mt-2 text-sm text-text-secondary">
                 The access rules could not be retrieved. Check your connection and try again.
               </p>
               <p className="mt-3 break-words text-xs text-text-muted">{error.message}</p>
               <Button className="mt-6" onClick={refetch}>
                 <RefreshCw aria-hidden="true" />
                 Try again
               </Button>
             </section>
           </div>
         ) : <AclCard acls={acls} vhosts={vhosts} onRefresh={refetch} />}
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