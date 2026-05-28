import { createRoot } from 'react-dom/client'
import { StrictMode } from 'react'
import '../index.css'
import { RequireAuth } from '@/auth/RequireAuth'

function App() {
  return <div />
}

const root = document.getElementById('app')
if (!root) throw new Error('Missing #app mount point')
createRoot(root).render(
  <StrictMode>
    <RequireAuth>
      <App />
    </RequireAuth>
  </StrictMode>
)