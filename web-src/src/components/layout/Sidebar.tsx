import { NAV_ITEMS } from "@/lib/navItems"
import { LogoLink } from "./LogoLink"
import { cn } from "@/lib/utils"
import { VhostSelector } from "./VhostSelector"
import { Vhosts } from "@/types/vhosts"
import { User } from "lucide-react"
import { Selector, SelectorTrigger, SelectorContent, SelectorItem, SelectorValue } from "../ui/selector"
import { useAuth } from "react-oidc-context"

function isActive(itemHref: string, currentPath: string): boolean {
    if (itemHref === '/') return currentPath === '/'
    return currentPath.startsWith(itemHref)
}

export function Sidebar({ Vhosts, Selected }: Vhosts ) {
    const currentPath = window.location.pathname
    const currentVhost = new URLSearchParams(window.location.search).get('vhost')
    const vhostParam = currentVhost ? `?vhost=${encodeURIComponent(currentVhost)}` : ''
    const auth = useAuth()

    return (
        <nav className="relative flex flex-col pt-4 border-r border-border-sidebar h-screen sticky top-0 overflow-y-auto">
            <LogoLink />
            <div className="p-4 flex gap-2 items-center border-y my-4">
                <span>Vhost:</span>
                <VhostSelector Vhosts={Vhosts} Selected={Selected} />
            </div>
            <div className="flex flex-col gap-1">
                {NAV_ITEMS.map((item) => {
                    const active = isActive(item.href, currentPath)
                    return (
                        <a href={item.href + vhostParam} className={cn(
                            "py-2 px-4", 
                            active 
                                ? "bg-surface-sidebar-active text-text-sidebar-active border-l-3 border-brand font-bold" 
                                : "text-text-sidebar hover:bg-surface-sidebar-active hover:text-text-sidebar-active"
                        )}>
                            {item.label}
                        </a>
                    )
                })}
            </div>
            <div className="mt-auto p-4">
                <Selector onValueChange={(value) => {
                    if (value === 'signout') {
                        auth.signoutRedirect()
                    } else {
                        window.location.href = value
                    }
                }}>
                    <SelectorTrigger className={cn(
                        "flex items-center gap-2 py-2 px-4 rounded-md w-full border-none shadow-none",
                        currentPath === '/profile'
                            ? "bg-surface-sidebar-active text-text-sidebar-active border-l-3 border-brand"
                            : "text-text-sidebar hover:bg-surface-sidebar-active hover:text-text-sidebar-active"
                    )}>
                        <User size={18} />
                        <SelectorValue placeholder="Account" />
                    </SelectorTrigger>
                    <SelectorContent>
                        <SelectorItem value="/profile">Profile</SelectorItem>
                        <SelectorItem value="signout">Sign out</SelectorItem>
                    </SelectorContent>
                </Selector>
            </div>
        </nav>
    )
}