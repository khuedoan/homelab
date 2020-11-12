#!/bin/python

import os

user = "root"

nodes = [
    "node0.homelab.local",
    "node1.homelab.local",
    "node2.homelab.local",
    "node3.homelab.local"
]


for node in nodes:
    print(f"ssh {user}@{node} poweroff")
    os.system(f"ssh {user}@{node} poweroff")
