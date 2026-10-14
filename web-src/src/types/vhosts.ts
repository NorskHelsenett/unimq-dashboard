export interface Vhosts {
  Vhosts: string[];
  Selected: string;
}

export interface VhostObj {
  name: string
}

export interface Limits {
  vhost: string
  max_connections: number;
  max_queues: number;
}
