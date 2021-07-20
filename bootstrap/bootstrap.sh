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

sleep 30

kubectl apply -f serverclasses
kubectl apply -f servers
kubectl apply -f clusters

wol '00:23:24:d1:f3:f0'
wol '00:23:24:d1:f4:d6'
wol '00:23:24:d1:f5:69'
wol '00:23:24:e7:04:60'

kubectl get talosconfig \
  -l cluster.x-k8s.io/cluster-name=management-plane \
  -o yaml -o jsonpath='{.items[0].status.talosConfig}' > management-plane-talosconfig.yaml

clusterctl init \
  --kubeconfig-context=admin@management-plane
  -i sidero -b talos -c talos

clusterctl move \
  --kubeconfig-context=admin@talos-default \
  --to-kubeconfig=$HOME/.kube/config \
  --to-kubeconfig-context=admin@management-plane

kubectl taint node talos-192-168-1-24 node-role.kubernetes.io/master:NoSchedule-

kubectl patch deploy -n sidero-system sidero-controller-manager --type='json' -p='[{"op": "add", "path": "/spec/template/spec/hostNetwork", "value": true}]'
