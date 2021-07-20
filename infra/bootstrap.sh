#!/bin/sh

set -euo pipefail

# Create ephemeral cluster
kind create cluster \
    --wait 60s \
    --kubeconfig ephemeral-cluster/kind-kubeconfig.yaml \
    --config ephemeral-cluster/kind.yaml

# Install Sidero
clusterctl init \
    --bootstrap talos \
    --control-plane talos \
    --infrastructure sidero \
    --kubeconfig ephemeral-cluster/kind-kubeconfig.yaml \
    --config clusterctl.yaml
kubectl wait pods \
    --kubeconfig ephemeral-cluster/kind-kubeconfig.yaml \
    --all \
    --all-namespaces \
    --timeout 300s \
    --for=condition=Ready

# Apply server classes
kubectl apply \
    --kubeconfig ephemeral-cluster/kind-kubeconfig.yaml \
    --filename serverclasses/

# Waker servers up
wol '00:23:24:d1:f3:f0'
# wol '00:23:24:d1:f4:d6'
# wol '00:23:24:d1:f5:69'
# wol '00:23:24:e7:04:60'

# Create cluster
export CONTROL_PLANE_ENDPOINT=$(kubectl get server \
    --kubeconfig ephemeral-cluster/kind-kubeconfig.yaml \
    --output jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
clusterctl config cluster \
    --infrastructure sidero \
    --kubeconfig ephemeral-cluster/kind-kubeconfig.yaml \
    --config clusterctl.yaml \
    homelab > cluster/homelab.yaml
kubectl apply \
    --kubeconfig ephemeral-cluster/kind-kubeconfig.yaml \
    --filename cluster/homelab.yaml

# Get kube config
clusterctl get kubeconfig \
    --kubeconfig ephemeral-cluster/kind-kubeconfig.yaml \
    homelab > kubeconfig.yaml

# Cleanup ephemeral cluster
# kind delete cluster --name bootstrap-cluster
# rm ephemeral-cluster/kind-kubeconfig.yaml
