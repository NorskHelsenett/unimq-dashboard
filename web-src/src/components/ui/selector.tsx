import * as React from "react"
import { Select } from "radix-ui"
import { Check, ChevronDown, ChevronUp } from "lucide-react"
import { cn } from "@/lib/utils"

function Selector({ ...props }: React.ComponentProps<typeof Select.Root>) {
  return <Select.Root {...props}/>
}

function SelectorTrigger({ className, children, hideChevron, ...props }: React.ComponentProps<typeof Select.Trigger> & { hideChevron?: boolean }) {
  return (
    <Select.Trigger
      data-slot="select-trigger"
      className={cn(
        "flex h-9 w-full data-[placeholder]:text-muted-foreground  items-center justify-between gap-2 rounded-md border border-input bg-transparent px-2.5 py-1 text-sm shadow-xs transition-[color,box-shadow] outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 [&>span]:truncate",
        className
      )}
      {...props}
    >
      {children}
      {!hideChevron && (
        <Select.Icon asChild>
          <ChevronDown className="size-4 shrink-0 text-muted-foreground" />
        </Select.Icon>
      )}
    </Select.Trigger>
  )
}


function SelectLabel({ className, ...props }: React.ComponentProps<typeof Select.Label>) {
  return (
    <Select.Group>
        <Select.Label
        data-slot='select-label'
        className={cn('text-muted-foreground px-2 py-1.5 text-xs', className)}
        {...props}
        >
        </Select.Label>
    </Select.Group>
  )
}

function SelectorContent({ className, children, ...props }: React.ComponentProps<typeof Select.Content>) {
  return (
    <Select.Portal>
      <Select.Content
        data-slot="select-content"
        position="popper"
        sideOffset={4}
        className={cn(
          "relative z-50 max-h-96 min-w-[var(--radix-select-trigger-width)] overflow-hidden rounded-md border border-border-card bg-surface-card text-text-primary shadow-lg",
          "data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95",
          "data-[side=bottom]:slide-in-from-top-2 data-[side=top]:slide-in-from-bottom-2",
          className
        )}
        {...props}
      >
        <Select.ScrollUpButton className="flex cursor-default items-center justify-center py-1 text-text-muted">
          <ChevronUp className="size-4" />
        </Select.ScrollUpButton>
        <Select.Viewport className="p-1">
          {children}
        </Select.Viewport>
        <Select.ScrollDownButton className="flex cursor-default items-center justify-center py-1 text-text-muted">
          <ChevronDown className="size-4" />
        </Select.ScrollDownButton>
      </Select.Content>
    </Select.Portal>
  )
}

function SelectorItem({ className, children, ...props }: React.ComponentProps<typeof Select.Item>) {
  return (
    <Select.Item
      data-slot="select-item"
      className={cn(
        "relative flex w-full cursor-default select-none items-center rounded-sm py-1.5 pl-8 pr-2 text-sm text-text-secondary outline-none focus:bg-surface-page focus:text-text-primary data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
        className
      )}
      {...props}
    >
      <span className="absolute left-2 flex size-3.5 items-center justify-center">
        <Select.ItemIndicator>
          <Check className="size-3.5" />
        </Select.ItemIndicator>
      </span>
      <Select.ItemText>{children}</Select.ItemText>
    </Select.Item>
  )
}

const SelectorValue = Select.Value

export { Selector, SelectLabel, SelectorTrigger, SelectorContent, SelectorItem, SelectorValue }