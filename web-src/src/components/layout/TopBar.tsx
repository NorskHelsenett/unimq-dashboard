import { useEffect, useState } from 'react'
import { getVhosts } from '@/services/vhosts'
import { VhostSelector } from './VhostSelector'
import { LiveDataWidget } from '../dashboard/LiveDataWidget'

export function TopBar() {
  const [vhosts, setVhosts] = useState<string[]>([])
  const [selected, setSelected] = useState('')

  useEffect(() => {
    getVhosts().then(names => {
      setVhosts(names)
      const fromUrl = new URLSearchParams(window.location.search).get('vhost')
      setSelected(fromUrl ?? names[0] ?? '')
    })
  }, [])

  if (vhosts.length === 0) return null

  return (
    <div className="flex items-center justify-end gap-3 px-6 pt-4 pb-3 border-b border-border-card/70">
      <VhostSelector Vhosts={vhosts} Selected={selected} />
      <LiveDataWidget />
    </div>
  )
}
