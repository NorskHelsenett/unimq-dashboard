import type { ReactNode } from 'react'
import { Sidebar } from './Sidebar'
import { TooltipProvider } from '../ui/tooltip'
import { SessionExpiryBanner } from './SessionExpiryBanner'
import { LayoutFooter } from './LayoutFooter'
import { TopBar } from './TopBar'

interface LayoutProps {
  children: ReactNode
}

export function Layout({ children }: LayoutProps) {
  return (
    <TooltipProvider>
        <div className="flex min-h-screen bg-surface-page">
            <Sidebar/>
            <div className="flex flex-col flex-1 min-w-0">
                <SessionExpiryBanner />
                <TopBar />
                <main className="flex-1 p-4 mx-6 mb-6 rounded-lg">
                    {children}
                </main>
                <LayoutFooter />
            </div>
        </div>
    </TooltipProvider>
  )
}