import base64
import subprocess
import time
from contextlib import contextmanager

import hvac


@contextmanager
def openbao(kubeconfig):
    # Gateway TLS needs the DNS-01 credential stored in OpenBao,
    # so we need to bootstrap OpenBao with port-forwarding to avoid a chicken-and-egg problem.
    kubectl = ["kubectl", "--kubeconfig", kubeconfig, "--namespace", "openbao"]

    for condition in ("create", "jsonpath={.data.vault-root}"):
        subprocess.run(
            kubectl
            + [
                "wait",
                f"--for={condition}",
                "secret/openbao-unseal",
                "--timeout=30m",
            ],
            check=True,
            stdout=subprocess.DEVNULL,
        )

    token = subprocess.check_output(
        kubectl + ["get", "secret/openbao-unseal", "-o", "jsonpath={.data.vault-root}"]
    )

    process = subprocess.Popen(
        kubectl + ["port-forward", "--address=127.0.0.1", "service/openbao", ":8200"],
        stdout=subprocess.PIPE,
        text=True,
    )
    try:
        address = process.stdout.readline().split()[2]
        client = hvac.Client(
            url=f"http://{address}",
            token=base64.b64decode(token).decode(),
            timeout=10,
        )

        for _ in range(150):
            try:
                mounts = client.sys.list_mounted_secrets_engines()["data"]
                options = mounts.get("secret/", {}).get("options") or {}

                if options.get("version") == "2":
                    try:
                        client.secrets.kv.v2.list_secrets(path="")
                    except hvac.exceptions.InvalidPath:
                        pass  # An empty KV v2 mount is ready.
                    break
            except hvac.exceptions.VaultDown:
                pass
            except hvac.exceptions.InvalidRequest as error:
                if not str(error).startswith(
                    "Upgrading from non-versioned to versioned data"
                ):
                    raise

            time.sleep(2)
        else:
            raise TimeoutError("OpenBao KV v2 mount did not become ready")

        yield client
    finally:
        process.terminate()
        process.wait()


def write_secret(client, path, values):
    try:
        current = client.secrets.kv.v2.read_secret_version(
            path,
            raise_on_deleted_version=True,
        )["data"]["data"]
    except hvac.exceptions.InvalidPath:
        current = {}

    if current != values:
        client.secrets.kv.v2.create_or_update_secret(path, values)
