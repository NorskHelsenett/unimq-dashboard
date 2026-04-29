import { NAV_ITEMS } from "@/lib/navItems"
import { LogoLink } from "./LogoLink"
import { cn } from "@/lib/utils"
import { VhostSelector } from "./VhostSelector"
import { Vhosts } from "@/types/vhosts"

function isActive(itemHref: string, currentPath: string): boolean {
    if (itemHref === '/') return currentPath === '/'
    return currentPath.startsWith(itemHref)
}

export function Sidebar({ Vhosts, Selected }: Vhosts) {
    const currentPath = window.location.pathname

    return (
        <nav className="flex flex-col pt-4 border-r border-border-sidebar h-full min-h-screen">
            <LogoLink />
            <div className="p-4 flex gap-2 items-center border-y my-4">
                <span>Vhost:</span> 
                <VhostSelector Vhosts={Vhosts} Selected={Selected} />
            </div>
            <div className="flex flex-col gap-1">
                {NAV_ITEMS.map((item) => {
                    const active = isActive(item.href, currentPath)
                    return (
                        <a href={item.href} className={cn(
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
        </nav>
    )
}