import { apiFetch } from "@/lib/apiClient"
import type { Metrics } from "@/types/metrics"
import { ApiResponse } from "@/services/vhosts"

export async function getMetrics(vhost: string): Promise<Metrics> {
    const res = await apiFetch(`/api/v1/vhosts/${encodeURIComponent(vhost)}/metrics`)
    if (!res.ok) throw new Error('Failed to fetch metrics data')
    const data: ApiResponse<Metrics> = await res.json()
    return data.body
}