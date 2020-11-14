#!/bin/python

import os
import time
import string

user = "root"

nodes = [
    {
        'name': 'Node 0',
        'mac': '00:23:24:d1:f3:f0',
        'ip': '192.168.1.110',
        'hostname': 'node0.homelab.khuedoan.com'
    },
    {
        'name': 'Node 1',
        'mac': '00:23:24:d1:f4:d6',
        'ip': '192.168.1.111',
        'hostname': 'node1.homelab.khuedoan.com'
    },
    {
        'name': 'Node 2',
        'mac': '00:23:24:e7:04:60',
        'ip': '192.168.1.112',
        'hostname': 'node2.homelab.khuedoan.com'
    },
    {
        'name': 'Node 3',
        'mac': '00:23:24:d1:f5:69',
        'ip': '192.168.1.113',
        'hostname': 'node3.homelab.khuedoan.com'
    }
]

def is_alive(node):
    return os.system(f"ping -q -c 1 {node['ip']}") == 0

def is_ready(node):
    return os.system(f"ssh -q -o StrictHostKeyChecking=no {user}@{node['ip']} exit") == 0

def generate_network_config(node):
    with open('./kickstart/config/00-00-00-00-00-00.ks.template', 'r') as template_file:
        template = string.Template(template_file.read())
        config = template.substitute({
            'NETWORK_DEVICE': 'eno1',
            'IP': node['ip'],
            'GATEWAY': '192.168.1.1',
            'NETMASK': '255.255.255.0',
            'HOSTNAME': node['hostname'],
            'DNS1': '8.8.8.8',
            'DNS2': '8.8.8.8'
        })

    with open(f"./kickstart/config/{node['mac'].replace(':', '-')}.ks", 'w') as config_file:
        config_file.write(config)

def poweroff(node):
    if is_alive(node):
        print(f"Poweroff {node['name']}")
        os.system(f"ssh -q -o StrictHostKeyChecking=no {user}@{node['ip']} poweroff")
    else:
        print(f"Node {node['name']} is already dead!")

def wake(node):
    os.system(f"wol {node['mac']}")

def forget(node):
    os.system(f"ssh-keygen -R {node['ip']}")
    os.system(f"ssh-keygen -R {node['hostname']}")

if __name__ == "__main__":
    os.chdir(f"./infra/pxe-server")

    for node in nodes:
        generate_network_config(node)

    os.system(f"docker-compose up -d --build")

    for node in nodes:
        poweroff(node)
        forget(node)

    time.sleep(10)

    for node in nodes:
        wake(node)

    while not all(is_ready(node) for node in nodes):
        os.system(f"docker-compose logs -t --tail={len(nodes)} nginx")
        print("Waiting for all servers to start up...")
        time.sleep(10)

    os.system(f"docker-compose down")
