import { useEffect } from "react"
import { AuthProvider, useAuth } from "react-oidc-context"
import { oidcConfig } from "./auth.config"
import { setAuthToken } from "@/lib/apiClient"

function AuthGuard({ children }: { children: React.ReactNode }) {
  const auth = useAuth()

  setAuthToken(auth.user?.access_token ?? null)

  useEffect(() => {
    return auth.events.addSilentRenewError(() => {
      auth.signinRedirect()
    })
  }, [auth.events, auth.signinRedirect])

  useEffect(() => {
    if (!auth.isLoading && !auth.isAuthenticated && !auth.error) {
      auth.signinRedirect()
    }
  }, [auth.isLoading, auth.isAuthenticated, auth.error])

  if (auth.isLoading) {
    return (
      <div className="flex items-center justify-center h-screen text-muted-foreground">
        Logging in…
      </div>
    )
  }

  if (auth.error) {
    return (
      <div className="flex flex-col items-center justify-center h-screen gap-2">
        <p className="text-destructive">Authentication error: {auth.error.message}</p>
        <button
          className="underline text-sm"
          onClick={() => auth.signinRedirect()}
        >
          Try again
        </button>
      </div>
    )
  }

  if (!auth.isAuthenticated) {
    return null
  }

  return <>{children}</>
}

export function RequireAuth({ children }: { children: React.ReactNode }) {
  return (
    <AuthProvider
      {...oidcConfig}
      onSigninCallback={() => {
        window.history.replaceState({}, document.title, window.location.pathname)
      }}
    >
      <AuthGuard>{children}</AuthGuard>
    </AuthProvider>
  )
}
