import { cn } from "@/lib/utils"
import { ReactNode } from "react"

export const Tile = ({ children, className }: { children: ReactNode, className?: string}) => {
    return (
        <div className={cn("p-4 bg-surface-card border  border-border-card rounded-lg", className)}>
            {children}
        </div>
    )
}