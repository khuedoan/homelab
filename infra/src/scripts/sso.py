import asyncio
import json
import subprocess
from contextlib import asynccontextmanager
from urllib.parse import urlencode

import hvac
from aiohttp import ClientConnectionError, ClientSession, ClientTimeout, FormData
from kanidm import KanidmClient
from kanidm_openapi_client.api.group_api import GroupApi
from kanidm_openapi_client.api.group_attr_api import GroupAttrApi
from kanidm_openapi_client.api.oauth2_api import Oauth2Api
from kanidm_openapi_client.api.person_api import PersonApi
from kanidm_openapi_client.models.entry import Entry
from kanidm_openapi_client.exceptions import ServiceException

from openbao import write_secret


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
        async with asyncio.timeout(900):
            while True:
                try:
                    if await client.status():
                        break
                except (ClientConnectionError, ServiceException):
                    pass

                await asyncio.sleep(2)

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


async def create_user(
    domain, kubeconfig, bao, username, display_name, email, groups=()
):
    async with kanidm(domain, admin_password(kubeconfig, bao)) as client:
        memberships = dict.fromkeys(["users", *groups])
        existing = {group.name for group in await client.group_list()}
        missing = memberships.keys() - existing
        if missing:
            raise ValueError(f"Unknown groups: {', '.join(sorted(missing))}")

        await PersonApi(client.openapi_client).person_post(
            Entry(
                attrs={
                    "name": [username],
                    "displayname": [display_name],
                    "mail": [email],
                }
            )
        )
        group_attrs = GroupAttrApi(client.openapi_client)
        for group in memberships:
            await group_attrs.group_id_attr_post(group, "member", [username])
        token = await client.person_account_credential_update_token(username)

    return f"https://auth.{domain}/ui/reset?{urlencode({'token': token.token})}"


