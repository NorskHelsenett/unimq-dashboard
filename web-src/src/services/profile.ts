import { apiFetch } from "@/lib/apiClient";

/** Mirrors models.Profile in internal/models/profile.go. */
export interface Profile {
    username: string;
    email: string;
    email_verified: boolean;
    groups: string[];
    subject: string;
    issuer: string;
    /** Unix seconds. */
    issued_at: number;
    /** Unix seconds. Used to warn before the session lapses. */
    expires_at: number;
}

/** The API wraps every payload in httpsuite.Response. */
interface ApiResponse<T> {
    code: number;
    message: string;
    body: T;
}

/**
 * The three outcomes the session probe has to tell apart. A plain
 * `Profile | null` would collapse "not an admin" into "something broke", and
 * those need very different screens: one is final, the other is retryable.
 */
export type ProfileResult =
    | { status: "ok"; profile: Profile }
    | { status: "forbidden" }
    | { status: "error"; message: string };

/**
 * Fetches the signed-in user. Doubles as the session probe: reaching this
 * endpoint at all proves the session cookie is valid, because the route sits
 * behind the Authorization middleware.
 *
 * A 401 never returns here — apiFetch turns it into an UnauthorizedError and
 * starts the redirect to Dex.
 */
export async function fetchProfile(signal?: AbortSignal): Promise<ProfileResult> {
    const res = await apiFetch("/api/v1/profile", { signal });

    // Authenticated, but not in ADMIN_GROUPS. Signing in again cannot fix it.
    if (res.status === 403) {
        return { status: "forbidden" };
    }

    if (!res.ok) {
        return { status: "error", message: `profile request failed (${res.status})` };
    }

    const data = (await res.json()) as ApiResponse<Profile | null>;
    const profile = data.body;

    if (!profile?.email) {
        return { status: "error", message: "profile response was empty" };
    }

    return { status: "ok", profile };
}
