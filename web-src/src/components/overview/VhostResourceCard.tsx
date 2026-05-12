"use client"

import { Tile } from "../layout/Tile"
import { useEffect, useState } from "react"
import { ClusterStats } from "@/types/clusterStats"
import { convertBytes } from "@/lib/bytes"

export const VhostResourceCard = ({ vhost }: { vhost: string }) => {
    const [data, setData] = useState<ClusterStats | null>(null)

    useEffect(() => {
        const load = () =>
        fetch('/api/cluster')
            .then((r) => {
            if (!r.ok) throw new Error()
            return r.json() as Promise<ClusterStats>
            })
            .then(setData)
            .catch(() => {})

        load()
        const id = setInterval(load, 15_000)
        return () => clearInterval(id)
    }, [])
    if (!data) return <p>ClusterStats does not exist</p>

    const vhostResources = data.vhost_resources
    const vhostResource = vhostResources.find(v => v.name === vhost)

    const DiskDescription = ({ free, limit }: { free: number, limit: number }) => (
        <div>
            <p>Lower limit: {convertBytes(limit)}</p>
            <p>Free disk space: {convertBytes(free)}</p>
        </div>
    )

    return (
        <div>
            <h2 className="text-lg font-semibold text-text-primary mb-3">Vhost Resources</h2>
            <Tile>
                <p className="mb-2">Data from vhost</p>
                <div className="border rounded-md">
                    <table>
                        <tr className="border-b">
                            <td className="py-2 pl-2 pr-8">Minne (meldinger)</td>
                            <td className="py-2 pl-8 pr-2">{convertBytes(vhostResource?.message_bytes || 0)}</td>
                        </tr>
                        <tr>
                            <td className="py-2 pl-2 pr-8">Disk (persistent)</td>
                            <td className="py-2 pl-8 pr-2">{convertBytes(vhostResource?.disk_bytes || 0)}</td>
                        </tr>
                    </table>
                </div>
            </Tile>
        </div>

    )
}