import * as React from "react"
import { cn } from "@/lib/utils"

const accentClasses = {
  blue:   "border-t-blue-400",
  amber:  "border-t-amber-500",
  green:  "border-t-green-400",
  danger: "border-t-red-500",
  none:   "border-t-transparent",
} as const

type Accent = keyof typeof accentClasses

function SectionCard({
  accent = "none",
  className,
  ...props
}: React.ComponentProps<"div"> & { accent?: Accent }) {
  return (
    <div
      className={cn(
        "bg-surface-card rounded-xl border border-border-card shadow-sm border-t-2 p-5",
        accentClasses[accent],
        className
      )}
      {...props}
    />
  )
}

function SectionCardHeader({
  title,
  icon,
  action,
  className,
}: {
  title: React.ReactNode
  icon?: React.ReactNode
  action?: React.ReactNode
  className?: string
}) {
  return (
    <div className={cn("flex justify-between items-center pb-3 mb-4 border-b border-border-card", className)}>
      <h3 className="text-base font-semibold text-text-primary flex items-center gap-2">
        {icon}
        {title}
      </h3>
      {action}
    </div>
  )
}

export { SectionCard, SectionCardHeader }
