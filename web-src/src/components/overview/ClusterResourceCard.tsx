"use client"

import { RatioChart } from "../charts/RatioChart"
import { GaugeChart } from "../charts/GaugeChart"
import { Tile } from "../layout/Tile"
import { useEffect, useState } from "react"
import { ClusterStats } from "@/types/clusterStats"
import { convertBytes } from "@/lib/bytes"

export const ClusterResourceCard = () => {
    const [data, setData] = useState<ClusterStats | null>(null)

    useEffect(() => {
        const load = () =>
        fetch('/v1/cluster')
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

    const totalMemUsed = data.total_mem_used || 0
    const totalMemLimit = data.total_mem_limit || 0
    const totalDiskFree = data.total_disk_free || 0
    const minDiskLimit = data.min_disk_limit || 0

    const DiskDescription = ({ free, limit }: { free: number, limit: number }) => (
        <div>
            <p>Lower limit: {convertBytes(limit)}</p>
            <p>Free disk space: {convertBytes(free)}</p>
        </div>
    )

    return (
        <div>
            <h2 className="text-lg font-semibold text-text-primary mb-3">Cluster Resources</h2>
            <Tile>
                <p className="mb-2">General data from cluster</p>
                <div className="flex gap-4">
                    <GaugeChart title="Memory" usage={totalMemUsed} max={totalMemLimit} labelText={`${convertBytes(totalMemUsed)}/${convertBytes(totalMemLimit)}`} fontSize={16} />
                    <RatioChart title={"Disk"} description={<DiskDescription free={totalDiskFree} limit={minDiskLimit}/>} free={totalDiskFree} limit={minDiskLimit} />
                </div>
            </Tile>
        </div>

    )
}