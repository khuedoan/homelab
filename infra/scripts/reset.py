#!/bin/python

import os
import time

user = "root"

nodes = [
    {
        'name': 'Node 0',
        'mac': '00:23:24:d1:f3:f0',
        'ip': '192.168.1.21',
        'hostname': 'node0.homelab.local'
    },
    {
        'name': 'Node 1',
        'mac': '00:23:24:d1:f4:d6',
        'ip': '192.168.1.22',
        'hostname': 'node1.homelab.local'
    },
    {
        'name': 'Node 2',
        'mac': '00:23:24:e7:04:60',
        'ip': '192.168.1.17',
        'hostname': 'node2.homelab.local'
    },
    {
        'name': 'Node 3',
        'mac': '00:23:24:d1:f5:69',
        'ip': '192.168.1.23',
        'hostname': 'node3.homelab.local'
    }
]

def is_alive(node):
    return os.system(f"ping -q -c 1 {node['ip']}") == 0

def is_ready(node):
    return os.system(f"ssh -o StrictHostKeyChecking=no {user}@{node['ip']} exit") == 0

def poweroff(node):
    if is_alive(node):
        print(f"Poweroff {node['name']}")
        os.system(f"ssh -o StrictHostKeyChecking=no {user}@{node['ip']} poweroff")
    else:
        print(f"Node {node['name']} is already dead!")

def wake(node):
    print(f"Waking up {node['name']}")
    os.system(f"wol {node['mac']}")

def forget(node):
    os.system(f"ssh-keygen -R {node['ip']}")
    os.system(f"ssh-keygen -R {node['hostname']}")

if __name__ == "__main__":
    os.chdir(f"./infra/pxe-server")
    os.system(f"docker-compose up -d --build")

    for node in nodes:
        poweroff(node)
        forget(node)

    time.sleep(10)

    for node in nodes:
        wake(node)

    while not all(is_ready(node) for node in nodes):
        print("Waiting for all servers to start up...")
        time.sleep(10)

    os.system(f"docker-compose down")
