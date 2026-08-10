import type { ReactNode } from 'react'
import { Sidebar } from './Sidebar'
import { TooltipProvider } from '../ui/tooltip'

interface LayoutProps {
  Vhosts: string[]
  Selected: string
  MaintenanceMode?: boolean
  children: ReactNode
}

export function Layout({ Vhosts, Selected, MaintenanceMode=false, children }: LayoutProps) {
  return (
    <TooltipProvider>
        <div className="flex min-h-screen bg-gray-50">
            <Sidebar Vhosts={Vhosts} Selected={Selected} MaintenanceMode={MaintenanceMode} />
            <div className="flex flex-col flex-1 min-w-0">
                <main className="flex-1 p-6 m-6 rounded-lg">
                    {children}
                </main>
            </div>
        </div>
    </TooltipProvider>
  )
}