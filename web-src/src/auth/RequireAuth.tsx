import {
    createContext,
    useCallback,
    useContext,
    useEffect,
    useMemo,
    useState,
} from "react";
import { LogIn, RefreshCw, ShieldX } from "lucide-react";
import { fetchProfile, type Profile } from "@/services/profile";
import { logout, refreshSession, startLogin, UnauthorizedError } from "@/lib/apiClient";

/** How long before expiry the session is renewed in the background. */
const RENEW_LEAD_SECONDS = 60;

interface SessionValue {
    profile: Profile;
    /** Renews the session. Resolves true when the session was extended. */
    refresh: () => Promise<boolean>;
}

/**
 * Holds the signed-in user for everything below <RequireAuth>. This replaces
 * react-oidc-context's AuthProvider: the frontend no longer speaks OIDC, so the
 * only thing left to share is the profile the backend already resolved.
 */
const SessionContext = createContext<SessionValue | null>(null);

function useSessionValue(): SessionValue {
    const value = useContext(SessionContext);

    if (!value) {
        throw new Error("useSession must be called inside <RequireAuth>");
    }

    return value;
}

/**
 * Reads the signed-in user. Only valid inside <RequireAuth>, which is
 * guaranteed to have a profile by the time it renders its children — hence the
 * non-nullable return type and the loud error if that assumption is broken.
 */
export function useSession(): Profile {
    return useSessionValue().profile;
}

/** Lets a component renew the session on demand, e.g. the expiry banner. */
export function useSessionRefresh(): () => Promise<boolean> {
    return useSessionValue().refresh;
}

/** Convenience for the common "is this user an admin" question. */
export function useIsAdmin(): boolean {
    // Reaching past RequireAuth means the backend already accepted the session,
    // and today that check is admin-or-nothing. Kept as a named hook so the
    // call sites do not have to change when real ACLs land.
    useSessionValue();
    return true;
}

type SessionState =
    | { phase: "loading" }
    | { phase: "ready"; profile: Profile }
    | { phase: "forbidden" }
    | { phase: "error"; message: string };

function FullScreen({ children }: { children: React.ReactNode }) {
    return (
        <div className="flex flex-col items-center justify-center h-screen gap-3 text-center px-6">
            {children}
        </div>
    );
}

/**
 * Gate in front of every page. It asks the backend who the user is; the answer
 * doubles as the session check.
 *
 * Each page is its own Vite entry point and mounts this separately, so the
 * probe runs once per page load rather than once per app.
 */
export function RequireAuth({ children }: { children: React.ReactNode }) {
    const [state, setState] = useState<SessionState>({ phase: "loading" });
    // Bumping this re-runs the probe; it is how the retry button works.
    const [attempt, setAttempt] = useState(0);

    const retry = useCallback(() => {
        setState({ phase: "loading" });
        setAttempt((n) => n + 1);
    }, []);

    /**
     * Renews the session, then re-reads the profile so the new expiry is
     * reflected in state. Returning a boolean lets callers show a failure
     * without this having to own any UI.
     */
    const refresh = useCallback(async (): Promise<boolean> => {
        if (!(await refreshSession())) {
            return false;
        }

        const result = await fetchProfile();
        if (result.status !== "ok") {
            return false;
        }

        setState({ phase: "ready", profile: result.profile });
        return true;
    }, []);

    useEffect(() => {
        const controller = new AbortController();
        // React 18 StrictMode mounts effects twice in development. Without this
        // the first (aborted) run could still call setState on the second.
        let active = true;

        fetchProfile(controller.signal)
            .then((result) => {
                if (!active) {
                    return;
                }

                if (result.status === "ok") {
                    setState({ phase: "ready", profile: result.profile });
                } else if (result.status === "forbidden") {
                    setState({ phase: "forbidden" });
                } else {
                    setState({ phase: "error", message: result.message });
                }
            })
            .catch((err: unknown) => {
                if (!active || controller.signal.aborted) {
                    return;
                }

                // apiFetch has already started the hop to Dex. Leaving the
                // spinner up avoids flashing an error during the navigation.
                if (err instanceof UnauthorizedError) {
                    return;
                }

                setState({
                    phase: "error",
                    message: err instanceof Error ? err.message : "unknown error",
                });
            });

        return () => {
            active = false;
            controller.abort();
        };
    }, [attempt]);

    /**
     * Renew shortly before the ID token lapses, so an active user is never
     * interrupted. A successful refresh replaces the profile, which re-runs
     * this effect and schedules the next renewal. A failed one leaves the
     * profile untouched, so nothing reschedules and the expiry banner takes
     * over rather than the page hammering the endpoint.
     */
    useEffect(() => {
        if (state.phase !== "ready" || !state.profile.expires_at) {
            return;
        }

        const dueInMs =
            (state.profile.expires_at - RENEW_LEAD_SECONDS) * 1000 - Date.now();
        const timer = setTimeout(() => void refresh(), Math.max(dueInMs, 0));

        return () => clearTimeout(timer);
    }, [state, refresh]);

    const value = useMemo<SessionValue | null>(
        () => (state.phase === "ready" ? { profile: state.profile, refresh } : null),
        [state, refresh],
    );

    if (state.phase === "loading") {
        return (
            <div className="flex items-center justify-center h-screen text-muted-foreground">
                Checking your session…
            </div>
        );
    }

    if (state.phase === "forbidden") {
        return (
            <FullScreen>
                <ShieldX className="text-destructive" size={28} />
                <p className="text-destructive font-medium">
                    Your account does not have access to this dashboard.
                </p>
                <p className="text-sm text-muted-foreground max-w-md">
                    Ask an administrator to grant one of your groups at least read
                    access to a virtual host through an ACL.
                </p>
                <button className="underline text-sm mt-2" onClick={() => void logout()}>
                    Sign out
                </button>
            </FullScreen>
        );
    }

    if (state.phase === "error") {
        return (
            <FullScreen>
                <p className="text-destructive">Could not verify your session.</p>
                <p className="text-sm text-muted-foreground">{state.message}</p>
                <div className="flex items-center gap-4 mt-2">
                    <button
                        className="underline text-sm inline-flex items-center gap-1.5"
                        onClick={retry}
                    >
                        <RefreshCw size={14} /> Try again
                    </button>
                    <button
                        className="underline text-sm inline-flex items-center gap-1.5"
                        onClick={startLogin}
                    >
                        <LogIn size={14} /> Sign in
                    </button>
                </div>
            </FullScreen>
        );
    }

    return <SessionContext.Provider value={value}>{children}</SessionContext.Provider>;
}
