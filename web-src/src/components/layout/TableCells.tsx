import { cn } from "@/lib/utils";

export const cellPaddingStyling = "py-2.5 px-3"
export function HeaderCell({ children, className, colSpan }: { children?: React.ReactNode; className?: string; colSpan?: number }) {
    return (
        <th colSpan={colSpan} className={cn('py-2 text-xs font-bold uppercase tracking-wide text-text bg-brand/30', className)}>
            {children}
        </th>
    )
}

export function SubHeaderCell({ children, className }: { children?: React.ReactNode; className?: string }) {
    return (
        <th className={cn('py-1 text-[10px] font-semibold text-text bg-brand/15', className)}>
            {children}
        </th>
    )
}