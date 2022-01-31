#!/usr/bin/python

# WIP
# TODO clean this up

"""
Basic configure script for new users
"""

import os
import platform
import sys

editor = os.getenv('EDITOR')
seed_repo = "github.com/khuedoan/homelab"
domain = "khuedoan.com"
terraform_workspace = "khuedoan"

if sys.version_info < (3, 10, 0):
    raise Exception("Must be using Python >= 3.10.0")

# confirm text editor
editor = str(input(f"Text editor ({editor}): ") or editor)

# Replace seed repo
seed_repo = str(input(f"Enter seed repo ({seed_repo}): ") or seed_repo)
os.system(f"./scripts/replace-gitops-repo {seed_repo}")

# Replace domain
domain = str(input(f"Enter your domain ({domain}): ") or domain)
os.system(f"./scripts/replace-domain {domain}")

# Change hardware info
os.system(f"{editor} 'metal/inventories/prod.yml'") # TODO use var for inventory

def replace_terraform_workspace(current, new):
    filename = 'external/versions.tf'
    # Read in the file
    with open(filename, 'r') as file:
        filedata = file.read()

    # Replace the target string
    filedata = filedata.replace(current, new)

    # Write the file out again
    with open(filename, 'w') as file:
        file.write(filedata)


new_terraform_workspace = str(
    input(f"Enter your Terraform Workspace ({terraform_workspace}): ")
    or terraform_workspace
)

replace_terraform_workspace(terraform_workspace, new_terraform_workspace)

# TODO switch to git lib
os.system("git diff")
