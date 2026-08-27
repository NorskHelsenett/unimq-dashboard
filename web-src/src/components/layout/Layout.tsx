import type { ReactNode } from 'react'
import { Sidebar } from './Sidebar'
import { TooltipProvider } from '../ui/tooltip'

interface LayoutProps {
  children: ReactNode
}

export function Layout({ children }: LayoutProps) {
  return (
    <TooltipProvider>
        <div className="flex min-h-screen bg-surface-page">
            <Sidebar/>
            <div className="flex flex-col flex-1 min-w-0">
                <main className="flex-1 p-4 m-6 rounded-lg">
                    {children}
                </main>
            </div>
        </div>
    </TooltipProvider>
  )
}