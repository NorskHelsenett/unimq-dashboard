import { useEffect, useState } from "react"
import { getQueues } from "@/services/queues"
import { QueueDetail } from "@/types/queues"

export function useQueues(vhost: string) {
    const [queues, setQueues] = useState<QueueDetail[]>([])
    const [loading, setLoading] = useState<boolean>(true)
    const [error, setError] = useState<Error | null>(null)

    useEffect(() => {
        if (!vhost) return
        getQueues(vhost)
            .then(data => {
                setQueues(data)
                setLoading(false)
            })
            .catch(err => {
                setError(err)
                setLoading(false)
            })
    }, [vhost])

    return { queues, loading, error }
}
