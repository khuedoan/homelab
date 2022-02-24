#!/usr/bin/python

# WIP
# TODO clean this up

"""
Basic configure script for new users
"""

import fileinput
import os
import subprocess

default_editor = os.getenv('EDITOR')
default_seed_repo = "https://github.com/khuedoan/homelab"
default_domain = "khuedoan.com"
default_timezone = "Asia/Ho_Chi_Minh"
default_terraform_workspace = "khuedoan"

editor = str(input(f"Text editor ({default_editor}): ") or default_editor)
domain = str(input(f"Enter your domain ({default_domain}): ") or default_domain)
seed_repo = str(input(f"Enter seed repo ({default_seed_repo}): ") or default_seed_repo)
timezone = str(input(f"Enter time zone ({default_timezone}): ") or default_timezone)
terraform_workspace = str(input(f"Enter your Terraform Workspace, skip if you don't want to use external resources yet ({default_terraform_workspace}): ") or default_terraform_workspace)


def find_and_replace(pattern: str, replacement: str, paths: list[str]) -> None:
    files_with_matches = subprocess.run(
        ["git", "grep", "--files-with-matches", pattern, "--"] + paths,
        capture_output=True,
        text=True
    ).stdout.splitlines()

    for file_with_maches in files_with_matches:
        with fileinput.FileInput(file_with_maches, inplace=True) as file:
            for line in file:
                print(line.replace(pattern, replacement), end='')


def main() -> None:
    find_and_replace(
        pattern=default_domain,
        replacement=domain,
        paths=[
            ".tekton",
            "apps",
            "bootstrap",
            "platform",
            "system",
            "external"
        ]
    )

    find_and_replace(
        pattern=default_seed_repo,
        replacement=seed_repo,
        paths=[
            "bootstrap",
            "platform"
        ]
    )

    find_and_replace(
        pattern=default_timezone,
        replacement=timezone,
        paths=[
            "apps",
            "system",
            "metal"
        ]
    )

    find_and_replace(
        pattern=default_terraform_workspace,
        replacement=terraform_workspace,
        paths=[
            "external/versions.tf"
        ]
    )

    subprocess.run(
        [editor, 'metal/inventories/prod.yml']
    )

    subprocess.run(
        ["git", "diff"]
    )


if __name__ == '__main__':
    main()
