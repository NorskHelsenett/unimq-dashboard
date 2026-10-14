import { apiFetch } from '../lib/apiClient'
import { NodeStats } from "@/types/clusterStats"
import { ApiResponse } from "./vhosts"

export async function getClusters(): Promise<NodeStats[]> {
    const res = await apiFetch('/api/v1/rabbitmq')
    if (!res.ok) throw new Error('Failed to fetch rabbitmq cluster node data')
    const data: ApiResponse<NodeStats[]> = await res.json()
    return data.body ?? []
}
