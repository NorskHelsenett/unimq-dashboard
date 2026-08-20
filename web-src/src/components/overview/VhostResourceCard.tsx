"use client"

import { Tile } from "../layout/Tile"
import { ClusterStats } from "@/types/clusterStats"
import { convertBytes } from "@/lib/bytes"

export const VhostResourceCard = ({ vhost, clusters }: { vhost: string, clusters: ClusterStats | null }) => {
    const clusterStats = clusters
    
    if (!clusterStats) return <p>ClusterStats does not exist</p>

    const vhostResources = clusterStats.vhost_resources ?? []
    const vhostResource = vhostResources.find(v => v.name === vhost)

    return (
        <div>
            <h2 className="text-lg font-semibold text-text-primary mb-3">Vhost Resources</h2>
            <Tile>
                <p className="mb-2">Data from vhost</p>
                <div className="border rounded-md">
                    <table>
                        <tbody>
                            <tr className="border-b">
                                <td className="py-2 pl-2 pr-8">Minne (meldinger)</td>
                                <td className="py-2 pl-8 pr-2">{convertBytes(vhostResource?.message_bytes || 0)}</td>
                            </tr>
                            <tr>
                                <td className="py-2 pl-2 pr-8">Disk (persistent)</td>
                                <td className="py-2 pl-8 pr-2">{convertBytes(vhostResource?.disk_bytes || 0)}</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </Tile>
        </div>

    )
}