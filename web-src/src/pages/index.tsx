import "../index.css";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { RequireAuth } from "@/auth/RequireAuth";
import { Layout } from "@/components/layout/Layout";
import { LimitsCard } from "@/components/overview/LimitsCard";
import { QueuesCard } from "@/components/overview/QueuesCard";
import { SizeDistributionCard } from "@/components/overview/SizeDistributionCard";
import { DashboardAlarmsSummaryWidget } from "@/components/dashboard/DashboardAlarmsSummaryWidget";
import { DashboardActiveRecipientsWidget } from "@/components/dashboard/DashboardActiveRecipientsWidget";
import { DashboardMaintenanceWidget } from "@/components/dashboard/DashboardMaintenanceWidget";
import { DashboardClusterWidget } from "@/components/dashboard/DashboardClusterWidget";
import { DashboardCustomizer } from "@/components/dashboard/DashboardCustomizer";
import { LiveDataWidget } from "@/components/dashboard/LiveDataWidget";
import { useIndex } from "@/hooks/useIndex";
import { useClusters } from "@/hooks/useClusters";
import { useQueues } from "@/hooks/useQueues";
import { useVhostNotification } from "@/hooks/useVhostNotification";
import { useScheduledMaintenance } from "@/hooks/useMaintenance";
import { useDashboard } from "@/hooks/useDashboard";
import { useAuth } from 'react-oidc-context'
import type { Metrics } from "@/types/metrics"

export interface Limits {
    MaxConnections: number;
    MaxQueues: number;
}

export interface IndexData {
    Vhosts: string[];
    Selected: string;
    Metrics: Metrics | null;
    Limits: Limits;
}

const root = document.getElementById("app");
if (!root) throw new Error("Missing #app mount point")

function GreetingHeader() {
    const hour = new Date().getHours()
    const auth = useAuth()
    const userName = auth.user?.profile?.name ?? "User"
    const greeting = hour < 12 ? "Good morning" : hour < 18 ? "Good afternoon" : "Good evening"

    return (
        <>
            <span className="text-text-muted font-normal">{greeting}, </span>
            <span className="text-text-primary font-normal">{userName}</span>
        </>
    )
}

const MainPage = () => {
    const { Vhosts, Selected, Metrics, Limits } = useIndex()
    const { clusters } = useClusters()
    const { queues, loading: queuesLoading, error: queuesError } = useQueues(Selected)
    const { notification } = useVhostNotification()
    const { maintenanceSchedule } = useScheduledMaintenance()
    const { isVisible, toggle, widgets } = useDashboard()

    const hasRightCards =
        isVisible('alarms') || isVisible('recipients') || isVisible('maintenance')

    return (
        <Layout Vhosts={Vhosts} Selected={Selected}>
            <div className="space-y-6">
                <div className="flex items-start justify-between gap-4 mb-12">
                    <div>
                        <h1 className="text-3xl font-bold tracking-tight">
                            <GreetingHeader />
                        </h1>
                        <LiveDataWidget vhost={Selected} />
                    </div>
                    <DashboardCustomizer widgets={widgets} isVisible={isVisible} toggle={toggle} />
                </div>

                {isVisible('limits') && Metrics && (
                    <LimitsCard
                        connections={Metrics.connections}
                        channels={Metrics.channels}
                        queues={Metrics.queues}
                        unacked={Metrics.unacked}
                        maxConnections={Limits.MaxConnections}
                        maxQueues={Limits.MaxQueues}
                    />
                )}

                {/* Notification cards row */}
                {hasRightCards && (
                    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
                        {isVisible('maintenance') && (
                            <DashboardMaintenanceWidget schedule={maintenanceSchedule} />
                        )}
                        {isVisible('alarms') && (
                            <DashboardAlarmsSummaryWidget notification={notification} />
                        )}
                        {isVisible('recipients') && (
                            <DashboardActiveRecipientsWidget notification={notification} />
                        )}
                    </div>
                )}

                {/* Queues table */}
                {isVisible('queues') && (
                    <QueuesCard
                        vhost={Selected}
                        queues={queues}
                        loading={queuesLoading}
                        error={queuesError}
                    />
                )}

                {/* Bottom row: cluster + size distribution */}
                {(isVisible('cluster') || isVisible('sizeDistribution')) && (
                    <div className="grid grid-cols-1 lg:grid-cols-3 gap-5">
                        {isVisible('cluster') && (
                            <div className={isVisible('sizeDistribution') ? 'lg:col-span-2' : 'lg:col-span-3'}>
                                <DashboardClusterWidget clusters={clusters} vhost={Selected} />
                            </div>
                        )}
                        {isVisible('sizeDistribution') && (
                            <div className={!isVisible('cluster') ? 'lg:col-span-3' : ''}>
                                <SizeDistributionCard queues={queues} />
                            </div>
                        )}
                    </div>
                )}
            </div>
        </Layout>
    )
}

createRoot(document.getElementById("app")!).render(
    <StrictMode>
        <RequireAuth>
            <MainPage />
        </RequireAuth>
    </StrictMode>,
);
