import { getACLs } from "@/services/acl";
import { ACL } from "@/types/acl";
import { useEffect, useState } from "react";

interface useACLResults {
    acls: ACL[];
    loading: boolean
    refetch: () => void
}

export function useACLs(): useACLResults {
    const [acls, setAcls] = useState<ACL[]>([])
    const [loading, setLoading] = useState<boolean>(true)
    const [tick, setTick] = useState(0)

    useEffect(() => {
        setLoading(true)
        getACLs()
        .then(setAcls)
        .catch((err) => {
            console.error("Failed to fetch ACLs:", err)
            setAcls([])
        })
        .finally(() => setLoading(false))
    }, [tick])
    
    return { acls, loading, refetch: () => setTick(t => t + 1) }
}