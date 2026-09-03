import { apiFetch } from '../lib/apiClient'
import { ClusterStats } from "@/types/clusterStats"
import { ApiResponse } from "./vhosts"

export async function getClusters(): Promise<ClusterStats> {
    const res = await apiFetch('/api/v1/cluster')
    if (!res.ok) throw new Error('Failed to fetch cluster data')
    const data: ApiResponse<ClusterStats> = await res.json()
    return data.body
}
