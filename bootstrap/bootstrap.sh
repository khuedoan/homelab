#!/bin/sh

set -e

# DHCP server

export PUBLIC_IP="192.168.1.19"

talosctl cluster create \
  -p 69:69/udp,8081:8081/tcp \
  --workers 0 \
  --endpoint $PUBLIC_IP

kubectl taint node talos-default-master-1 node-role.kubernetes.io/master:NoSchedule-

SIDERO_CONTROLLER_MANAGER_HOST_NETWORK=true \
SIDERO_CONTROLLER_MANAGER_API_ENDPOINT=$PUBLIC_IP \
SIDERO_CONTROLLER_MANAGER_AUTO_ACCEPT_SERVERS=true \
clusterctl init -b talos -c talos -i sidero

kubectl get servers -o wide

wol '00:23:24:d1:f3:f0'

kubectl edit server.metal.sidero.dev/d7a60000-720c-7a4c-81d5-1db10300eae2
# spec:
#   configPatches:
#   - op: replace
#     path: /machine/install/disk
#     value: /dev/sda

export CONTROL_PLANE_SERVERCLASS=any
export WORKER_SERVERCLASS=any
export KUBERNETES_VERSION=v1.20.1
export CONTROL_PLANE_PORT=6443
export CONTROL_PLANE_ENDPOINT=192.168.1.24
clusterctl config cluster management-plane -i sidero > management-plane.yaml
kubectl delete -f management-plane.yaml
