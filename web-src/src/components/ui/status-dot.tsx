import { cn } from "@/lib/utils"

const colorMap = {
  danger: "bg-status-danger",
  ok: "bg-status-ok",
  warning: "bg-status-warning",
  blue: "bg-blue-500",
} as const

type DotColor = keyof typeof colorMap

export function StatusDot({
  color,
  pulse = true,
  className,
}: {
  color: DotColor
  pulse?: boolean
  className?: string
}) {
  const c = colorMap[color]
  return (
    <span className={cn("relative flex size-3 shrink-0 items-center justify-center", className)}>
      {pulse && (
        <span className={cn("absolute inline-flex size-2 animate-ping rounded-full opacity-50", c)} />
      )}
      <span className={cn("relative inline-flex size-2 rounded-full", c)} />
    </span>
  )
}
