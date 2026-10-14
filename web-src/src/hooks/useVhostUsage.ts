import { useEffect, useState } from 'react'
import { getVhostUsage } from '@/services/vhosts'
import { VhostResources } from '@/types/clusterStats'

export function useVhostUsage(vhost: string) {
    const [usage, setUsage] = useState<VhostResources | null>(null)

    useEffect(() => {
        if (!vhost) return
        const controller = new AbortController()
        getVhostUsage(vhost, controller.signal)
            .then(setUsage)
            .catch(() => {
                if (!controller.signal.aborted) setUsage(null)
            })
        return () => controller.abort()
    }, [vhost])

    return { usage }
}
