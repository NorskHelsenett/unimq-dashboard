/**
 * The session now lives in an HttpOnly cookie that the backend sets at the end
 * of the Dex login. Page scripts cannot read that cookie, so there is no token
 * for the frontend to store, refresh or attach to requests — the browser sends
 * it automatically. This module only has to notice when it has gone away.
 */

/** Backend endpoint that starts the Dex authorization code flow. */
const LOGIN_URL = "/api/v1/login/redirect";

/** Backend endpoint that clears the session cookie. */
const LOGOUT_URL = "/api/v1/logout";

/** Backend endpoint that trades the refresh token for a fresh session. */
const REFRESH_URL = "/api/v1/login/refresh";

/**
 * Thrown when the backend reports 401. By the time this surfaces, a redirect
 * to the login page has already been scheduled, so callers normally just let
 * it propagate rather than rendering an error.
 */
export class UnauthorizedError extends Error {
    constructor() {
        super("session is missing or expired");
        this.name = "UnauthorizedError";
    }
}

/**
 * Several requests can fail with 401 at the same moment (a dashboard fires off
 * a handful on mount). Without this latch each one would trigger its own
 * navigation and the last writer would win, discarding the others mid-flight.
 */
let loginStarted = false;

/** Sends the browser to Dex via the backend. Safe to call more than once. */
export function startLogin(): void {
    if (loginStarted) {
        return;
    }
    loginStarted = true;
    window.location.assign(LOGIN_URL);
}

/**
 * Clears the server-side session and returns to the root. Dex exposes no
 * end_session_endpoint, so the identity provider session itself is left alone;
 * signing back in will not prompt for credentials again until Dex expires it.
 */
export async function logout(): Promise<void> {
    try {
        await fetch(LOGOUT_URL, { method: "POST", credentials: "include" });
    } finally {
        // Navigate even if the request failed, otherwise a network blip would
        // leave the user looking at a page they believe they have left.
        window.location.assign("/");
    }
}

/**
 * Asks the backend to extend the session using the refresh token, and reports
 * whether it worked.
 *
 * This deliberately uses raw fetch rather than apiFetch: a 401 here means the
 * refresh token is spent or revoked, and the caller — not this module — should
 * decide whether that warrants throwing the user back to the login page.
 */
export async function refreshSession(): Promise<boolean> {
    try {
        const res = await fetch(REFRESH_URL, {
            method: "POST",
            credentials: "include",
        });
        return res.ok;
    } catch {
        // A network blip should not be reported as a dead session; the caller
        // will try again on the next tick.
        return false;
    }
}

/**
 * Wrapper around fetch for every call to our API.
 *
 * - `credentials: "include"` guarantees the session cookie is sent.
 * - A 401 means the session is gone, which is never something an individual
 *   screen can recover from, so it is handled centrally here.
 *
 * Any other status, 403 included, is handed back untouched so existing callers
 * that inspect `res.ok` keep behaving exactly as they did before.
 */
export async function apiFetch(
    input: RequestInfo | URL,
    init?: RequestInit,
): Promise<Response> {
    const res = await fetch(input, { ...init, credentials: "include" });

    if (res.status === 401) {
        startLogin();
        throw new UnauthorizedError();
    }

    return res;
}
