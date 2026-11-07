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
from sso import USER_GROUPS, admin_password, kanidm


def parse_groups(groups, allowed):
    memberships = {"users"} | {
        group.strip() for group in groups.split(",") if group.strip()
    }

    missing = memberships - allowed
    if missing:
        raise ValueError(f"Unknown groups: {', '.join(sorted(missing))}")

    return memberships


async def set_groups(client, username, memberships):
    attributes = GroupAttrApi(client.openapi_client)

    for group in sorted(USER_GROUPS):
        if group in memberships:
            await attributes.group_id_attr_post(group, "member", [username])
        else:
            await attributes.group_id_attr_delete(group, "member", [username])


async def main():
    action, target, username, *arguments = sys.argv[1:]

    if action not in ("create", "reset", "groups"):
        raise ValueError(f"Unknown user action: {action}")

    if action == "create":
        display_name, email, *arguments = arguments
        arguments = arguments or [""]

    if action != "reset":
        (groups,) = arguments
        memberships = parse_groups(groups, USER_GROUPS)

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
            people = PersonApi(client.openapi_client)

            if action == "create":
                await people.person_post(
                    Entry(
                        attrs={
                            "name": [username],
                            "displayname": [display_name],
                            "mail": [email],
                        }
                    )
                )
            elif action == "groups":
                await people.person_id_get(username)

            if action != "reset":
                await set_groups(client, username, memberships)

            if action == "groups":
                print(f"Groups for {username}: {', '.join(sorted(memberships))}")
                return

            token = await client.person_account_credential_update_token(username)

    print(f"https://auth.{domain}/ui/reset?{urlencode({'token': token.token})}")


if __name__ == "__main__":
    asyncio.run(main())
