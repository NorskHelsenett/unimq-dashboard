export interface Metrics {
    connections: number;
    channels: number;
    queues: number;
    unacked: number;
    name: string;
}