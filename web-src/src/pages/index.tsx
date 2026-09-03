import "../index.css";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { RequireAuth } from "@/auth/RequireAuth";
import { Layout } from "@/components/layout/Layout";
import { LimitsCard } from "@/components/overview/LimitsCard";
import { QueueSizeInfoCard } from "@/components/overview/QueueSizeInfoCard";
import { QueuesCard } from "@/components/overview/QueuesCard";
import { ClusterResourceCard } from "@/components/overview/ClusterResourceCard";
import { VhostResourceCard } from "@/components/overview/VhostResourceCard";
import { SizeDistributionCard } from "@/components/overview/SizeDistributionCard";
import { useIndex } from "@/hooks/useIndex";
import { useClusters } from "@/hooks/useClusters";
import { useQueues } from "@/hooks/useQueues";
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
if (!root) throw new Error("Missing #app mount point");

const MainPage = () => {
    const { Vhosts, Selected, Metrics, Limits } = useIndex()
    const { clusters, loading } = useClusters()
    const { queues, loading: queuesLoading, error: queuesError } = useQueues(Selected)

    return (
        <Layout Vhosts={Vhosts} Selected={Selected}>
            {loading ? (
                <div className="p-8 text-text-muted">Loading...</div>
            ) : (
                <div>
                    <h1 className="text-4xl mb-6">{Selected}</h1>
                    <div className="flex gap-8 items-end flex-wrap">
                        {Metrics ? (
                            <LimitsCard
                                connections={Metrics.connections}
                                channels={Metrics.channels}
                                queues={Metrics.queues}
                                unacked={Metrics.unacked}
                                maxConnections={Limits.MaxConnections}
                                maxQueues={Limits.MaxQueues}
                            />
                        ) : (
                            <p className="text-sm text-text-muted">No metrics available.</p>
                        )}

                        <QueueSizeInfoCard />
                        <QueuesCard vhost={Selected} queues={queues} loading={queuesLoading} error={queuesError} />
                        <SizeDistributionCard queues={queues} />
                        <div className="flex gap-4">
                            <ClusterResourceCard clusters={clusters}/>
                            <VhostResourceCard vhost={Selected} clusters={clusters} />
                        </div>
                    </div>
                </div>
            )}
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
