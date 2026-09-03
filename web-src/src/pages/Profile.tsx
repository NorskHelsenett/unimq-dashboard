import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '../index.css'
import { RequireAuth } from '@/auth/RequireAuth'
import { Layout } from '@/components/layout/Layout'
import { useAuth } from 'react-oidc-context'
import { ProfileHeroCard } from '@/components/profile/ProfileHeroCard'
import { AccountDetailsCard } from '@/components/profile/AccountDetailsCard'
import { AccessPermissionsCard } from '@/components/profile/AccessPermissionsCard'
import { DEFAULT_ROLES, resolveIdProvider } from '@/components/profile/profileUtils'

const ProfilePage = () => {
  const auth = useAuth()
  const p = auth.user?.profile as Record<string, unknown>

  const name     = p?.name as string | undefined
  const email    = p?.email as string | undefined
  const verified = p?.email_verified as boolean | undefined
  const sub      = p?.sub as string | undefined
  const username = p?.preferred_username as string | undefined
  const iss      = p?.iss as string | undefined
  const iat      = p?.iat as number | undefined
  const exp      = auth.user?.expires_at

  const realmRoles     = (p?.realm_access as { roles?: string[] })?.roles ?? []
  const resourceAccess = (p?.resource_access as Record<string, { roles: string[] }>) ?? {}
  const groups         = (p?.groups as string[]) ?? []
  const displayRoles   = realmRoles.filter(r => !DEFAULT_ROLES.has(r))

  return (
    <Layout>
      <div className="max-w-5xl space-y-6">
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-2xl font-bold text-text-primary">Profile</h1>
            <p className="text-sm text-text-muted mt-1">Manage your account and access.</p>
          </div>
          <button
            onClick={() => auth.signoutRedirect()}
            className="flex items-center gap-2 px-4 py-2 rounded-md border border-destructive/30 text-destructive hover:bg-destructive/15 transition-colors text-sm font-medium"
          >
            Sign out
          </button>
        </div>

        <ProfileHeroCard
          name={name}
          email={email}
          verified={verified}
          primaryRole={displayRoles[0]}
          iat={iat}
          exp={exp}
        />

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <AccountDetailsCard
            name={name}
            email={email}
            verified={verified}
            sub={sub}
            username={username}
            iss={iss}
            idProvider={resolveIdProvider(iss)}
            iat={iat}
            exp={exp}
            rawProfile={auth.user?.profile}
          />
          <AccessPermissionsCard
            displayRoles={displayRoles}
            resourceAccess={resourceAccess}
            groups={groups}
          />
        </div>
      </div>
    </Layout>
  )
}

createRoot(document.getElementById('app')!).render(
  <StrictMode>
    <RequireAuth>
      <ProfilePage />
    </RequireAuth>
  </StrictMode>,
)

