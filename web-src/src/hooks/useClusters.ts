import { useEffect, useState } from 'react'
import { getClusters } from '@/services/clusters'
import { NodeStats } from '@/types/clusterStats'

export function useClusters() {
    const [clusters, setClusters] = useState<NodeStats[] | null>(null)
    const [loading, setLoading] = useState(true)
    
    useEffect(() => {
        getClusters()
            .then(setClusters)
            .finally(() => setLoading(false))
    }, [])
    return {clusters, loading}
}