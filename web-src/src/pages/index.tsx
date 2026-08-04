import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "../index.css";
import { RequireAuth } from "@/auth/RequireAuth";
import { getPageData } from "@/lib/pageData";
import { Layout } from "@/components/layout/Layout";
import { LimitsCard } from "@/components/overview/LimitsCard";
import { QueueSizeInfoCard } from "@/components/overview/QueueSizeInfoCard";
import { QueuesCard } from "@/components/overview/QueuesCard";
import { ClusterResourceCard } from "@/components/overview/ClusterResourceCard";
import { VhostResourceCard } from "@/components/overview/VhostResourceCard";
import { SizeDistributionCard } from "@/components/overview/SizeDistributionCard";
import { useIndex } from "@/hooks/useIndex";

export interface Metrics {
    connections: number;
    channels: number;
    queues: number;
    unacked: number;
    name: string;
}

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

//const data = getPageData<IndexData>();

const root = document.getElementById("app");
if (!root) throw new Error("Missing #app mount point");

const MainPage = ({ selected, metrics, limits }: { selected: string; metrics: Metrics | null; limits: Limits }) => {
    return (
        <div className="text-text-primary text-base">
            <h1 className="text-4xl mb-6">{selected}</h1>
            <div className="flex gap-8 items-end flex-wrap">
                {metrics ? (
                    <LimitsCard
                        connections={metrics.connections}
                        channels={metrics.channels}
                        queues={metrics.queues}
                        unacked={metrics.unacked}
                        maxConnections={limits.MaxConnections}
                        maxQueues={limits.MaxQueues}
                    />
                ) : (
                    <p className="text-sm text-text-muted">No metrics available.</p>
                )}

                {/* Respons er endret til dette formatet */}
                {/* type Response struct { */}
                {/* Code    int    `json:"code"` */}
                {/* Message string `json:"message"` */}
                {/* Body    T      `json:"body"` */}

                <QueueSizeInfoCard />
                {/* todo: queuesCard og sizedistributionCard */}
                <QueuesCard vhost={selected} /> 
                <SizeDistributionCard vhost={selected} />
                <div className="flex gap-4">
                    <ClusterResourceCard />
                    {/* Fix VhostResourceCard */}
                    <VhostResourceCard vhost={selected} />
                </div>
            </div>
        </div>
    );
};

const App = () => {
    const { Vhosts, Selected, Metrics, Limits } = useIndex();
    return (
        <Layout Vhosts={Vhosts} Selected={Selected}>
            <MainPage selected={Selected} metrics={Metrics} limits={Limits} />
        </Layout>
    );
};

createRoot(document.getElementById("app")!).render(
    <StrictMode>
        <RequireAuth>
            <App />
        </RequireAuth>
    </StrictMode>,
);
