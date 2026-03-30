import { createRoot } from 'react-dom/client'
import '../index.css'

function App() {
  return <div />
}

const root = document.getElementById('app')
if (!root) throw new Error('Missing #app mount point')
createRoot(root).render(<App />)