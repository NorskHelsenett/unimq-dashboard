import { useEffect, useState } from "react";
import { getSelectedVhost, getVhosts } from "@/services/vhosts";

export function useVhost() {
    const [vhosts, setVhosts] = useState<string[]>([])
    const [selected, setSelected] = useState('')
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        const controller = new AbortController()
        getVhosts(controller.signal)
            .then(names => {
                if (controller.signal.aborted) return
                setVhosts(names)
                setSelected(getSelectedVhost(names))
            })
            .catch((reason: unknown) => {
                if (controller.signal.aborted) return
                setError(reason instanceof Error ? reason.message : 'Could not load virtual hosts')
            })
            .finally(() => {
                if (!controller.signal.aborted) setLoading(false)
            })
        return () => controller.abort()
    }, [])

    return { vhosts, selected, setSelected, loading, error }
}