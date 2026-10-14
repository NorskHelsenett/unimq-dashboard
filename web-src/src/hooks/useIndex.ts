import type { IndexData } from "@/pages/index"
import { getSelectedVhost, getVhosts, getVhostLimits } from "@/services/vhosts"
import { getMetrics } from "@/services/metrics"
import type { Metrics } from "@/types/metrics"
import { useEffect, useState } from "react"
import { Limits } from "@/types/vhosts"

export function useIndex(): IndexData {
    const [vhosts, setVhosts] = useState<string[]>([])
    const [selected, setSelected] = useState<string>('')
    const [metrics, setMetrics] = useState<Metrics | null>(null)

    const [limits, setLimits] = useState<Limits | null>(null)

    useEffect(() => {
        if (!selected) return
        getVhostLimits(selected)
            .then(setLimits)
            .catch(() => setLimits(null))
    }, [selected])

    useEffect(() => {
        getVhosts().then(names => {
            setVhosts(names)
            setSelected(getSelectedVhost(names))
        }).catch(() => {
            setVhosts([])
            setSelected('')
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