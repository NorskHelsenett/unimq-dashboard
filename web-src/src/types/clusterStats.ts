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
