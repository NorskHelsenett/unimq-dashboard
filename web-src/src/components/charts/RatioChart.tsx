
import { Bar, BarChart, CartesianGrid, Label, Legend, Pie, PieChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { ChartContainer, type ChartConfig } from "@/components/ui/chart"
import { fmtBytes } from "@/lib/format"
import { convertBytes } from "../overview/ClusterResourceCard"
import { ReactNode } from "react"

export const description = "A radial chart with stacked sections"

interface RatioChartProps {
    title: string
    description?: ReactNode
    free: number
    limit: number
}

export const RatioChart = ({ title, description, free, limit }: RatioChartProps) => {
    const ratio = limit / free

    const chartColor = ratio < 0.33 ? "var(--status-ok)" : ratio < 0.66 ? "var(--status-warning)" : "var(--status-danger)"
    const chartBgColor = ratio < 0.33 ? "var(--status-ok-border)" : ratio < 0.66 ? "var(--status-warning-border)" : "var(--status-danger-border)"

    const data = [
        { name: "Usage ratio", part1: ratio, part2: 1 - ratio },
    ];
    
    return (
        <Card className="flex flex-col w-66">
            <CardHeader className="items-center pb-0">
                <CardTitle>{title}</CardTitle>
            </CardHeader>
            <CardContent className="items-center pb-0">
                {description}
                <ChartContainer
                    config={{}}
                    className="mx-auto aspect-square w-full max-w-62 h-12 mt-7"
                >
                    <BarChart
                        accessibilityLayer
                        layout="vertical"
                        data={data}
                    >
                        <XAxis type="number" hide />
                        <YAxis type="category" hide />
                        <Bar dataKey="part1" stackId="a" fill={chartColor} isAnimationActive={true} radius={5} barSize={36} />
                        <Bar dataKey="part2" stackId="a" fill={chartBgColor} isAnimationActive={true} radius={5} barSize={36} />
                    </BarChart>
                </ChartContainer>
                <p className="text-xs">Displays ratio between limit / free space</p>
            </CardContent>
        </Card>
    )
}

