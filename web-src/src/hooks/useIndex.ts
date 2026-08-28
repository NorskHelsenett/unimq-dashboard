import type { IndexData, Limits } from "@/pages/index"
import { getVhosts } from "@/services/vhosts"
import { getMetrics } from "@/services/metrics"
import type { Metrics } from "@/types/metrics"
import { useEffect, useState } from "react"

export function useIndex(): IndexData {
    const [vhosts, setVhosts] = useState<string[]>([])
    const [selected, setSelected] = useState<string>('')
    const [metrics, setMetrics] = useState<Metrics | null>(null)
    const [limits] = useState<Limits>({ MaxConnections: 10, MaxQueues: 20 })

    useEffect(() => {
        getVhosts().then(names => {
            setVhosts(names)
            const sel = new URLSearchParams(window.location.search).get('vhost') ?? names[0] ?? ''
            setSelected(sel)
        })
    }, [])

    useEffect(() => {
        if (!selected) return
        getMetrics(selected)
            .then(setMetrics)
            .catch(() => setMetrics(null))
    }, [selected])

    return { Vhosts: vhosts, Selected: selected, Metrics: metrics, Limits: limits }
}