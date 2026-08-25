import { useState, useRef, useEffect } from 'react'
import { Settings2 } from 'lucide-react'
import { Button } from '../ui/button'
import { Switch } from '../ui/switch'
import type { DashboardWidget } from '@/hooks/useDashboard'

interface DashboardCustomizerProps {
  widgets: DashboardWidget[]
  isVisible: (id: string) => boolean
  toggle: (id: string) => void
}

export function DashboardCustomizer({ widgets, isVisible, toggle }: DashboardCustomizerProps) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [open])

  return (
    <div className="relative" ref={ref}>
      <Button variant="ghost" size="sm" onClick={() => setOpen(o => !o)}>
        <Settings2 className="w-4 h-4 mr-1.5" />
        Customize
      </Button>

      {open && (
        <div className="absolute right-0 top-full mt-1 z-50 bg-white border border-border-card rounded-lg shadow-lg p-4 w-64">
          <p className="text-sm font-semibold mb-3 text-text-primary">Dashboard Widgets</p>
          <div className="space-y-2.5">
            {widgets.map(w => (
              <div key={w.id} className="flex items-center justify-between gap-4">
                <label
                  htmlFor={`widget-toggle-${w.id}`}
                  className="text-sm cursor-pointer select-none text-text-primary"
                >
                  {w.label}
                </label>
                <Switch
                  id={`widget-toggle-${w.id}`}
                  checked={isVisible(w.id)}
                  onCheckedChange={() => toggle(w.id)}
                />
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
