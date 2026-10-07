import { getACLs, accessToACLs } from "@/services/acl";
import { ACL } from "@/types/acl";
import { useEffect, useState } from "react";

interface useACLResults {
    acls: ACL[];
    loading: boolean,
    error: Error | null,
    refetch: () => void
}

export function useACLs(): useACLResults {
    const [acls, setAcls] = useState<ACL[]>([])
    const [loading, setLoading] = useState<boolean>(true)
    const [error, setError] = useState<Error | null>(null)
    const [tick, setTick] = useState(0)

    useEffect(() => {
        setLoading(true)
        setError(null)
        getACLs()
        .then(setAcls)
        .catch((err) => {
            console.error("Failed to fetch ACLs:", err)
            setAcls([])
            setError(err)
        })
        .finally(() => setLoading(false))
    }, [tick])
    
    return { acls, loading, error, refetch: () => setTick(t => t + 1) }
}

export function useACLAccess(): boolean {
    const [hasAccess, setHasAccess] = useState(false)

    useEffect(() => {
        let cancelled = false

        accessToACLs()
            .then((allowed) => {
                if (!cancelled) setHasAccess(allowed)
            })
            .catch((error) => {
                if (!cancelled) console.error("Failed to check ACL access", error)
            })

        return () => {
            cancelled = true
        }
    }, [])

    return hasAccess
}