export interface NodeStats {
  name: string;
  mem_used: number;
  mem_limit: number;
  disk_free: number;
  disk_free_limit: number;
}

export interface VhostResources {
  name: string;
  message_bytes: number;
  disk_bytes: number;
}

export interface ClusterStats {
  nodes: NodeStats[];
  total_mem_used: number;
  total_mem_limit: number;
  total_disk_free: number;
  min_disk_limit: number;
  vhost_resources: VhostResources[];
}
