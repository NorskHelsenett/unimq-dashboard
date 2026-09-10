declare global {
    interface Window {
        __ENV__?: Record<string, string>;
    }
}

/** Reads a runtime-injected value from /env.js (see public/env.js.tmpl +
 *  dockerfiles/entrypoint-frontend.sh), falling back to Vite's build-time
 *  env for local dev (`npm run dev`, which never has env.js). */
export function getEnv(key: string, fallback?: string): string {
    const runtime = window.__ENV__?.[key];
    return runtime && runtime !== "" ? runtime : (fallback ?? "");
}
