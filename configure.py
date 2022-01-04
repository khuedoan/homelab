#!/usr/bin/python

"""
Basic configure script for new users
"""

import os
import platform
import sys

if sys.version_info < (3, 10, 0):
    raise Exception("Must be using Python >= 3.10.0")

if platform.system() != 'Linux':
    raise Exception("Only Linux is supported, please us a Linux VM or switch operating system")

# TODO
# - check if docker installed
# - check if make installed
# - confirm text editor $EDITOR
# - change domain
# - change seed repo
# - change gitops repo
# - add Gitea remote?
# - change hardware info
os.system(f"{os.getenv('EDITOR')} 'metal/inventories/prod.yml'")
# - change Terraform workspace (and make it optional?)
