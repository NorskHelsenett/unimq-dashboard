import type { ReactNode } from 'react'
import { Sidebar } from './Sidebar'

interface LayoutProps {
  Vhosts: string[]
  Selected: string
  children: ReactNode
}

export function Layout({ Vhosts, Selected, children }: LayoutProps) {
  return (
    <div className="flex min-h-screen bg-gray-50">
      <Sidebar Vhosts={Vhosts} Selected={Selected} />
      <div className="flex flex-col flex-1 min-w-0">
        <main className="flex-1 p-6 m-6 bg-white border rounded-lg">
          {children}
        </main>
      </div>
    </div>
  )
}