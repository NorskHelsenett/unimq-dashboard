import {apiFetch} from '../lib/apiClient'
import type { QueueDetail } from '@/types/queues'
import { ApiResponse } from './vhosts'


export async function getQueues(vhost: string): Promise<QueueDetail[]> {
    const res = await apiFetch(`/api/v1/vhosts/${encodeURIComponent(vhost)}/queues`)
    if (!res.ok) throw new Error('Failed to fetch queues data')
    const data: ApiResponse<QueueDetail[]> = await res.json()
    return data.body ?? []
}