import { getACLs } from "@/services/acl";
import { ACL } from "@/types/acl";
import { useEffect, useState } from "react";

interface useACLResults {
    acls: ACL[];
    loading: boolean;
}

export function useACLs(): useACLResults {
    const [acls, setAcls] = useState<ACL[]>([]);
    const [loading, setLoading] = useState<boolean>(true);

    useEffect(() => {
        async function fetchAcls() {
            setLoading(true);
            try {
                setAcls(await getACLs());
            } catch (error) {
                console.error("Failed to fetch ACLs:", error);
                setAcls([]);
            } finally {
                setLoading(false);
            }
        }
        fetchAcls();
    }, [])
    
    return { acls, loading };
}