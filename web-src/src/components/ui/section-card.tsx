import * as React from "react"
import { cn } from "@/lib/utils"

const accentClasses = {
  blue: "border-l-blue-400",
  amber: "border-l-amber-500",
  green: "border-l-green-200",
  danger: "border-l-red-500",
  none: "",
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
        "bg-white rounded-lg shadow p-4 min-w-[300px] border border-border-card border-l-4",
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
    <div className={cn("flex justify-between items-center mb-4", className)}>
      <h3 className="text-lg font-semibold flex items-center gap-2">
        {icon}
        {title}
      </h3>
      {action}
    </div>
  )
}

export { SectionCard, SectionCardHeader }
