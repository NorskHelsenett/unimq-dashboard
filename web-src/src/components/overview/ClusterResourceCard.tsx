"use client"

import { RatioChart } from "../charts/RatioChart"
import { GaugeChart } from "../charts/GaugeChart"
import { Tile } from "../layout/Tile"
import { ClusterStats } from "@/types/clusterStats"
import { convertBytes } from "@/lib/bytes"

export const ClusterResourceCard = ({ clusters }: { clusters: ClusterStats | null }) => {
    if (!clusters) return <p>ClusterStats does not exist</p>

    const totalMemUsed = clusters.total_mem_used
    const totalMemLimit = clusters.total_mem_limit 
    const totalDiskFree = clusters.total_disk_free
    const minDiskLimit = clusters.min_disk_limit

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
                <div className="flex gap-4">
                    <GaugeChart title="Memory" usage={totalMemUsed} max={totalMemLimit} labelText={`${convertBytes(totalMemUsed)}/${convertBytes(totalMemLimit)}`} fontSize={16} />
                    <RatioChart title={"Disk"} description={<DiskDescription free={totalDiskFree} limit={minDiskLimit}/>} free={totalDiskFree} limit={minDiskLimit} />
                </div>
            </Tile>
        </div>

    )
}