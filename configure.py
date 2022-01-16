#!/usr/bin/python

# WIP
# clean this up

"""
Basic configure script for new users
"""

import os
import platform
import sys

editor = os.getenv('EDITOR')
seed_repo = "github.com/khuedoan/homelab"
domain = "khuedoan.com"

if sys.version_info < (3, 10, 0):
    raise Exception("Must be using Python >= 3.10.0")

if platform.system() != 'Linux':
    raise Exception("Only Linux is supported, please us a Linux VM or switch operating system")

# confirm text editor
editor = str(input(f"Text editor ({editor}): ") or editor)

# Replace seed repo
seed_repo = str(input(f"Enter seed repo ({seed_repo}): ") or seed_repo)
os.system(f"./scripts/replace-gitops-repo {seed_repo}")

# Replace domain
domain = str(input(f"Enter your domain ({domain}): ") or domain)
os.system(f"./scripts/replace-domain {domain}")

# change hardware info
os.system(f"{editor} 'metal/inventories/prod.yml'") # TODO use var for inventory

# TODO change Terraform workspace

# TODO switch to git lib
os.system("git diff")
