import { NAV_ITEMS } from "@/lib/navItems"
import { LogoLink } from "./LogoLink"
import { cn } from "@/lib/utils"
import { User, Sun, Moon } from "lucide-react"
import { Selector, SelectorTrigger, SelectorContent, SelectorItem } from "../ui/selector"
import { useAuth } from "react-oidc-context"
import { useSidebarResize } from "@/hooks/useSidebarResize"
import { useTheme } from "@/hooks/useTheme"

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
    const { theme, toggle } = useTheme()

    return (
        <aside style={{ width }} className="relative overflow-x-auto">
            <nav className="h-full flex flex-col pt-4 border-r border-border-sidebar">
            <div
                className="absolute right-0 top-0 h-full w-1 cursor-col-resize"
                onMouseDown={onMouseDown}
            />
                    <LogoLink collapsed={collapsed} />
                    <div className="flex flex-col gap-1 pt-10 flex-1 overflow-y-auto">
                        {NAV_ITEMS.map((item) => {
                            const active = isActive(item.href, currentPath)
                            return (
                                <a key={item.href} href={item.href + vhostParam} className={cn(
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
                        {collapsed ? (
                            <div className="flex flex-col items-center gap-1">
                                <button
                                    onClick={toggle}
                                    className="flex items-center justify-center p-2 rounded-md hover:bg-surface-sidebar-active text-text-sidebar hover:text-text-sidebar-active"
                                >
                                    {theme === "dark" ? <Sun size={18} /> : <Moon size={18} />}
                                </button>
                                <Selector onValueChange={(value) => {
                                    if (value === 'signout') {
                                        auth.signoutRedirect()
                                    } else {
                                        window.location.href = value
                                    }
                                }}>
                                    <SelectorTrigger hideChevron className={cn(
                                        "flex items-center justify-center p-2 rounded-md w-auto h-auto border-none shadow-none",
                                        currentPath === '/profile'
                                            ? "bg-surface-sidebar-active border-l-3"
                                            : "hover:bg-surface-sidebar-active"
                                    )}>
                                        <User size={18} className="shrink-0" />
                                    </SelectorTrigger>
                                    <SelectorContent>
                                        <SelectorItem value="/profile">Profile</SelectorItem>
                                        <SelectorItem value="signout">Sign out</SelectorItem>
                                    </SelectorContent>
                                </Selector>
                            </div>
                        ) : (
                            <div className="flex items-center justify-between">
                                <Selector onValueChange={(value) => {
                                    if (value === 'signout') {
                                        auth.signoutRedirect()
                                    } else {
                                        window.location.href = value
                                    }
                                }}>
                                    <SelectorTrigger hideChevron className={cn(
                                        "flex items-center gap-2 px-2 py-2 rounded-md w-auto h-auto border-none shadow-none",
                                        currentPath === '/profile'
                                            ? "bg-surface-sidebar-active border-l-3"
                                            : "hover:bg-surface-sidebar-active"
                                    )}>
                                        <User size={18} className="shrink-0" />
                                    </SelectorTrigger>
                                    <SelectorContent>
                                        <SelectorItem value="/profile">Profile</SelectorItem>
                                        <SelectorItem value="signout">Sign out</SelectorItem>
                                    </SelectorContent>
                                </Selector>
                                <button
                                    onClick={toggle}
                                    className="flex items-center justify-center p-2 rounded-md hover:bg-surface-sidebar-active text-text-sidebar hover:text-text-sidebar-active"
                                >
                                    {theme === "dark" ? <Sun size={18} /> : <Moon size={18} />}
                                </button>
                            </div>
                        )}
                    </div>
                </nav>
        </aside>
    )
}