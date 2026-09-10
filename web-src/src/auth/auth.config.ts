import type { UserManagerSettings } from "oidc-client-ts";
import { WebStorageStateStore } from "oidc-client-ts";
import { getEnv } from "@/env";

export const oidcConfig: UserManagerSettings = {
    /** The URL of the OIDC/OAuth2 provider */
    authority: getEnv("AUTH_ISSUER", import.meta.env.VITE_AUTH_ISSUER),

    /** Your client application's identifier as registered with the OIDC/OAuth2 */
    client_id: getEnv("AUTH_CLIENT_ID", import.meta.env.VITE_AUTH_CLIENT_ID),
    client_secret: getEnv(
        "AUTH_CLIENT_SECRET",
        import.meta.env.VITE_AUTH_CLIENT_SECRET,
    ),
    /** The redirect URI of your client application to receive a response from the OIDC/OAuth2 provider */
    redirect_uri: `${window.location.origin}/callback`,
    /** The OIDC/OAuth2 post-logout redirect URI */
    post_logout_redirect_uri: `${window.location.origin}/`,
    /** The type of response desired from the OIDC/OAuth2 provider (default: "code") */
    response_type: "code",
    /** The scope being requested from the OIDC/OAuth2 provider (default: "openid") */
    scope: "openid profile email groups",
    /** Flag to indicate if there should be an automatic attempt to renew the access token prior to its expiration. The automatic renew attempt starts 1 minute before the access token expires (default: true) */
    automaticSilentRenew: true,
    /** The number of seconds before an access token is to expire to raise the accessTokenExpiring event (default: 60) */
    accessTokenExpiringNotificationTimeInSeconds: 60,
    /** Storage object used to persist User for currently authenticated user (default: window.sessionStorage, InMemoryWebStorage iff no window). */
    userStore: new WebStorageStateStore({ store: window.sessionStorage }),
};
