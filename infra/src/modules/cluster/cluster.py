import json
import os
import shlex
import subprocess
import sys
import tempfile
import time
from collections.abc import Iterator
from concurrent.futures import ThreadPoolExecutor
from functools import partial
from pathlib import Path

import yaml


def successful_outputs(command: list[str], attempts: int) -> Iterator[str]:
    for _ in range(attempts):
        try:
            yield subprocess.check_output(
                command,
                stderr=subprocess.PIPE,
                text=True,
                timeout=10,
            )
        except (subprocess.CalledProcessError, subprocess.TimeoutExpired):
            pass

        time.sleep(2)


def ssh(host: str, command: str) -> list[str]:
    return [
        "ssh",
        "-o", "BatchMode=yes",
        "-o", "ConnectTimeout=5",
        "-o", "StrictHostKeyChecking=no",
        "-o", "UserKnownHostsFile=/dev/null",
        f"root@{host}",
        command,
    ]


def read_file(host: str, path: str) -> str:
    responses = successful_outputs(ssh(host, f"cat -- {shlex.quote(path)}"), attempts=30)
    value = next(filter(str.strip, responses), None)

    if value is None:
        raise TimeoutError(f"Could not read {path} from {host}")

    return value


def join(token: str, host: str) -> None:
    for command, data in [
        ("install -d -m 0700 /var/lib/rancher/k3s/server", None),
        ("install -m 0600 /dev/stdin /var/lib/rancher/k3s/server/token", token),
        ("systemctl --no-block start k3s", None),
    ]:
        subprocess.run(
            ssh(host, command),
            input=data,
            text=True,
            capture_output=True,
            timeout=60,
            check=True,
        )


def with_api_server(config: dict, vip: str) -> dict:
    return config | {
        "clusters": [
            entry | {"cluster": entry["cluster"] | {"server": f"https://{vip}:6443"}}
            for entry in config["clusters"]
        ],
    }


def export_config(config: dict, destination: str) -> None:
    path = Path(destination)

    with tempfile.TemporaryDirectory(dir=path.parent) as directory:
        temporary = Path(directory) / "kubeconfig"
        temporary.write_text(yaml.safe_dump(config))
        temporary.chmod(0o600)
        temporary.replace(path)


def enroll(config: dict) -> None:
    initializer = config["hosts"][config["initializer"]]["ip"]
    token = read_file(initializer, "/var/lib/rancher/k3s/server/token").strip()

    joiners = (
        host["ip"]
        for name, host in config["hosts"].items()
        if name != config["initializer"]
    )

    with ThreadPoolExecutor() as pool:
        tuple(pool.map(partial(join, token), joiners))

    kubeconfig = yaml.safe_load(read_file(initializer, "/etc/rancher/k3s/k3s.yaml"))
    export_config(with_api_server(kubeconfig, config["vip"]), config["kubeconfig"])

    ready = successful_outputs(
        ["kubectl", "--kubeconfig", config["kubeconfig"], "get", "--raw=/readyz"],
        attempts=30,
    )
    if next(ready, None) is None:
        raise TimeoutError("Cluster API did not become ready")


if __name__ == "__main__":
    try:
        if sys.argv[1] == "identities":
            query = json.load(sys.stdin)
            json.dump(
                {
                    name: read_file(host["ip"], "/etc/machine-id").strip()
                    for name, host in json.loads(query["hosts"]).items()
                },
                sys.stdout,
            )
        elif sys.argv[1] == "enroll":
            enroll(json.loads(os.environ["CLUSTER_CONFIG"]))
        else:
            sys.exit("Expected identities or enroll")
    except (OSError, subprocess.SubprocessError, yaml.YAMLError, ValueError):
        sys.exit("Cluster enrollment failed; check SSH access and K3s")
