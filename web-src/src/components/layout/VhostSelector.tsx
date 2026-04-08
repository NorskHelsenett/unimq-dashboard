import { Vhosts } from "@/types/vhosts"
import { ChangeEvent } from "react"

export function VhostSelector({ Vhosts, Selected }: Vhosts) {
    function handleChange(e: ChangeEvent<HTMLSelectElement>) {
        const params = new URLSearchParams(window.location.search)
        params.set('vhost', e.target.value)
        window.location.search = params.toString()
    }

    return (
        <select
            value={Selected}
            onChange={handleChange}
            className="text-sm text-text-secondary border border-border-card rounded-md px-2.5 py-1.5 bg-surface-card focus:outline-none focus:ring-2 focus:ring-brand focus:border-transparent cursor-pointer w-full"
        >
            {Vhosts.map((v) => (
                <option key={v} value={v}>
                    {v}
                </option>
            ))}
        </select>
    )
}