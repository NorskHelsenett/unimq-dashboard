export interface QueueDetail {
    name: string
    messages: number
    message_bytes: number
    history: number[]
    consumers: number
    publish_rate: number
    deliver_rate: number
    redeliver_rate: number
    messages_unacknowledged: number
}

export interface QueuesCardProps {
    vhost: string
    queues: QueueDetail[]
    loading: boolean
    error: Error | null
}


export interface SizeDistributionCardProps {
    queues: QueueDetail[]
}

