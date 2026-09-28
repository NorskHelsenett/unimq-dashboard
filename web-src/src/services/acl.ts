import { apiFetch } from "@/lib/apiClient";
import { ApiResponse } from "./vhosts";
import { ACL } from "@/types/acl";

export async function getACLs(): Promise<ACL[]> {
    const res = await apiFetch(`/api/v1/acls`)
    if (!res.ok) {
        throw new Error(`Failed to fetch ACLs: ${res.status} ${res.statusText}`)
    }
    const data: ApiResponse<ACL[]> = await res.json();
    return data.body
}

