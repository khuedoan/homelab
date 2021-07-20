#!/bin/sh

set -e

# DHCP server

talosctl cluster create \
  -p 69:69/udp,8081:8081/tcp \
  --workers 0 \
  --endpoint '192.168.1.19'

kubectl taint node talos-default-master-1 node-role.kubernetes.io/master:NoSchedule-

clusterctl init -b talos -c talos -i sidero --config clusterctl.yaml

wol '00:23:24:d1:f3:f0'
# wol '00:23:24:d1:f4:d6'
# wol '00:23:24:d1:f5:69'
# wol '00:23:24:e7:04:60'

kubectl apply -f serverclasses

# Wait for servers to register

sleep 30

clusterctl config cluster management-plane -i sidero --config clusterctl.yaml > cluster/management-plane.yaml

kubectl apply -f cluster/

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
