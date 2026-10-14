"use client"

import { Tile } from "../layout/Tile"
import { convertBytes } from "@/lib/bytes"
import { useVhostUsage } from "@/hooks/useVhostUsage"

export const VhostResourceCard = ({ vhost }: { vhost: string }) => {
    const { usage: vhostResource } = useVhostUsage(vhost)

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