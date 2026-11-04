import json
import subprocess
from contextlib import asynccontextmanager

import hvac
from kanidm import KanidmClient
from kanidm_openapi_client.api.group_api import GroupApi
from kanidm_openapi_client.api.oauth2_api import Oauth2Api
from kanidm_openapi_client.models.entry import Entry


def admin_password(kubeconfig, bao):
    path = "kanidm/idm-admin"

    try:
        return bao.secrets.kv.v2.read_secret_version(
            path,
            raise_on_deleted_version=True,
        )["data"]["data"]["password"]
    except hvac.exceptions.InvalidPath:
        pass

    kubectl = ["kubectl", "--kubeconfig", kubeconfig, "--namespace", "kanidm"]

    for condition in ("create", "condition=Ready"):
        subprocess.run(
            kubectl + ["wait", f"--for={condition}", "pod/kanidm-0", "--timeout=15m"],
            check=True,
            stdout=subprocess.DEVNULL,
        )

    output = subprocess.check_output(
        kubectl
        + [
            "exec",
            "kanidm-0",
            "--",
            "/sbin/kanidmd",
            "scripting",
            "recover-account",
            "-c",
            "/data/server.toml",
            "idm_admin",
        ]
    )
    password = json.loads(output)["output"]

    bao.secrets.kv.v2.create_or_update_secret(path, {"password": password})
    return password


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


async def configure_groups(client):
    existing = {group.name for group in await client.group_list()}
    missing = {"users", "developers", "admins"} - existing
    groups = GroupApi(client.openapi_client)

    for name in sorted(missing):
        await groups.group_post(Entry(attrs={"name": [name]}))


def client_attributes(client):
    attributes = {
        "displayname": [client["title"]],
        "oauth2_rs_origin_landing": [client["redirects"][0]],
        "oauth2_rs_origin": client["redirects"],
        "oauth2_strict_redirect_uri": ["true"],
        "oauth2_prefer_short_username": ["true"],
        "oauth2_jwt_legacy_crypto_enable": [str(client.get("rs256", False)).lower()],
    }
    authentication = (
        {
            "oauth2_allow_localhost_redirect": [
                str(client.get("localhost", False)).lower()
            ]
        }
        if client.get("public")
        else {"oauth2_allow_insecure_client_disable_pkce": ["false"]}
    )
    return attributes | authentication


async def configure_clients(client, domain):
    oauth = Oauth2Api(client.openapi_client)
    existing = {entry.attrs["name"][0]: entry for entry in await oauth.oauth2_get()}

    for definition in oauth_clients(domain):
        name = definition["name"]
        attributes = client_attributes(definition)
        kind = "public" if definition.get("public") else "basic"

        if name in existing:
            if f"oauth2_resource_server_{kind}" not in existing[name].attrs["class"]:
                raise ValueError(f"OAuth client {name} has a different client type")
            await oauth.oauth2_id_patch(name, Entry(attrs=attributes))
        else:
            create = (
                oauth.oauth2_public_post
                if definition.get("public")
                else oauth.oauth2_basic_post
            )
            await create(Entry(attrs=attributes | {"name": [name]}))


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
