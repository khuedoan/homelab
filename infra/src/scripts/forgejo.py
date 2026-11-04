import base64
import os
import subprocess
from contextlib import contextmanager

import hvac
import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry

from openbao import write_secret


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


def configure_woodpecker(session, domain, bao):
    endpoint = f"https://git.{domain}/api/v1/user/applications/oauth2"
    app = next(
        (app for app in pages(session, endpoint) if app["name"] == "woodpecker"), None
    )
    callback = f"https://ci.{domain}/authorize"

    try:
        saved = bao.secrets.kv.v2.read_secret_version(
            "woodpecker/forgejo",
            raise_on_deleted_version=True,
        )["data"]["data"]
    except hvac.exceptions.InvalidPath:
        saved = {}

    if (
        app is not None
        and app["confidential_client"]
        and app["redirect_uris"] == [callback]
        and app["client_id"] == saved.get("client_id")
        and saved.get("client_secret")
        and saved.get("redirect_uri") == callback
    ):
        return

    desired = {
        "name": "woodpecker",
        "confidential_client": True,
        "redirect_uris": [callback],
    }
    app = (
        request(session, "POST", endpoint, json=desired)
        if app is None
        else request(
            session,
            "PATCH",
            f"{endpoint}/{app['id']}",
            json=desired,
        )
    )
    write_secret(
        bao,
        "woodpecker/forgejo",
        {
            "client_id": app["client_id"],
            "client_secret": app["client_secret"],
            "redirect_uri": callback,
        },
    )


def push(session, domain, repository):
    authorization = base64.b64encode(":".join(session.auth).encode()).decode()
    subprocess.run(
        [
            "git",
            "-c",
            "credential.helper=",
            "-c",
            "http.followRedirects=false",
            "--config-env=http.extraHeader=FORGEJO_AUTHORIZATION",
            "push",
            "-o",
            "repo.private=false",
            f"https://git.{domain}/ops/homelab.git",
            "HEAD:refs/heads/master",
        ],
        cwd=repository,
        env=os.environ
        | {
            "GIT_TERMINAL_PROMPT": "0",
            "GIT_ASKPASS": "",
            "FORGEJO_AUTHORIZATION": f"Authorization: Basic {authorization}",
        },
        check=True,
    )


def setup(domain, bao, repository):
    with forgejo(bao) as session:
        configure_organization(session, domain)
        team = configure_team(session, domain)
        configure_woodpecker(session, domain, bao)
        push(session, domain, repository)

        request(
            session,
            "PUT",
            f"https://git.{domain}/api/v1/teams/{team['id']}/repos/ops/homelab",
        )
