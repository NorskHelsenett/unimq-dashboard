"use client"

import { RatioChart } from "../charts/RatioChart"
import { GaugeChart } from "../charts/GaugeChart"
import { Tile } from "../layout/Tile"
import { NodeStats } from "@/types/clusterStats"
import { convertBytes } from "@/lib/bytes"

export const ClusterResourceCard = ({ clusters }: { clusters: NodeStats[] | null }) => {
    if (!clusters) return <p>ClusterStats does not exist</p>

    const DiskDescription = ({ free, limit }: { free: number, limit: number }) => (
        <div>
            <p>Lower limit: {convertBytes(limit)}</p>
            <p>Free disk space: {convertBytes(free)}</p>
        </div>
    )

    return (
        <div>
            <h2 className="text-lg font-semibold text-text-primary mb-3">Cluster Resources</h2>
            {clusters.map(node => (
                <Tile key={node.name}>
                    <p className="mb-2">{node.name}</p>
                    <div className="flex gap-4">
                        <GaugeChart title="Memory" usage={node.mem_used} max={node.mem_limit} labelText={`${convertBytes(node.mem_used)}/${convertBytes(node.mem_limit)}`} fontSize={16} />
                        <RatioChart title={"Disk"} description={<DiskDescription free={node.disk_free} limit={node.disk_free_limit}/>} free={node.disk_free} limit={node.disk_free_limit} />
                    </div>
                </Tile>
            ))}
        </div>

    )
}