def client_attributes(client):
    attributes = {
        "displayname": [client["title"]],
        "oauth2_rs_origin_landing": [client["landing"]],
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


async def configure_icons(client, domain):
    async with ClientSession(timeout=ClientTimeout(total=30)) as session:
        for definition in oauth_clients(domain):
            name = definition["name"]
            async with session.get(definition["icon_url"]) as response:
                response.raise_for_status()
                image = await response.read()
                content_type = response.content_type

            form = FormData()
            form.add_field("image", image, filename=name, content_type=content_type)
            async with session.post(
                f"https://auth.{domain}/v1/oauth2/{name}/_image",
                headers={"Authorization": f"Bearer {client.config.auth_token}"},
                data=form,
            ) as response:
                response.raise_for_status()


async def configure_access(client, domain):
    oauth = Oauth2Api(client.openapi_client)
    groups = {group.name: group for group in await client.group_list()}
    existing = {entry.name: entry for entry in await client.oauth2_rs_list()}

    for definition in oauth_clients(domain):
        name = definition["name"]
        group = groups[definition.get("group", "users")]

        previous = {mapping.group for mapping in existing[name].oauth2_rs_scope_map}

        await oauth.oauth2_id_scopemap_post(
            name,
            group.uuid,
            ["openid", "profile", "email", "groups"],
        )

        for membership in sorted(previous - {group.spn}):
            await oauth.oauth2_id_scopemap_delete(name, membership)

        for claim, membership, value in definition.get("claims", []):
            await oauth.oauth2_id_claimmap_post(name, claim, membership, [value])


async def sync_client_secrets(client, domain, bao):
    records = {
        f"{definition['name']}/sso": {
            "client_id": definition["name"],
            "client_secret": await client.oauth2_rs_secret_get(definition["name"]),
        }
        for definition in oauth_clients(domain)
        if not definition.get("public")
    }

    for path, values in records.items():
        write_secret(bao, path, values)


def configure_openbao(domain, bao):
    secret = bao.secrets.kv.v2.read_secret_version(
        "openbao/sso",
        raise_on_deleted_version=True,
    )["data"]["data"]

    if "oidc/" not in bao.sys.list_auth_methods()["data"]:
        bao.sys.enable_auth_method("oidc", path="oidc", description="SSO")

    bao.auth.oidc.configure(
        oidc_discovery_url=f"https://auth.{domain}/oauth2/openid/openbao",
        oidc_client_id=secret["client_id"],
        oidc_client_secret=secret["client_secret"],
        default_role="sso",
        jwt_supported_algs=["ES256"],
    )

    bao.auth.oidc.create_role(
        name="sso",
        role_type="oidc",
        user_claim="sub",
        bound_audiences=[secret["client_id"]],
        allowed_redirect_uris=[
            f"https://openbao.{domain}/ui/vault/auth/oidc/oidc/callback"
        ],
        oidc_scopes=["profile", "email", "groups"],
        token_policies=["default"],
        token_ttl="1h",
    )


async def setup(domain, kubeconfig, bao):
    async with kanidm(domain, admin_password(kubeconfig, bao)) as client:
        await configure_groups(client)
        await configure_clients(client, domain)
        await configure_icons(client, domain)
        await configure_access(client, domain)
        await sync_client_secrets(client, domain, bao)

    configure_openbao(domain, bao)


def oauth_clients(domain):
    def url(host, path):
        return f"https://{host}.{domain}{path}"

    return [
        {
            "name": "hister",
            "title": "Hister",
            "landing": url("hister", "/"),
            "icon_url": "https://raw.githubusercontent.com/asciimoo/hister/v0.21.0/assets/logo.svg",
            "redirects": [url("hister", "/api/oauth/callback?provider=oidc")],
        },
        {
            "name": "grafana",
            "title": "Grafana",
            "landing": url("grafana", "/"),
            "icon_url": "https://grafana.com/static/img/menu/grafana2.svg",
            "redirects": [url("grafana", "/login/generic_oauth")],
        },
        {
            "name": "forgejo",
            "title": "Forgejo",
            "landing": url("git", "/"),
            "icon_url": "https://forgejo.org/favicon.svg",
            "redirects": [url("git", "/user/oauth2/SSO/callback")],
        },
        {
            "name": "argocd",
            "title": "Argo CD",
            "landing": url("argocd", "/"),
            "icon_url": "https://avatars.githubusercontent.com/u/30269780",
            "redirects": [url("argocd", "/auth/callback")],
        },
        {
            "name": "actualbudget",
            "title": "Actual Budget",
            "landing": url("budget", "/"),
            "icon_url": "https://github.com/actualbudget.png?size=128",
            # TODO remove RS256 once Actual Budget supports ES256: https://github.com/actualbudget/actual/issues/6524
            "rs256": True,
            "redirects": [url("budget", "/openid/callback")],
        },
        {
            "name": "matrix",
            "title": "Matrix",
            "landing": url("chat", "/"),
            "icon_url": "https://raw.githubusercontent.com/krille-chan/fluffychat/main/assets/logo/mini/logo_mini.png",
            "redirects": [
                url("matrix", "/_matrix/client/unstable/login/sso/callback/matrix")
            ],
        },
        {
            "name": "immich",
            "title": "Immich",
            "landing": url("photos", "/"),
            "icon_url": "https://github.com/immich-app.png?size=128",
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
            "landing": url("kiali", "/"),
            "icon_url": "https://raw.githubusercontent.com/kiali/kiali.io/staging/static/apple-touch-icon-180x180.png",
            "group": "admins",
            "rs256": True,
            "redirects": [url("kiali", "/")],
        },
        {
            "name": "wireguard",
            "title": "WireGuard Portal",
            "landing": url("vpn", "/"),
            "icon_url": "https://www.wireguard.com/img/icons/favicon-192.png",
            "redirects": [url("vpn", "/api/v0/auth/login/sso/callback")],
        },
        {
            "name": "opencloud",
            "title": "OpenCloud",
            "landing": url("drive", "/"),
            "icon_url": "https://github.com/opencloud-eu.png?size=128",
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
            "landing": url("registry", "/"),
            "icon_url": "https://zotregistry.dev/v2.0.2/assets/images/logo.svg",
            "public": True,
            "redirects": [url("registry", "/zot/auth/callback/oidc")],
        },
        {
            "name": "openbao",
            "title": "OpenBao",
            "landing": url("openbao", "/"),
            "icon_url": "https://openbao.org/img/favicon.svg",
            "group": "admins",
            "redirects": [url("openbao", "/ui/vault/auth/oidc/oidc/callback")],
        },
    ]
