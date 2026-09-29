import { useEffect, useState } from "react";
import { getVhosts } from "@/services/vhosts";

export function useVhost() {
    const [vhosts, setVhosts] = useState<string[]>([])
    const [selected, setSelected] = useState('')
    const [loading, setLoading] = useState(true)

    useEffect(() => {
        getVhosts().then(names => {
            setVhosts(names)
            const fromUrl = new URLSearchParams(window.location.search).get('vhost')
            setSelected(fromUrl ?? names[0] ?? '')
            setLoading(false)
        })
    }, [])

    return { vhosts, selected, setSelected, loading }
}