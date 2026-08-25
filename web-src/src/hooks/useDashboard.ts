import { useLocalStorage } from './useLocalStorage'

export interface DashboardWidget {
  id: string
  label: string
  defaultVisible: boolean
}

export const ALL_WIDGETS: DashboardWidget[] = [
  { id: 'limits',           label: 'Limits & Metrics',          defaultVisible: true },
  { id: 'queues',           label: 'Queues',                    defaultVisible: true },
  { id: 'alarms',           label: 'Alarms',                    defaultVisible: true },
  { id: 'recipients',       label: 'Notification Recipients',   defaultVisible: true },
  { id: 'maintenance',      label: 'Upcoming Maintenance',      defaultVisible: true },
  { id: 'cluster',          label: 'Cluster & Vhost Resources', defaultVisible: true },
  { id: 'sizeDistribution', label: 'Message Size Distribution', defaultVisible: false },
]

type WidgetVisibility = Record<string, boolean>

export function useDashboard() {
  const defaultVisibility = Object.fromEntries(
    ALL_WIDGETS.map(w => [w.id, w.defaultVisible])
  )

  const [visibility, setVisibility] = useLocalStorage<WidgetVisibility>(
    'dashboard-widgets-v1',
    defaultVisibility
  )

  const isVisible = (id: string): boolean =>
    visibility[id] ?? (ALL_WIDGETS.find(w => w.id === id)?.defaultVisible ?? true)

  const toggle = (id: string) =>
    setVisibility(prev => ({ ...prev, [id]: !isVisible(id) }))

  return { isVisible, toggle, widgets: ALL_WIDGETS }
}
