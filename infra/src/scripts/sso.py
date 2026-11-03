from contextlib import asynccontextmanager

from kanidm import KanidmClient


@asynccontextmanager
async def kanidm(domain, password):
    client = KanidmClient(uri=f"https://auth.{domain}")

    async with client.openapi_client:
        await client.authenticate_password(
            "idm_admin",
            password,
            update_internal_auth_token=True,
        )
        yield client


def oauth_clients(domain):
    def url(host, path):
        return f"https://{host}.{domain}{path}"

    return [
        {
            "name": "grafana",
            "title": "Grafana",
            "redirects": [url("grafana", "/login/generic_oauth")],
        },
        {
            "name": "forgejo",
            "title": "Forgejo",
            "redirects": [url("git", "/user/oauth2/SSO/callback")],
        },
        {
            "name": "argocd",
            "title": "Argo CD",
            "redirects": [url("argocd", "/auth/callback")],
        },
        {
            "name": "actualbudget",
            "title": "Actual Budget",
            # TODO remove RS256 once Actual Budget supports ES256: https://github.com/actualbudget/actual/issues/6524
            "rs256": True,
            "redirects": [url("budget", "/openid/callback")],
        },
        {
            "name": "matrix",
            "title": "Matrix",
            "redirects": [
                url("matrix", "/_matrix/client/unstable/login/sso/callback/matrix")
            ],
        },
        {
            "name": "immich",
            "title": "Immich",
            "redirects": [
                url("photos", path)
                for path in (
                    "/auth/login",
                    "/user-settings",
                    "/api/oauth/mobile-redirect",
                )
            ],
            "claims": [
                ("immich_role", "users", "user"),
                ("immich_role", "admins", "admin"),
            ],
        },
        {
            "name": "kiali",
            "title": "Kiali",
            "group": "admins",
            "rs256": True,
            "redirects": [url("kiali", "/")],
        },
        {
            "name": "wireguard",
            "title": "WireGuard Portal",
            "redirects": [url("vpn", "/api/v0/auth/login/sso/callback")],
        },
        {
            "name": "opencloud",
            "title": "OpenCloud",
            "public": True,
            "localhost": True,
            "redirects": [
                url("drive", path)
                for path in ("/", "/oidc-callback.html", "/oidc-silent-redirect.html")
            ]
            + ["oc://android.opencloud.eu", "oc://ios.opencloud.eu"],
        },
        {
            "name": "zot",
            "title": "Zot",
            "public": True,
            "redirects": [url("registry", "/zot/auth/callback/oidc")],
        },
        {
            "name": "openbao",
            "title": "OpenBao",
            "group": "admins",
            "redirects": [url("openbao", "/ui/vault/auth/oidc/oidc/callback")],
        },
    ]
