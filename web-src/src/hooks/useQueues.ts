import { useEffect, useState } from "react"
import { getQueues } from "@/services/queues"
import { QueueDetail } from "@/types/queues"

export function useQueues(vhost: string) {
    const [queues, setQueues] = useState<QueueDetail[]>([])
    const [loading, setLoading] = useState<boolean>(true)
    const [error, setError] = useState<Error | null>(null)

useEffect(() => {
    if (!vhost) {
        setQueues([])
        setError(null)
        setLoading(false)
        return
    }

    let cancelled = false
    setLoading(true)
    setError(null)

    getQueues(vhost)
        .then(data => {
            if (cancelled) return
            setQueues(data)
            setLoading(false)
        })
        .catch(err => {
            if (cancelled) return
            setError(err)
            setLoading(false)
        })

    return () => { cancelled = true }
}, [vhost])

    return { queues, loading, error }
}
