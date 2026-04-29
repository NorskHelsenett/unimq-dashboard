"use client"

import { RatioChart } from "../charts/RatioChart"
import { GaugeChart } from "../charts/GaugeChart"
import { Tile } from "../layout/Tile"
import { useEffect, useState } from "react"



interface NodeStats {
    name: string
    mem_used: number       
    mem_limit: number      
    disk_free: number      
    disk_free_limit: number 
}

interface VhostResources {
    name: string
    message_bytes: number  
    disk_bytes: number  
}

interface ClusterStats {
    nodes: NodeStats[]      
    total_mem_used: number            
    total_mem_limit: number            
    total_disk_free: number            
    min_disk_limit: number            
    vhost_resources: VhostResources[] 
}

const BYTE_SIZE_UNITS = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB']

export function convertBytes(bytes: number) {
    let count = 0
    let unit = BYTE_SIZE_UNITS[0]
    let num = bytes
    while (num.toString().split(".")[0].length > 3) {
        num = num/1000
        count++
        unit = BYTE_SIZE_UNITS[count]
        console.log("TEst", count, unit, num)
    }
    return `${num.toFixed(2)} ${unit}`
}

export const ClusterResourceCard = () => {
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
            <Tile className="flex gap-4">
                <GaugeChart title="Memory" usage={totalMemUsed} max={totalMemLimit} labelText={`${convertBytes(totalMemUsed)}/${convertBytes(totalMemLimit)}`} fontSize={16} />
                <RatioChart title={"Disk"} description={<DiskDescription free={totalDiskFree} limit={minDiskLimit}/>} free={totalDiskFree} limit={minDiskLimit} />
            </Tile>
        </div>

    )
}