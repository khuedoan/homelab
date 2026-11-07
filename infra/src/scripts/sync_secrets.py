import json
import subprocess
import sys
from pathlib import Path

from openbao import openbao, write_secret


def outputs(unit):
    output = subprocess.check_output(
        [
            "terragrunt", "run", "--non-interactive", "--no-color",
            "--tf-forward-stdout", "--", "output", "-json",
        ],
        cwd=unit,
    )
    return json.loads(output).get("secrets", {}).get("value", {})

target = Path(sys.argv[1]).resolve()
# TODO revisit if we can get the entire stack output at once
records = {
    path: values
    for config in sorted(target.glob("*/terragrunt.hcl"))
    for path, values in outputs(config.parent).items()
}

with openbao(target / "kubeconfig.yaml") as client:
    for path, values in records.items():
        write_secret(client, path, values)
