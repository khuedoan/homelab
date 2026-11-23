import base64
import json
import os
import secrets
import subprocess
import sys
from contextlib import contextmanager
from pathlib import Path
from uuid import UUID

import hvac
import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry

from openbao import openbao, write_secret


@contextmanager
def forgejo(bao):
    password = bao.secrets.kv.v2.read_secret_version(
        "forgejo/admin",
        raise_on_deleted_version=True,
    )["data"]["data"]["password"]

    with requests.Session() as session:
        session.auth = ("forgejo_admin", password)
        session.mount(
            "https://",
            HTTPAdapter(
                max_retries=Retry(
                    total=60,
                    backoff_factor=1,
                    backoff_max=5,
                    status_forcelist=[502, 503, 504],
                )
            ),
        )
        yield session


def request(session, method, url, **arguments):
    response = session.request(
        method, url, timeout=30, allow_redirects=False, **arguments
    )
    if method == "GET" and response.status_code == 404:
        return None

    response.raise_for_status()
    return response.json() if response.content else None


def pages(session, url):
    while url:
        response = session.get(
            url, params={"limit": 50}, timeout=30, allow_redirects=False
        )
        response.raise_for_status()
        yield from response.json()
        url = response.links.get("next", {}).get("url")


def configure_organization(session, domain):
    endpoint = f"https://git.{domain}/api/v1/orgs"
    return request(session, "GET", f"{endpoint}/ops") or request(
        session,
        "POST",
        endpoint,
        json={"username": "ops", "description": "Operations"},
    )


def configure_team(session, domain):
    endpoint = f"https://git.{domain}/api/v1"
    team = next(
        (
            team
            for team in pages(session, f"{endpoint}/orgs/ops/teams")
            if team["name"] == "developers"
        ),
        None,
    )
    permissions = {
        "name": "developers",
        "permission": "write",
        "units_map": {"repo.code": "write", "repo.pulls": "write"},
        "includes_all_repositories": False,
        "can_create_org_repo": False,
    }

    if team is None:
        return request(session, "POST", f"{endpoint}/orgs/ops/teams", json=permissions)

    request(session, "PATCH", f"{endpoint}/teams/{team['id']}", json=permissions)
    return team


def push(session, domain, repository, target):
    url = f"https://git.{domain}/ops/homelab.git"
    authorization = base64.b64encode(":".join(session.auth).encode()).decode()
    for key, value in {
        f"remote.{target}.url": url,
        f"remote.{target}.push": "HEAD:refs/heads/master",
        f"http.{url}.extraHeader": f"Authorization: Basic {authorization}",
    }.items():
        subprocess.run(
            [
                "git",
                "config",
                "--local",
                "--replace-all",
                key,
                value,
            ],
            cwd=repository,
            check=True,
        )

    subprocess.run(
        [
            "git",
            "-c",
            "http.followRedirects=false",
            "push",
            "-o",
            "repo.private=false",
            target,
        ],
        cwd=repository,
        env=os.environ | {"GIT_TERMINAL_PROMPT": "0", "GIT_ASKPASS": ""},
        check=True,
    )


def forgejo_command(kubeconfig):
    return [
        "kubectl",
        "--kubeconfig", kubeconfig,
        "--namespace", "forgejo",
        "exec",
        "-i",
        "deployment/forgejo",
        "--",
        "forgejo",
        "--config", "/data/gitea/conf/app.ini",
    ]


def configure_runner(bao, kubeconfig):
    path = "forgejo/linux-runner"
    try:
        saved = bao.secrets.kv.v2.read_secret_version(
            path,
            raise_on_deleted_version=True,
        )["data"]["data"]
    except hvac.exceptions.InvalidPath:
        saved = {"secret": secrets.token_hex(20)}
        write_secret(bao, path, saved)

    registration = subprocess.check_output(
        forgejo_command(kubeconfig) + [
            "forgejo-cli",
            "actions",
            "register",
            "--name", "linux",
            "--labels", "linux",
            "--secret-stdin=1",
        ],
        input=saved["secret"],
        text=True,
        timeout=120,
    )
    uuid = str(UUID(registration.strip()))

    write_secret(bao, path, {"secret": saved["secret"], "uuid": uuid})


def configure_bot(session, domain):
    api = f"https://git.{domain}/api/v1"
    return request(session, "GET", f"{api}/users/bot") or request(
        session,
        "POST",
        f"{api}/admin/users",
        json={
            "username": "bot",
            "full_name": "Bot",
            "email": f"bot@{domain}",
            "password": secrets.token_urlsafe(32),
            "must_change_password": False,
            "restricted": True,
            "max_repo_creation": 0,
        },
    )


def configure_bot_integration(
    session,
    domain,
    kubeconfig,
    repository,
    owner,
    *,
    name,
    workflow_ref,
    events,
    scopes,
):
    api = f"https://git.{domain}/api/v1"
    prefix = f"u:{owner['id']}:"
    variable = f"{api}/repos/{repository}/actions/variables/{name.upper()}_AUDIENCE"
    current = request(session, "GET", variable)
    if current and current["data"].startswith(prefix):
        return

    command = forgejo_command(kubeconfig) + [
        "admin",
        "user",
        "create-authorized-integration",
        "--username", owner["login"],
        "--name", name,
        "--issuer", "urn:forgejo:authorized-integrations:actions",
        "--claim-eq", f"repository={repository}",
        "--claim-eq", f"workflow_ref={repository}/{workflow_ref}",
        "--claim-in", f"event_name={','.join(events)}",
    ]
    output = subprocess.check_output(
        command + [arg for scope in scopes for arg in ("--scope", scope)],
        input="",
        text=True,
        timeout=120,
    )
    audience = json.loads(output)["audience"]
    request(session, "PUT" if current else "POST", variable, json={"value": audience})


def setup(domain, bao, repository, target, kubeconfig):
    with forgejo(bao) as session:
        configure_organization(session, domain)
        team = configure_team(session, domain)
        push(session, domain, repository, target)
        configure_runner(bao, kubeconfig)
        bot = configure_bot(session, domain)
        request(
            session,
            "PUT",
            f"https://git.{domain}/api/v1/repos/ops/homelab/collaborators/{bot['login']}",
            json={"permission": "write"},
        )
        configure_bot_integration(
            session,
            domain,
            kubeconfig,
            "ops/homelab",
            bot,
            name="Renovate",
            workflow_ref=".forgejo/workflows/renovate.yaml@refs/heads/master",
            events=("schedule", "workflow_dispatch"),
            scopes=("write:repository", "write:issue", "read:user", "read:organization"),
        )

        request(
            session,
            "PUT",
            f"https://git.{domain}/api/v1/teams/{team['id']}/repos/ops/homelab",
        )


if __name__ == "__main__":
    target = Path(sys.argv[1]).resolve()
    output = subprocess.check_output(
        ["terragrunt", "render", "--json"],
        cwd=target / "platform",
    )
    domain = json.loads(output)["inputs"]["domain"]

    with openbao(target / "kubeconfig.yaml") as bao:
        setup(domain, bao, Path.cwd().parent, target.name, target / "kubeconfig.yaml")
