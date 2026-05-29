import { StrictMode, useState } from 'react'
import { createRoot } from 'react-dom/client'
import '../index.css'
import { RequireAuth } from '@/auth/RequireAuth'
import { getPageData } from '@/lib/pageData'
import { Layout } from '@/components/layout/Layout'
import { useAuth } from 'react-oidc-context'
import { Eye, EyeOff } from 'lucide-react'

interface ProfileData {
  Vhosts: string[]
  Selected: string
}

const data = getPageData<ProfileData>()

const getInitials = (name?: string) => {
  if (!name) return '?'
  return name
    .split(' ')
    .filter(Boolean)
    .slice(0, 2)
    .map((n) => n[0].toUpperCase())
    .join('')
}

const ProfilePage = () => {
  const auth = useAuth()
  const user = auth.user?.profile
  const [showSub, setShowSub] = useState(false)

  return (
    <div className='max-w-sm flex flex-col gap-6'>
      <div className='flex items-center gap-4'>
        <div className='size-14 rounded-full bg-brand/10 border border-brand/20 flex items-center justify-center flex-shrink-0'>
          <span className='text-xl font-semibold text-brand'>{getInitials(user?.name)}</span>
        </div>
        <div className='min-w-0'>
          <h1 className='text-xl font-semibold text-text-primary truncate'>{user?.name ?? '—'}</h1>
          {user?.email && (
            <p className='text-sm text-text-muted truncate'>{user.email}</p>
          )}
        </div>
      </div>

      {/* Info card */}
      <div className='flex flex-col divide-y divide-border-card border border-border-card rounded-lg bg-surface-card'>
        {user?.name && (
          <div className='flex items-center justify-between px-4 py-3'>
            <span className='text-xs text-text-muted uppercase tracking-wide'>Name</span>
            <span className='text-sm text-text-primary font-medium'>{user.name}</span>
          </div>
        )}
        {user?.email && (
          <div className='flex items-center justify-between px-4 py-3'>
            <span className='text-xs text-text-muted uppercase tracking-wide'>Email</span>
            <span className='text-sm text-text-primary'>{user.email}</span>
          </div>
        )}
        {user?.sub && (
          <div className='flex items-center justify-between px-4 py-3 gap-4'>
            <span className='text-xs text-text-muted uppercase tracking-wide flex-shrink-0'>User ID</span>
            <div className='flex items-center gap-2 min-w-0'>
              <span className='text-xs text-text-muted font-mono truncate'>
                {showSub ? user.sub : '••••••••••••'}
              </span>
              <button
                onClick={() => setShowSub((v) => !v)}
                className='text-xs text-text-muted hover:text-text-primary flex-shrink-0 transition-colors'
              >
                {showSub ? <Eye size={16} /> : <EyeOff size={16} />}
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Sign out */}
      <div className='border-t border-border-card pt-4'>
        <button
          onClick={() => auth.signoutRedirect()}
          className='w-full px-4 py-2 rounded-md border border-destructive/30 text-destructive hover:bg-destructive/5 text-sm font-medium transition-colors'
        >
          Sign out
        </button>
      </div>
    </div>
  )
}

createRoot(document.getElementById('app')!).render(
  <StrictMode>
    <RequireAuth>
      <Layout Vhosts={data.Vhosts} Selected={data.Selected}>
        <ProfilePage />
      </Layout>
    </RequireAuth>
  </StrictMode>,
)
