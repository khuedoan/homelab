#!/bin/python

import os
import time

user = "root"

nodes = [
    {
        'mac': '00:23:24:d1:f3:f0',
        'ip': '192.168.1.21',
        'hostname': 'node0.homelab.local'
    },
    {
        'mac': '00:23:24:d1:f4:d6',
        'ip': '192.168.1.22',
        'hostname': 'node1.homelab.local'
    },
    {
        'mac': '00:23:24:e7:04:60',
        'ip': '192.168.1.17',
        'hostname': 'node2.homelab.local'
    },
    {
        'mac': '00:23:24:d1:f5:69',
        'ip': '192.168.1.23',
        'hostname': 'node3.homelab.local'
    }
]

def poweroff(nodes):
    for node in nodes:
        print(f"Poweroff {node['hostname']}")
        os.system(f"ssh {user}@{node['ip']} poweroff")

def wake(nodes):
    for node in nodes:
        print(f"Waking up {node['hostname']}")
        os.system(f"wol {node['mac']}")

if __name__ == "__main__":
    os.chdir(f"./infra/pxe-server")
    os.system(f"docker-compose up -d --build")

    poweroff(nodes)
    time.sleep(10)
    wake(nodes)
