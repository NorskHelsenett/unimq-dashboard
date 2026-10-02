import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '../index.css'
import { RequireAuth, useSession } from '@/auth/RequireAuth'
import { Layout } from '@/components/layout/Layout'
import { logout } from '@/lib/apiClient'
import { ProfileHeroCard } from '@/components/profile/ProfileHeroCard'
import { AccountDetailsCard } from '@/components/profile/AccountDetailsCard'
import { AccessPermissionsCard } from '@/components/profile/AccessPermissionsCard'
import { DEFAULT_ROLES, resolveIdProvider } from '@/components/profile/profileUtils'

const ProfilePage = () => {
  const profile = useSession()

  const name     = profile.username || undefined
  const email    = profile.email || undefined
  const verified = profile.email_verified
  const sub      = profile.subject || undefined
  const username = profile.username || undefined
  const iss      = profile.issuer || undefined
  const iat      = profile.issued_at || undefined
  const exp      = profile.expires_at || undefined

  // Realm and resource roles are Keycloak-specific and are not forwarded by
  // the backend profile endpoint. Groups carry the access model today.
  const realmRoles: string[] = []
  const resourceAccess: Record<string, { roles: string[] }> = {}
  const groups         = profile.groups ?? []
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
            onClick={() => void logout()}
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
            rawProfile={profile}
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

