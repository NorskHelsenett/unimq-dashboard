import * as React from 'react'
import { Slot } from '@radix-ui/react-slot'
import { cva, type VariantProps } from 'class-variance-authority'

import { cn } from '@/lib/utils'

const pillVariants = cva(
  'inline-flex items-center justify-center rounded-md border px-2 py-0.5 text-xs font-medium w-fit whitespace-nowrap shrink-0 [&>svg]:size-3 gap-1 [&>svg]:pointer-events-none focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive transition-[color,box-shadow] overflow-hidden',
  {
    variants: {
      variant: {
        default: 'border-transparent bg-primary text-primary-foreground [a&]:hover:bg-primary/90',
        secondary: 'border-transparent bg-secondary text-secondary-foreground [a&]:hover:bg-secondary/90',
        destructive:
          'border-transparent bg-destructive text-white [a&]:hover:bg-destructive/90 focus-visible:ring-destructive/20 dark:focus-visible:ring-destructive/40 dark:bg-destructive/60',
        outline: 'text-foreground [a&]:hover:bg-accent [a&]:hover:text-accent-foreground',
        red: 'border-transparent bg-red-500 dark:bg-red-600 text-primary-background',
        orange: 'border-transparent bg-orange-500 dark:bg-orange-600 text-primary-background',
        amber: 'border-transparent bg-amber-500 dark:bg-amber-600 text-primary-background',
        yellow: 'border-transparent bg-yellow-500 dark:bg-yellow-600 text-primary-background',
        lime: 'border-transparent bg-lime-500 dark:bg-lime-600 text-primary-background',
        green: 'border-transparent bg-green-500 dark:bg-green-600 text-primary-background',
        lightGreen: 'border-transparent bg-green-200 dark:bg-green-400 text-primary-background',
        emerald: 'border-transparent bg-emerald-500 dark:bg-emerald-600 text-primary-background',
        teal: 'border-transparent bg-teal-500 dark:bg-teal-600 text-primary-background',
        cyan: 'border-transparent bg-cyan-500 dark:bg-cyan-600 text-primary-background',
        sky: 'border-transparent bg-sky-500 dark:bg-sky-600 text-primary-background',
        blue: 'border-transparent bg-blue-500 dark:bg-blue-600 text-primary-background',
        lightBlue: 'border-transparent bg-blue-300 dark:bg-blue-400 text-primary-background',
        indigo: 'border-transparent bg-indigo-500 dark:bg-indigo-600 text-primary-background',
        violet: 'border-transparent bg-violet-500 dark:bg-violet-600 text-primary-background',
        purple: 'border-transparent bg-purple-500 dark:bg-purple-600 text-primary-background',
        fuchsia: 'border-transparent bg-fuchsia-500 dark:bg-fuchsia-600 text-primary-background',
        pink: 'border-transparent bg-pink-500 dark:bg-pink-600 text-primary-background',
        rose: 'border-transparent bg-rose-500 dark:bg-rose-600 text-primary-background',
        gray: 'border-transparent bg-gray-500 dark:bg-gray-600 text-primary-background',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  }
)

type PillProps = React.HTMLAttributes<HTMLSpanElement> &
  VariantProps<typeof pillVariants> & {
    asChild?: boolean
  }

const Pill = React.forwardRef<HTMLSpanElement, PillProps>(({ className, variant, asChild = false, ...props }, ref) => {
  const Comp = asChild ? Slot : 'span'
  return (
    <Comp ref={ref} data-slot='pill' className={cn(pillVariants({ variant }), className, 'rounded-full')} {...props} />
  )
})

Pill.displayName = 'Pill'

export { Pill, pillVariants }
