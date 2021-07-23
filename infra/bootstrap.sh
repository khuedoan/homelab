#!/bin/sh

set -euo pipefail

# Create ephemeral cluster
kind create cluster \
    --wait 60s \
    --kubeconfig ephemeral-cluster/kind-kubeconfig.yaml \
    --config ephemeral-cluster/kind.yaml

export KUBECONFIG="$PWD/ephemeral-cluster/kind-kubeconfig.yaml"

# Install Sidero
clusterctl init \
    --bootstrap talos \
    --control-plane talos \
    --infrastructure sidero \
    --config clusterctl.yaml

# Create DHCP server
# kubectl create configmap dhcp-server \
#     --from-file dhcp-server/dhcpd.conf
# kubectl apply \
#     --filename dhcp-server/deployment.yaml
docker run --detach \
    --name bootstrap-dhcp-server \
    --network=host \
    --volume $PWD/dhcp-server/dhcpd.conf:/data/dhcpd.conf \
    networkboot/dhcpd:1.1.0

# Wait for all pods to be ready
kubectl wait pods \
    --all \
    --all-namespaces \
    --timeout 300s \
    --for=condition=Ready

# Apply server classes
kubectl apply \
    --filename serverclasses/

# Waker servers up
wol '00:23:24:d1:f3:f0'

# Create cluster
while true; do
  kubectl get server --output jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}' && break
done

export CONTROL_PLANE_ENDPOINT=$(kubectl get server \
    --output jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}') && break

clusterctl config cluster \
    --infrastructure sidero \
    --config clusterctl.yaml \
    homelab | kubectl apply --filename -

# Get kube config
while true; do
  clusterctl get kubeconfig homelab > cluster/homelab-kubeconfig.yaml && break
done

export KUBECONFIG="$PWD/cluster/homelab-kubeconfig.yaml"

while true; do
  kubectl cluster-info && break
done

# Pivot Sidero to new cluster
clusterctl init \
    --bootstrap talos \
    --control-plane talos \
    --infrastructure sidero \
    --kubeconfig cluster/homelab-kubeconfig.yaml \
    --config clusterctl.yaml

clusterctl move \
  --kubeconfig=ephemeral-cluster/kind-kubeconfig.yaml \
  --to-kubeconfig=cluster/homelab-kubeconfig.yaml

# clusterctl config cluster \
#     --infrastructure sidero \
#     --config clusterctl.yaml \
#     --config clusterctl.yaml \
#     --worker-machine-count 3 \
#     homelab > cluster/homelab.yaml
# kubectl apply --filename cluster/homelab.yaml

# wol '00:23:24:d1:f4:d6'
# wol '00:23:24:d1:f5:69'
# wol '00:23:24:e7:04:60'

# Cleanup ephemeral cluster
kind delete cluster --name bootstrap-cluster
rm ephemeral-cluster/kind-kubeconfig.yaml
docker rm --force bootstrap-dhcp-server
