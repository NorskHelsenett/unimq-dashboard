import type { IndexData, Metrics, Limits } from "@/pages/index"
import { getVhosts } from "@/services/notifications"
import { apiFetch } from "@/lib/apiClient"
import { useEffect, useState } from "react"

export function useIndex(): IndexData {
    const [vhosts, setVhosts] = useState<string[]>([])
    const [selected, setSelected] = useState<string>('')
    const [metrics, setMetrics] = useState<Metrics | null>(null)
    const [limits, setLimits] = useState<Limits>({ MaxConnections: 0, MaxQueues: 0 })

    useEffect(() => {
        getVhosts().then(names => {
            setVhosts(names)
            const sel = new URLSearchParams(window.location.search).get('vhost') ?? names[0] ?? ''
            setSelected(sel)
            // Fetch metrics and limits for the selected vhost
            apiFetch(`/v1/vhosts/${encodeURIComponent(sel)}/metrics`)
                .then(res => res.json())
                .then(data => {
                    setMetrics(data.metrics)
                    setLimits(data.limits)
                })
                .catch(() => {
                    setMetrics(null)
                    setLimits({ MaxConnections: 0, MaxQueues: 0 })
                })
        })
    }, []) 

    return { Vhosts: vhosts, Selected: selected, Metrics: metrics, Limits: limits }
}