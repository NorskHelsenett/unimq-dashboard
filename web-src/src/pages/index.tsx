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
    const firsName = (auth.user?.profile?.name ?? "User").split(" ")[0]
    const greeting = hour < 12 ? "Good morning" : hour < 18 ? "Good afternoon" : "Good evening"

    return (
        <span className="text-text-primary font-normal">{greeting}, {firsName} 👋</span>
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
        <Layout>
            <div className="space-y-6">
                <div className="flex items-start justify-between mb-12">
                    <div>
                        <h1 className="text-3xl font-bold tracking-tight">
                            <GreetingHeader />
                        </h1>
                        <p className="flex items-center gap-1.5 text-sm text-text-muted mt-1">
                            Here is what's happening on your RabbitMQ instance.
                        </p>
                        {/* <LiveDataWidget vhost={Selected} /> */}
                    </div>
                    <div className="flex items-center gap-4">
                        <DashboardCustomizer widgets={widgets} isVisible={isVisible} toggle={toggle} />
                    </div>
                </div>

                {/* Notification cards row */}
                {hasRightCards && (
                    <div className="grid grid-cols-1 lg:grid-cols-3 gap-5">
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

                {/* Cluster + limits row */}
                {(isVisible('cluster') || isVisible('limits')) && (
                    <div className="grid grid-cols-1 lg:grid-cols-6 gap-5">
                        {isVisible('cluster') && (
                            <div className={isVisible('limits') ? 'lg:col-span-3' : 'lg:col-span-6'}>
                                <DashboardClusterWidget clusters={clusters} vhost={Selected} />
                            </div>
                        )}
                        {isVisible('limits') && Metrics && (
                            <div className={isVisible('cluster') ? 'lg:col-span-3' : 'lg:col-span-6'}>
                                <LimitsCard
                                    connections={Metrics.connections}
                                    channels={Metrics.channels}
                                    queues={Metrics.queues}
                                    unacked={Metrics.unacked}
                                    maxConnections={Limits.MaxConnections}
                                    maxQueues={Limits.MaxQueues}
                                />
                            </div>
                        )}
                    </div>
                )}


                {/* Size distribution row */}
                {isVisible('sizeDistribution') && (
                    <SizeDistributionCard queues={queues} />
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
