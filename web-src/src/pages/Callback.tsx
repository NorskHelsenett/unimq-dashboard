import { StrictMode } from "react"
import { createRoot } from "react-dom/client"
import { AuthProvider, useAuth } from "react-oidc-context"
import { oidcConfig } from "@/auth/auth.config"
import "../index.css"

function CallbackPage() {
  const auth = useAuth()

  if (auth.isLoading) {
    return (
      <div className="flex items-center justify-center h-screen text-muted-foreground">
        Completing login…
      </div>
    )
  }

  if (auth.error) {
    return (
      <div className="flex flex-col items-center justify-center h-screen gap-2">
        <p className="text-destructive">Login failed: {auth.error.message}</p>
        <a href="/" className="underline text-sm">Go back</a>
      </div>
    )
  }

  return null
}

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <AuthProvider
      {...oidcConfig}
      onSigninCallback={() => {
        window.location.replace("/")
      }}
    >
      <CallbackPage />
    </AuthProvider>
  </StrictMode>
)
