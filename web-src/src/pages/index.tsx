import { createRoot } from 'react-dom/client'
import '../index.css'

function App() {
  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center">
      <div className="bg-white border border-gray-200 rounded-lg p-8 shadow-sm">
        <h1 className="text-xl font-semibold text-gray-900">UniMQ</h1>
        <p className="text-sm text-gray-500 mt-1">Pipeline working ✓</p>
      </div>
    </div>
  )
}

const root = document.getElementById('app')
if (!root) throw new Error('Missing #app mount point')
createRoot(root).render(<App />)