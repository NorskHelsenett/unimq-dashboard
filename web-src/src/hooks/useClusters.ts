import { useEffect, useState } from 'react'
import { getClusters } from '@/services/clusters'
import { ClusterStats } from '@/types/clusterStats'

export function useClusters() {
    const [clusters, setClusters] = useState<ClusterStats | null>(null)
    const [loading, setLoading] = useState(true)
    
    useEffect(() => {
        getClusters()
            .then(setClusters)
            .finally(() => setLoading(false))
    }, [])
    return {clusters, loading}
}