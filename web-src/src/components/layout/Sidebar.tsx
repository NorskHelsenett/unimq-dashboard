import { NAV_ITEMS } from "@/lib/navItems"
import { LogoLink } from "./LogoLink"
import { cn } from "@/lib/utils"
import { Vhosts } from "@/types/vhosts"
import { User } from "lucide-react"
import { Selector, SelectorTrigger, SelectorContent, SelectorItem, SelectorValue } from "../ui/selector"
import { useAuth } from "react-oidc-context"
import { useSidebarResize } from "@/hooks/useSidebarResize"

function isActive(itemHref: string, currentPath: string): boolean {
    if (itemHref === '/') return currentPath === '/'
    return currentPath.startsWith(itemHref)
}

export function Sidebar() {
    const currentPath = window.location.pathname
    const currentVhost = new URLSearchParams(window.location.search).get('vhost')
    const vhostParam = currentVhost ? `?vhost=${encodeURIComponent(currentVhost)}` : ''
    const { width, collapsed, onMouseDown } = useSidebarResize()
    const auth = useAuth()

    return (

        <aside style={{ width }} className="overflow-hidden">
            <nav className="relative flex flex-col pt-4 border-r border-border-sidebar h-screen sticky top-0 overflow-y-auto">
            <div
                className="absolute right-0 top-0 h-full w-1 cursor-col-resize"
                onMouseDown={onMouseDown}
            />
                    <LogoLink collapsed={collapsed} />
                    <div className="flex flex-col gap-1 pt-10">
                        {NAV_ITEMS.map((item) => {
                            const active = isActive(item.href, currentPath)
                            return (
                                <a href={item.href + vhostParam} className={cn(
                                    "flex items-center py-2",
                                    collapsed ? "justify-center px-2" : "gap-2 px-4",
                                    active 
                                        ? "bg-surface-sidebar-active text-text-sidebar-active border-l-3 border-brand font-bold" 
                                        : "text-text-sidebar hover:bg-surface-sidebar-active hover:text-text-sidebar-active"
                                )}>
                                    {item.icon && <item.icon size={18} className="shrink-0" />}
                                    {!collapsed && <span className="truncate">{item.label}</span>}
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
                                "flex items-center py-2 rounded-md w-full border-none shadow-none",
                                collapsed ? "justify-center px-2" : "gap-2 px-4",
                                currentPath === '/profile'
                                    ? "bg-surface-sidebar-active text-text-sidebar-active border-l-3 border-brand"
                                    : "text-text-sidebar hover:bg-surface-sidebar-active hover:text-text-sidebar-active"
                            )}>
                                <User size={18} className="shrink-0" />
                                {!collapsed && <SelectorValue placeholder="Account" />}
                            </SelectorTrigger>
                            <SelectorContent>
                                <SelectorItem value="/profile">Profile</SelectorItem>
                                <SelectorItem value="signout">Sign out</SelectorItem>
                            </SelectorContent>
                        </Selector>
                    </div>
                </nav>
        </aside>
    )
}