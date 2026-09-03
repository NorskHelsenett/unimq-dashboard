
import { Pie, PieChart } from "recharts"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { ChartContainer, type ChartConfig } from "@/components/ui/chart"

export const description = "A radial chart with stacked sections"

interface GaugeChartProps {
    title: string
    usage: number
    max: number
    labelText: string
    fontSize: number
}

export const GaugeChart = ({ title, usage, max, labelText, fontSize }: GaugeChartProps) => {
    const nonUsage = max-usage
    const widthNum = 210
    const usagePercentage = usage/max*100

    const chartColor = usagePercentage < 33 ? "var(--status-ok)" : usagePercentage < 66 ? "var(--status-warning)" : "var(--status-danger)"
    const chartBgColor = usagePercentage < 33 ? "var(--status-ok-border)" : usagePercentage < 66 ? "var(--status-warning-border)" : "var(--status-danger-border)"
    
    const chartConfig = {
        used: {
            label: "Used",
            color: chartColor,
        },
        unused: {
            label: "Unused",
            color: chartBgColor,
        },
    } satisfies ChartConfig

    const pieProps = {
        startAngle: 180,
        endAngle: 0,
        cx: widthNum / 2,
        cy: widthNum / 2,
        isAnimationActive: true
    };
    const usageData = [
        { name: "Usage", value: usage, fill: chartColor },
        { name: "NonUsage", value: nonUsage, fill: chartBgColor },
    ];
    
    return (
        <Card className="flex flex-col w-66">
            <CardHeader className="items-center pt-4 pb-0">
                <CardTitle>{title}</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col items-center pb-3 px-2">
                {/* overflow-hidden clips the empty bottom half of the square SVG */}
                <div className="overflow-hidden w-full max-w-62 max-h-36">
                    <ChartContainer
                        config={chartConfig}
                        className="mx-auto w-full aspect-square"
                    >
                        <PieChart accessibilityLayer data={usageData}>
                            <Pie
                                stroke="none"
                                data={usageData}
                                fill={chartColor}
                                innerRadius={(widthNum / 2) * 0.8}
                                outerRadius={(widthNum / 2)}
                                cornerRadius="5%"
                                {...pieProps}
                            />
                        </PieChart>
                    </ChartContainer>
                </div>
                <p className="text-sm font-bold mt-1" style={{ fontSize: `${fontSize}px` }}>{labelText}</p>
            </CardContent>
        </Card>
    )
}

