import asyncio
import json
import subprocess
import sys
import tomllib
from pathlib import Path
from urllib.parse import urlencode

from kanidm_openapi_client.api.group_attr_api import GroupAttrApi
from kanidm_openapi_client.api.person_api import PersonApi
from kanidm_openapi_client.models.entry import Entry

from openbao import openbao
from sso import admin_password, kanidm


async def create_user(client, username, display_name, email, groups=""):
    groups = [group.strip() for group in groups.split(",") if group.strip()]
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


async def main():
    action, target, username, *arguments = sys.argv[1:]
    if action not in ("create", "reset"):
        raise ValueError(f"Unknown user action: {action}")

    kubeconfig = Path(target).resolve() / "kubeconfig.yaml"
    output = subprocess.check_output(
        [
            "kubectl",
            "--kubeconfig",
            str(kubeconfig),
            "--namespace",
            "kanidm",
            "get",
            "configmap",
            "kanidm-config",
            "-o",
            "json",
        ]
    )
    config = tomllib.loads(json.loads(output)["data"]["server.toml"])
    domain = config["domain"].removeprefix("auth.")

    with openbao(kubeconfig) as bao:
        async with kanidm(domain, admin_password(kubeconfig, bao)) as client:
            if action == "create":
                await create_user(client, username, *arguments)

            token = await client.person_account_credential_update_token(username)

    print(f"https://auth.{domain}/ui/reset?{urlencode({'token': token.token})}")


if __name__ == "__main__":
    asyncio.run(main())
