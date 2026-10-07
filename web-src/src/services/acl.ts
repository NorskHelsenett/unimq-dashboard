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

export async function upsertACL({group, vhost_ids, permissions}: {group: string, vhost_ids: string[], permissions: string[]}): Promise<void> {
    const res = await apiFetch(`/api/v1/acls`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ group, vhost_ids, permissions })
    })
    if (!res.ok) {
        throw new Error(`Failed to save ACL: ${res.status} ${res.statusText}`)
    }
}

export async function deleteAcl(id: string): Promise<void> {
    const res = await apiFetch(`/api/v1/acls/${id}`, {
        method: 'DELETE'
    })
    if (!res.ok) {
        throw new Error(`Failed to delete ACL: ${res.status} ${res.statusText}`)
    }
}


export async function accessToACLs(): Promise<boolean> {
    const res = await apiFetch(`/api/v1/acls`)
    if (res.status === 200) {
        return true
    }
    if (res.status === 404) {
        return false
    }
    throw new Error(`Failed to check ACL access: ${res.status} ${res.statusText}`)
}